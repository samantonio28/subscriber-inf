package delivery

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	mock "github.com/samantonio28/subscriber-inf/internal/mocks"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
)

// noopLogger — заглушка логгера для тестов.
type noopLogger struct{}

func (noopLogger) WithFields(map[string]any) logger.Entry { return nil }
func (noopLogger) WithError(error) logger.Entry           { return nil }
func (noopLogger) Debug(args ...any)                      {}
func (noopLogger) Info(args ...any)                       {}
func (noopLogger) Warn(args ...any)                       {}
func (noopLogger) Error(args ...any)                      {}

func sampleSub() domain.Subscription {
	return domain.Subscription{
		SubId:       1,
		UserID:      uuid.New(),
		ServiceName: "Netflix",
		Price:       1000,
		SubType:     domain.SubTypeUsual,
		StartDate:   time.Now(),
		PlanID:      1,
	}
}

// newSubServerV2 собирает ServerV2 с usecase'ами поверх mock-репозитория подписок.
func newSubServerV2(subRepo domain.SubscriptionRepository) *ServerV2 {
	getUC, _ := usecase.NewGetSubUC(subRepo, noopLogger{})
	getSubsUC, _ := usecase.NewGetSubsUC(subRepo, noopLogger{})
	deleteUC, _ := usecase.NewDeleteSubUC(subRepo, noopLogger{})
	updateUC, _ := usecase.NewUpdateSubUC(subRepo, noopLogger{})

	subs := &SubsServer{
		GetSubUC:     *getUC,
		GetSubsUC:    *getSubsUC,
		DeleteSubUC:  *deleteUC,
		UpdateSubUC:  *updateUC,
	}
	return NewServerV2(subs, nil)
}

// ---------- GET subscription ----------

func TestGetSubscriptionById_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(1)).Return(sampleSub(), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/subscriptions/1", nil)
	rr := httptest.NewRecorder()
	srv.GetSubscriptionById(rr, req, 1)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestGetSubscriptionById_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(999)).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/subscriptions/999", nil)
	rr := httptest.NewRecorder()
	srv.GetSubscriptionById(rr, req, 999)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---------- LIST subscriptions ----------

func TestListSubscriptions_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	uid := uuid.New()
	mockSubRepo.EXPECT().UserSubs(gomock.Any(), uid).Return([]domain.Subscription{sampleSub()}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/subscriptions?user_id="+uid.String(), nil)
	rr := httptest.NewRecorder()
	srv.ListSubscriptions(rr, req, apiv2ListSubsParams(uid))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp []map[string]any
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	if len(resp) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(resp))
	}
}

// ---------- DELETE subscription ----------

func TestDeleteSubscription_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().DeleteSub(gomock.Any(), domain.SubID(1)).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/v2/subscriptions/1", nil)
	rr := httptest.NewRecorder()
	srv.DeleteSubscription(rr, req, 1)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().DeleteSub(gomock.Any(), domain.SubID(999)).Return(domain.ErrNoSubscriptionDeleted)

	req := httptest.NewRequest(http.MethodDelete, "/api/v2/subscriptions/999", nil)
	rr := httptest.NewRecorder()
	srv.DeleteSubscription(rr, req, 999)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---------- PATCH subscription ----------

func TestPatchSubscription_PartialUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	sub := sampleSub()
	// Sub вызывается 3 раза: get current + check в UpdateSub + get updated
	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(1)).Return(sub, nil).AnyTimes()
	mockSubRepo.EXPECT().UpdateSub(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	req := httptest.NewRequest(http.MethodPatch, "/api/v2/subscriptions/1", strings.NewReader(`{"price":500}`))
	rr := httptest.NewRecorder()
	srv.PatchSubscription(rr, req, 1)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestPatchSubscription_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(999)).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound)

	req := httptest.NewRequest(http.MethodPatch, "/api/v2/subscriptions/999", strings.NewReader(`{"price":500}`))
	rr := httptest.NewRecorder()
	srv.PatchSubscription(rr, req, 999)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---------- UPDATE (PUT) subscription ----------

func TestUpdateSubscription_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	sub := sampleSub()
	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(1)).Return(sub, nil).AnyTimes()
	mockSubRepo.EXPECT().UpdateSub(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	body := `{"user_id":"` + sub.UserID.String() + `","plan_id":1,"start_date":"10-2026"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v2/subscriptions/1", strings.NewReader(body))
	rr := httptest.NewRecorder()
	srv.UpdateSubscription(rr, req, 1)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSubRepo := mock.NewMockSubscriptionRepository(ctrl)
	srv := newSubServerV2(mockSubRepo)

	mockSubRepo.EXPECT().Sub(gomock.Any(), domain.SubID(999)).Return(domain.Subscription{}, domain.ErrSubscriptionNotFound)

	body := `{"user_id":"` + uuid.New().String() + `","plan_id":1,"start_date":"10-2026"}`
	req := httptest.NewRequest(http.MethodPut, "/api/v2/subscriptions/999", strings.NewReader(body))
	rr := httptest.NewRecorder()
	srv.UpdateSubscription(rr, req, 999)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

// ---------- валидация (без mock) ----------

func TestCreateUser_InvalidJSON(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/users", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	srv.CreateUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateSubscription_MissingPlanId(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	body := `{"user_id":"11111111-1111-1111-1111-111111111111","start_date":"08-2024"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/subscriptions", strings.NewReader(body))
	rr := httptest.NewRecorder()
	srv.CreateSubscription(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateSubscription_InvalidJSON(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/subscriptions", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	srv.CreateSubscription(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateService_InvalidJSON(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/services", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	srv.CreateService(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreatePromocode_InvalidJSON(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/promocodes", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	srv.CreatePromocode(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateSubscriptionPlan_InvalidJSON(t *testing.T) {
	srv := NewServerV2(&SubsServer{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v2/subscription-plans", strings.NewReader(`{not json`))
	rr := httptest.NewRecorder()
	srv.CreateSubscriptionPlan(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// apiv2ListSubsParams — хелпер для построения параметров ListSubscriptions.
func apiv2ListSubsParams(uid uuid.UUID) struct{ UserId interface{} } {
	return struct{ UserId interface{} }{UserId: uid}
}
