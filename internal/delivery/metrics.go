package delivery

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество HTTP-запросов по методу, маршруту и экземпляру.",
		},
		[]string{"method", "route", "instance"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Длительность обработки HTTP-запросов по методу, маршруту и экземпляру.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "instance"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// statusRecorder запоминает код ответа, чтобы метрики считали фактический статус.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware собирает метрики Prometheus по каждому запросу.
// Регистрируется последним (наружным), чтобы считать все запросы, включая отклонённые.
// Заодно проставляет заголовок X-Backend с именем экземпляра — по нему видно,
// какой инстанс обслужил запрос (проверка балансировки 2:1:1).
func MetricsMiddleware(instance string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Backend", instance)
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			route := routeLabel(r)
			httpRequestsTotal.WithLabelValues(r.Method, route, instance).Inc()
			httpRequestDuration.WithLabelValues(r.Method, route, instance).Observe(time.Since(start).Seconds())
		})
	}
}

// routeLabel даёт стабильную низкокардинальную метку маршрута:
// /api/v2/subscription-plans, /api/v2/services, /health, / и т.д.
// Динамические сегменты ({id}) не попадают в метку, чтобы не плодить временные ряды.
func routeLabel(r *http.Request) string {
	p := r.URL.Path
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) >= 3 && parts[0] == "api" {
		return "/" + strings.Join(parts[:3], "/")
	}
	return p
}

// MetricsHandler отдаёт метрики в формате Prometheus (promhttp).
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
