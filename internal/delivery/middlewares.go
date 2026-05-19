package delivery

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/samantonio28/subscriber-inf/internal/domain"
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware извлекает X-User-ID заголовок, находит пользователя в БД и сохраняет его ID и роль в контекст.
// Также устанавливает параметр app.current_user_id в сеансе БД для работы политик RLS.
func AuthMiddleware(userRepo domain.UserRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userIDHeader := r.Header.Get("X-User-ID")
			if userIDHeader == "" {
				// Если заголовок отсутствует, пропускаем (демо-режим)
				next.ServeHTTP(w, r)
				return
			}

			userID, err := uuid.Parse(userIDHeader)
			if err != nil {
				http.Error(w, "Invalid X-User-ID format", http.StatusBadRequest)
				return
			}

			user, err := userRepo.GetUser(r.Context(), userID)
			if err != nil {
				http.Error(w, "User not found", http.StatusUnauthorized)
				return
			}

			// Установка параметра app.current_user_id для текущего сеанса БД
			if err := userRepo.SetAppCurrentUserID(r.Context(), userID); err != nil {
				// Логируем ошибку, но не прерываем запрос (RLS может не работать, но проверки на уровне приложения остаются)
				log.Printf("failed to set app.current_user_id: %v", err)
			}

			ctx := r.Context()
			ctx = WithUserID(ctx, userID.String())
			ctx = WithUserRole(ctx, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
