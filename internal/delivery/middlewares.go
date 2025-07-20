package delivery

// import (
// 	"net/http"
// 	"time"

// 	"github.com/gorilla/mux"
// 	"github.com/sirupsen/logrus"
// )

// func AccessLogMiddleware(logger logger.Logger) mux.MiddlewareFunc {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			start := time.Now()
// 			lrw := NewResponseWriter(w)

// 			requestID := ""
// 			if ctxVal := r.Context().Value("request_id"); ctxVal != nil {
// 				requestID = ctxVal.(string)
// 			}

// 			next.ServeHTTP(lrw, r)

// 			logger.WithFields(&logrus.Fields{
// 				"method":      r.Method,
// 				"path":        r.URL.Path,
// 				"remote_addr": r.RemoteAddr,
// 				"user_agent":  r.UserAgent(),
// 				"request_id":  requestID,
// 				"status":      lrw.statusCode,
// 				"duration":    time.Since(start).String(),
// 			}).Info("request completed")
// 		})
// 	}
// }
