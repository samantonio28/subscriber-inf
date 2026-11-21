package delivery

// HTTP-хендлеры аутентификации (ЛР №5): регистрация, вход с 2FA по email-коду,
// подтверждение кода, плановая смена пароля, восстановление после блокировки.
// Каждая доменная ошибка мапится на конкретный HTTP-статус: см. authErrorResponse.

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

// authErrorResponse мапит доменные ошибки аутентификации на HTTP-статусы.
func authErrorResponse(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
	case errors.Is(err, domain.ErrInvalidCredentials):
		utils.MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "invalid email or login secret"})
	case errors.Is(err, domain.ErrAccountLocked):
		utils.MakeResponse(w, http.StatusLocked, map[string]string{
			"message": "account locked: too many failed login attempts; use /auth/recover to restore access",
		})
	case errors.Is(err, domain.ErrInvalidCode):
		utils.MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "invalid or expired confirmation code"})
	case errors.Is(err, domain.ErrUserNotFound):
		utils.MakeResponse(w, http.StatusNotFound, map[string]string{"message": "user not found"})
	case errors.Is(err, domain.ErrEmailTaken):
		utils.MakeResponse(w, http.StatusConflict, map[string]string{"message": "email already registered"})
	default:
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{"message": "internal error"})
	}
}

// decodeAuthBody декодирует JSON-тело запроса аутентификации.
func decodeAuthBody(w http.ResponseWriter, r *http.Request, req any) bool {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{"message": "invalid json"})
		return false
	}
	return true
}

// PostAuthRegister — POST /auth/register: создание пользователя для входа
// (пароль сохраняется только как bcrypt-хэш).
func (s *SubsServer) PostAuthRegister(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	balance := 0
	if req.Balance != nil {
		balance = int(*req.Balance)
	}
	dto := usecase.RegisterAuthDTO{
		Email:        string(req.Email),
		UserName:     req.UserName,
		Age:          req.Age,
		Balance:      balance,
		ReferralCode: req.ReferralCode,
		Secret:       req.LoginSecret,
	}
	userID, err := s.RegisterAuthUC.Register(r.Context(), dto)
	if err != nil {
		authErrorResponse(w, err)
		return
	}
	utils.MakeResponse(w, http.StatusCreated, api.RegisterResponse{UserId: &userID})
}

// PostAuthLogin — POST /auth/login: первый фактор (email + пароль).
// При успехе одноразовый код уходит на email (второй фактор).
func (s *SubsServer) PostAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	userID, err := s.LoginUC.Login(r.Context(), usecase.LoginDTO{
		Email:  string(req.Email),
		Secret: req.LoginSecret,
	})
	if err != nil {
		authErrorResponse(w, err)
		return
	}
	status := "code_sent"
	utils.MakeResponse(w, http.StatusOK, api.LoginResponse{UserId: &userID, Status: &status})
}

// PostAuthVerify — POST /auth/verify: второй фактор — подтверждение кода из email.
func (s *SubsServer) PostAuthVerify(w http.ResponseWriter, r *http.Request) {
	var req api.VerifyRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	if err := s.VerifyAuthCodeUC.Verify(r.Context(), usecase.VerifyCodeDTO{
		UserID: uuid.UUID(req.UserId),
		Code:   req.Code,
	}); err != nil {
		authErrorResponse(w, err)
		return
	}
	utils.MakeResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PutAuthPassword — PUT /auth/password: плановая смена пароля.
// Требуется аутентификация (X-User-ID, совпадающий с user_id в теле)
// и знание текущего пароля.
func (s *SubsServer) PutAuthPassword(w http.ResponseWriter, r *http.Request) {
	var req api.ChangePasswordRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	authUserID, ok := UserIDFromContext(r.Context())
	if !ok {
		utils.MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "authentication required (X-User-ID)"})
		return
	}
	if authUserID != req.UserId.String() {
		utils.MakeResponse(w, http.StatusUnauthorized, map[string]string{"message": "cannot change password of another user"})
		return
	}
	if err := s.ChangePasswordUC.Change(r.Context(), usecase.ChangePasswordDTO{
		UserID:        uuid.UUID(req.UserId),
		CurrentSecret: req.CurrentLoginSecret,
		NewSecret:     req.NewLoginSecret,
	}); err != nil {
		authErrorResponse(w, err)
		return
	}
	utils.MakeResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PostAuthRecover — POST /auth/recover: запрос кода восстановления
// (для заблокированного аккаунта).
func (s *SubsServer) PostAuthRecover(w http.ResponseWriter, r *http.Request) {
	var req api.RecoverRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	if err := s.RecoverAuthUC.Recover(r.Context(), string(req.Email)); err != nil {
		authErrorResponse(w, err)
		return
	}
	utils.MakeResponse(w, http.StatusOK, map[string]string{"status": "code_sent"})
}

// PostAuthRecoverConfirm — POST /auth/recover/confirm: подтверждение
// восстановления кодом из email — снимает блокировку и задаёт новый пароль.
func (s *SubsServer) PostAuthRecoverConfirm(w http.ResponseWriter, r *http.Request) {
	var req api.RecoverConfirmRequest
	if !decodeAuthBody(w, r, &req) {
		return
	}
	if err := s.RecoverAuthUC.Confirm(r.Context(), string(req.Email), req.Code, req.NewLoginSecret); err != nil {
		authErrorResponse(w, err)
		return
	}
	utils.MakeResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}
