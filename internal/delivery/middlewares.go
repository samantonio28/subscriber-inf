package delivery

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

func AccessLogMiddleware(logger logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := ""
			if ctxVal := r.Context().Value("request_id"); ctxVal != nil {
				requestID = ctxVal.(string)
			}

			next.ServeHTTP(w, r)

			logger.WithFields(map[string]any{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"request_id":  requestID,
				"duration":    time.Since(start).String(),
			}).Logger.Info("request completed")
		})
	}
}

// CORSMiddleware добавляет заголовки CORS для разрешения запросов из браузера.
func CORSMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Разрешаем любые origin для разработки
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
