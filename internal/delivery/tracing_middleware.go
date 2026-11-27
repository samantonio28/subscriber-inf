package delivery

// TracingMiddleware — HTTP-мидлварь OpenTelemetry (ЛР №6).
// На каждый запрос создаётся спан "HTTP {method} {route}" с атрибутами:
// метод, маршрут, статус ответа, длина тела. Когда трассировка выключена
// (no-op TracerProvider), создание спана ничего не стоит.

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samantonio28/subscriber-inf/internal/tracing"
	"go.opentelemetry.io/otel/attribute"
)

func TracingMiddleware() mux.MiddlewareFunc {
	tracer := tracing.Tracer()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := tracer.Start(r.Context(), "HTTP "+r.Method+" "+r.URL.Path)
			defer span.End()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			span.SetAttributes(
				attribute.Int("http.status_code", rec.status),
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
			)
		})
	}
}

// statusRecorder запоминает статус ответа для атрибута спана.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
