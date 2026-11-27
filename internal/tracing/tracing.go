// Package tracing — интеграция OpenTelemetry (ЛР №6).
//
// Экспортер OTLP/HTTP отправляет спаны в Jaeger (all-in-one принимает OTLP
// на :4318). Трассировка управляется env-флагами, что позволяет измерять
// накладные расходы:
//
//	TRACING_ENABLED=true|false        — включить/выключить трассировку
//	OTEL_EXPORTER_OTLP_ENDPOINT=...   — endpoint OTLP-экспортёра (по умолчанию http://localhost:4318)
//	OTEL_SERVICE_NAME=...             — имя сервиса в Jaeger (по умолчанию subscriber-inf)
//
// При TRACING_ENABLED=false устанавливается no-op TracerProvider: вызовы
// трассировки остаются в коде, но не выполняют работы (замер «без трассировки»).
package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// TracerName — instrumentation scope для всех спанов сервиса.
const TracerName = "subscriber-inf"

// TracingEnabled читает env-флаг трассировки (по умолчанию выключена —
// чтобы замер «без трассировки» был значением по умолчанию).
func TracingEnabled() bool {
	v := os.Getenv("TRACING_ENABLED")
	return v == "true" || v == "1" || v == "TRUE"
}

// Endpoint возвращает endpoint OTLP-экспортёра в формате host:port —
// именно его ждёт otlptracehttp.WithEndpoint (схему он добавляет сам,
// поэтому "http://" из env нужно отрезать).
func Endpoint() string {
	v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if v == "" {
		v = "http://localhost:4318"
	}
	v = strings.TrimPrefix(v, "https://")
	v = strings.TrimPrefix(v, "http://")
	v = strings.TrimSuffix(v, "/")
	return v
}

// ServiceName возвращает имя сервиса для Jaeger.
func ServiceName() string {
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		return v
	}
	return "subscriber-inf"
}

// Setup настраивает глобальный TracerProvider.
//   - enabled=true: OTLP/HTTP-экспортёр с батчингом (5 c, 512 спанов);
//   - enabled=false: no-op провайдер — нулевые накладные расходы.
//
// Возвращает shutdown-функцию (гарантированный flush при остановке сервиса).
func Setup(ctx context.Context, enabled bool) (func(context.Context) error, error) {
	if !enabled {
		noopProvider := noop.NewTracerProvider()
		otel.SetTracerProvider(noopProvider)
		otel.SetTextMapPropagator(propagation.TraceContext{})
		fmt.Println("tracing: disabled (no-op tracer provider)")
		return func(context.Context) error { return nil }, nil
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(Endpoint()),
		// Jaeger all-in-one принимает OTLP по HTTP без TLS.
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(5*time.Second),
		sdktrace.WithMaxExportBatchSize(512),
	)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName()),
		)),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	fmt.Println("tracing: enabled, OTLP exporter ->", Endpoint())

	shutdown := func(sctx context.Context) error {
		return tp.Shutdown(sctx)
	}
	return shutdown, nil
}

// Tracer — общий трасер сервиса.
func Tracer() trace.Tracer {
	return otel.GetTracerProvider().Tracer(TracerName)
}
