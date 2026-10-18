package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	apiv2 "github.com/samantonio28/subscriber-inf/internal/api/v2"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

// ServerV2 реализует v2.ServerInterface (новый REST API, openapi.yaml).
// Переиспользует usecase'ы из SubsServer + добавляет репозиторий сервисов.
type ServerV2 struct {
	subs        *SubsServer
	serviceRepo domain.ServiceRepository
}

var _ apiv2.ServerInterface = (*ServerV2)(nil)

func NewServerV2(subs *SubsServer, serviceRepo domain.ServiceRepository) *ServerV2 {
	return &ServerV2{subs: subs, serviceRepo: serviceRepo}
}

// ---------- helpers ----------

func parseMonth(s string) (time.Time, error) {
	return time.Parse("01-2006", s)
}

func formatMonth(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("01-2006")
}

func errResp(w http.ResponseWriter, status int, msg string) {
	utils.MakeResponse(w, status, apiv2.Error{Message: msg})
}

func subDTOToV2(dto usecase.SubscriptionDTO) apiv2.Subscription {
	res := apiv2.Subscription{
		SubId:       dto.SubId,
		UserId:      dto.UserId,
		PlanId:      dto.PlanID,
		Price:       dto.Price,
		SubType:     apiv2.SubscriptionType(dto.SubType),
		StartDate:   formatMonth(dto.StartDate),
		PromocodeId: dto.PromocodeID,
	}
	if dto.ServiceName != "" {
		sn := dto.ServiceName
		res.ServiceName = &sn
	}
	if !dto.EndDate.IsZero() {
		ed := formatMonth(dto.EndDate)
		res.EndDate = &ed
	}
	return res
}

func promoToV2(p domain.Promocode) apiv2.Promocode {
	res := apiv2.Promocode{
		ServiceId: p.ServiceID,
		Value:     p.Value,
		PlanId:    p.PlanID,
		Discount:  p.Discount,
		MaxUses:   p.MaxUses,
	}
	id := int(p.PromocodeID)
	res.PromocodeId = &id
	cur := p.CurUses
	res.CurUses = &cur
	dur := p.DurationDays
	res.DurationDays = &dur
	status := apiv2.PromocodeStatus(p.Status)
	res.Status = &status
	if !p.ExpiresAt.IsZero() {
		t := p.ExpiresAt
		res.ExpiresAt = &t
	}
	if !p.CreatedAt.IsZero() {
		t := p.CreatedAt
		res.CreatedAt = &t
	}
	return res
}

func planToV2(p domain.SubscriptionPlan) apiv2.SubscriptionPlan {
	res := apiv2.SubscriptionPlan{
		DurationDays: p.DurationDays,
		Name:         p.Name,
		Price:        p.Price,
		ServiceId:    p.ServiceID,
	}
	id := int(p.PlanID)
	res.PlanId = &id
	return res
}

func serviceToV2(s domain.Service) apiv2.Service {
	res := apiv2.Service{ServiceName: s.ServiceName}
	id := s.ServiceID
	res.ServiceId = &id
	uc := s.UsersCount
	res.UsersCount = &uc
	hp := s.HasPromocodes
	res.HasPromocodes = &hp
	return res
}

// ---------- Users ----------

func (s *ServerV2) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req apiv2.UserCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	dto := usecase.UserDTO{
		UserID:   uuid.New(),
		Email:    string(req.Email),
		Password: stringPtr(req.Password),
		UserName: req.UserName,
		Age:      req.Age,
		Balance:  req.Balance,
		Role:     domain.RoleUser,
	}

	userID, err := s.subs.CreateUserUC.NewUser(r.Context(), dto)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			errResp(w, http.StatusConflict, "user already exists")
			return
		}
		errResp(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	utils.MakeResponse(w, http.StatusCreated, apiv2.UserResponse{
		UserId:   userID,
		Email:    req.Email,
		UserName: req.UserName,
		Age:      req.Age,
		Balance:  req.Balance,
	})
}

func (s *ServerV2) GetUserById(w http.ResponseWriter, r *http.Request, userId openapi_types.UUID) {
	dto, err := s.subs.GetUserUC.UserById(r.Context(), userId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			errResp(w, http.StatusNotFound, "user not found")
			return
		}
		errResp(w, http.StatusInternalServerError, "failed to get user: "+err.Error())
		return
	}

	utils.MakeResponse(w, http.StatusOK, apiv2.UserResponse{
		UserId:       dto.UserID,
		Email:        openapi_types.Email(dto.Email),
		UserName:     dto.UserName,
		Age:          dto.Age,
		Balance:      dto.Balance,
		ReferralCode: dto.ReferralCode,
	})
}

// ---------- Services ----------

func (s *ServerV2) ListServices(w http.ResponseWriter, r *http.Request) {
	services, err := s.serviceRepo.ListServices(r.Context())
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to list services: "+err.Error())
		return
	}
	res := make([]apiv2.Service, 0, len(services))
	for _, svc := range services {
		res = append(res, serviceToV2(svc))
	}
	utils.MakeResponse(w, http.StatusOK, res)
}

func (s *ServerV2) CreateService(w http.ResponseWriter, r *http.Request) {
	var req apiv2.ServiceCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	svc := domain.Service{ServiceName: req.ServiceName}
	if req.UsersCount != nil {
		svc.UsersCount = *req.UsersCount
	}
	if req.HasPromocodes != nil {
		svc.HasPromocodes = *req.HasPromocodes
	}

	id, err := s.serviceRepo.StoreService(r.Context(), svc)
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to create service: "+err.Error())
		return
	}
	svc.ServiceID = id
	utils.MakeResponse(w, http.StatusCreated, serviceToV2(svc))
}

func (s *ServerV2) GetServiceById(w http.ResponseWriter, r *http.Request, serviceId int) {
	svc, err := s.serviceRepo.GetService(r.Context(), serviceId)
	if err != nil {
		errResp(w, http.StatusNotFound, "service not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, serviceToV2(svc))
}

func (s *ServerV2) UpdateService(w http.ResponseWriter, r *http.Request, serviceId int) {
	var req apiv2.ServiceCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	svc := domain.Service{ServiceID: serviceId, ServiceName: req.ServiceName}
	if req.UsersCount != nil {
		svc.UsersCount = *req.UsersCount
	}
	if req.HasPromocodes != nil {
		svc.HasPromocodes = *req.HasPromocodes
	}

	if err := s.serviceRepo.UpdateService(r.Context(), svc); err != nil {
		errResp(w, http.StatusNotFound, "service not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, serviceToV2(svc))
}

func (s *ServerV2) DeleteService(w http.ResponseWriter, r *http.Request, serviceId int) {
	if err := s.serviceRepo.DeleteService(r.Context(), serviceId); err != nil {
		errResp(w, http.StatusNotFound, "service not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Subscription plans ----------

func (s *ServerV2) ListSubscriptionPlans(w http.ResponseWriter, r *http.Request, params apiv2.ListSubscriptionPlansParams) {
	filter := usecase.SubscriptionPlanFilter{
		ServiceID:    params.ServiceId,
		NameContains: params.NameContains,
		PriceMin:     params.PriceMin,
		PriceMax:     params.PriceMax,
	}
	plans, err := s.subs.GetFilteredSubscriptionPlansUC.GetFiltered(r.Context(), filter)
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to list plans: "+err.Error())
		return
	}
	res := make([]apiv2.SubscriptionPlan, 0, len(plans))
	for _, p := range plans {
		res = append(res, apiv2.SubscriptionPlan{
			PlanId:       intPtr(p.PlanID),
			ServiceId:    p.ServiceID,
			Name:         p.Name,
			DurationDays: p.DurationDays,
			Price:        p.Price,
		})
	}
	utils.MakeResponse(w, http.StatusOK, res)
}

func (s *ServerV2) CreateSubscriptionPlan(w http.ResponseWriter, r *http.Request) {
	var req apiv2.SubscriptionPlanCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}
	id, err := s.subs.CreateSubscriptionPlanUC.Create(r.Context(), usecase.CreateSubscriptionPlanInput{
		ServiceID:    req.ServiceId,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        req.Price,
	})
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to create plan: "+err.Error())
		return
	}
	utils.MakeResponse(w, http.StatusCreated, apiv2.SubscriptionPlan{
		PlanId:       intPtr(int(id)),
		ServiceId:    req.ServiceId,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        req.Price,
	})
}

func (s *ServerV2) GetSubscriptionPlanById(w http.ResponseWriter, r *http.Request, planId int) {
	plan, err := s.subs.GetSubscriptionPlanUC.ByID(r.Context(), domain.PlanID(planId))
	if err != nil {
		errResp(w, http.StatusNotFound, "plan not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, planToV2(plan))
}

func (s *ServerV2) UpdateSubscriptionPlan(w http.ResponseWriter, r *http.Request, planId int) {
	var req apiv2.SubscriptionPlanCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}
	err := s.subs.UpdateSubscriptionPlanUC.Update(r.Context(), usecase.UpdateSubscriptionPlanInput{
		ID:           domain.PlanID(planId),
		ServiceID:    req.ServiceId,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        req.Price,
	})
	if err != nil {
		errResp(w, http.StatusNotFound, "plan not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, apiv2.SubscriptionPlan{
		PlanId:       intPtr(planId),
		ServiceId:    req.ServiceId,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        req.Price,
	})
}

func (s *ServerV2) PatchSubscriptionPlan(w http.ResponseWriter, r *http.Request, planId int) {
	var req apiv2.SubscriptionPlanPatch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	cur, err := s.subs.GetSubscriptionPlanUC.ByID(r.Context(), domain.PlanID(planId))
	if err != nil {
		errResp(w, http.StatusNotFound, "plan not found")
		return
	}

	input := usecase.UpdateSubscriptionPlanInput{
		ID:           domain.PlanID(planId),
		ServiceID:    cur.ServiceID,
		Name:         cur.Name,
		DurationDays: cur.DurationDays,
		Price:        cur.Price,
	}
	if req.Name != nil {
		input.Name = *req.Name
	}
	if req.DurationDays != nil {
		input.DurationDays = *req.DurationDays
	}
	if req.Price != nil {
		input.Price = *req.Price
	}

	if err := s.subs.UpdateSubscriptionPlanUC.Update(r.Context(), input); err != nil {
		errResp(w, http.StatusInternalServerError, "failed to update plan: "+err.Error())
		return
	}
	utils.MakeResponse(w, http.StatusOK, apiv2.SubscriptionPlan{
		PlanId:       intPtr(planId),
		ServiceId:    input.ServiceID,
		Name:         input.Name,
		DurationDays: input.DurationDays,
		Price:        input.Price,
	})
}

func (s *ServerV2) DeleteSubscriptionPlan(w http.ResponseWriter, r *http.Request, planId int) {
	if err := s.subs.DeleteSubscriptionPlanUC.Delete(r.Context(), domain.PlanID(planId)); err != nil {
		errResp(w, http.StatusNotFound, "plan not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Subscriptions ----------

func (s *ServerV2) ListSubscriptions(w http.ResponseWriter, r *http.Request, params apiv2.ListSubscriptionsParams) {
	dtos, err := s.subs.GetSubsUC.SubsByUserId(r.Context(), params.UserId)
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to list subscriptions: "+err.Error())
		return
	}
	res := make([]apiv2.Subscription, 0, len(dtos))
	for _, dto := range dtos {
		res = append(res, subDTOToV2(dto))
	}
	utils.MakeResponse(w, http.StatusOK, res)
}

func (s *ServerV2) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req apiv2.SubscriptionCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	if req.PlanId == nil {
		errResp(w, http.StatusBadRequest, "plan_id is required")
		return
	}

	plan, err := s.subs.GetSubscriptionPlanUC.ByID(r.Context(), domain.PlanID(*req.PlanId))
	if err != nil {
		errResp(w, http.StatusNotFound, "plan not found")
		return
	}
	svc, err := s.serviceRepo.GetService(r.Context(), plan.ServiceID)
	if err != nil {
		errResp(w, http.StatusNotFound, "service not found")
		return
	}

	subID, err := s.subs.PurchaseSubscriptionUC.Purchase(r.Context(), usecase.PurchaseSubscriptionDTO{
		UserID:       req.UserId,
		ServiceName:  svc.ServiceName,
		PlanID:       *req.PlanId,
		Price:        plan.Price,
		DurationDays: plan.DurationDays,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) {
			errResp(w, http.StatusPaymentRequired, "insufficient balance")
			return
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			errResp(w, http.StatusNotFound, "user not found")
			return
		}
		errResp(w, http.StatusInternalServerError, "failed to purchase subscription: "+err.Error())
		return
	}

	dto, _ := s.subs.GetSubUC.SubById(r.Context(), subID)
	utils.MakeResponse(w, http.StatusCreated, subDTOToV2(dto))
}

func (s *ServerV2) GetSubscriptionById(w http.ResponseWriter, r *http.Request, subId int) {
	dto, err := s.subs.GetSubUC.SubById(r.Context(), subId)
	if err != nil {
		errResp(w, http.StatusNotFound, "subscription not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, subDTOToV2(dto))
}

func (s *ServerV2) UpdateSubscription(w http.ResponseWriter, r *http.Request, subId int) {
	var req apiv2.SubscriptionUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	stDate, err := parseMonth(req.StartDate)
	if err != nil {
		errResp(w, http.StatusBadRequest, "invalid start_date")
		return
	}
	var enDate time.Time
	if req.EndDate != nil {
		enDate, err = parseMonth(*req.EndDate)
		if err != nil {
			errResp(w, http.StatusBadRequest, "invalid end_date")
			return
		}
	}

	subType := "usual"
	if req.SubType != nil {
		subType = string(*req.SubType)
	}
	price := 0
	if req.Price != nil {
		price = *req.Price
	}
	planID := 0
	if req.PlanId != nil {
		planID = *req.PlanId
	}

	cur, err := s.subs.GetSubUC.SubById(r.Context(), subId)
	if err != nil {
		errResp(w, http.StatusNotFound, "subscription not found")
		return
	}

	err = s.subs.UpdateSubUC.UpdateSub(r.Context(), subId, usecase.SubscriptionDTO{
		SubId:       subId,
		UserId:      req.UserId,
		ServiceName: cur.ServiceName,
		Price:       price,
		SubType:     subType,
		StartDate:   stDate,
		EndDate:     enDate,
		PlanID:      planID,
	})
	if err != nil {
		errResp(w, http.StatusNotFound, "subscription not found")
		return
	}

	dto, _ := s.subs.GetSubUC.SubById(r.Context(), subId)
	utils.MakeResponse(w, http.StatusOK, subDTOToV2(dto))
}

func (s *ServerV2) PatchSubscription(w http.ResponseWriter, r *http.Request, subId int) {
	var req apiv2.SubscriptionPatch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Применение промокода
	if req.Promocode != nil {
		_, err := s.subs.ApplyPromocodeUC.Apply(r.Context(), usecase.ApplyPromocodeInput{
			SubscriptionID: subId,
			PromocodeValue: *req.Promocode,
		})
		if err != nil {
			if errors.Is(err, domain.ErrPromocodeNotFound) || errors.Is(err, domain.ErrSubscriptionNotFound) {
				errResp(w, http.StatusNotFound, err.Error())
				return
			}
			errResp(w, http.StatusBadRequest, err.Error())
			return
		}
		dto, _ := s.subs.GetSubUC.SubById(r.Context(), subId)
		utils.MakeResponse(w, http.StatusOK, subDTOToV2(dto))
		return
	}

	// Частичное обновление полей
	cur, err := s.subs.GetSubUC.SubById(r.Context(), subId)
	if err != nil {
		errResp(w, http.StatusNotFound, "subscription not found")
		return
	}
	merged := cur
	if req.Price != nil {
		merged.Price = *req.Price
	}
	if req.SubType != nil {
		merged.SubType = string(*req.SubType)
	}
	if req.StartDate != nil {
		st, err := parseMonth(*req.StartDate)
		if err != nil {
			errResp(w, http.StatusBadRequest, "invalid start_date")
			return
		}
		merged.StartDate = st
	}
	if req.EndDate != nil {
		en, err := parseMonth(*req.EndDate)
		if err != nil {
			errResp(w, http.StatusBadRequest, "invalid end_date")
			return
		}
		merged.EndDate = en
	}

	if err := s.subs.UpdateSubUC.UpdateSub(r.Context(), subId, merged); err != nil {
		errResp(w, http.StatusInternalServerError, "failed to update subscription: "+err.Error())
		return
	}
	dto, _ := s.subs.GetSubUC.SubById(r.Context(), subId)
	utils.MakeResponse(w, http.StatusOK, subDTOToV2(dto))
}

func (s *ServerV2) DeleteSubscription(w http.ResponseWriter, r *http.Request, subId int) {
	if err := s.subs.DeleteSubUC.DeleteSub(r.Context(), subId); err != nil {
		errResp(w, http.StatusNotFound, "subscription not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Promocodes ----------

func (s *ServerV2) ListPromocodes(w http.ResponseWriter, r *http.Request, params apiv2.ListPromocodesParams) {
	filter := usecase.PromocodeFilter{
		ServiceID:    params.ServiceId,
		CodeContains: params.Code,
		DiscountMin:  params.DiscountMin,
		DiscountMax:  params.DiscountMax,
	}
	if params.Status != nil {
		status := domain.PromocodeStatus(*params.Status)
		filter.Status = &status
	}
	promos, err := s.subs.GetFilteredPromocodesUC.GetFiltered(r.Context(), filter)
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to list promocodes: "+err.Error())
		return
	}
	res := make([]apiv2.Promocode, 0, len(promos))
	for _, p := range promos {
		res = append(res, promoToV2(p))
	}
	utils.MakeResponse(w, http.StatusOK, res)
}

func (s *ServerV2) CreatePromocode(w http.ResponseWriter, r *http.Request) {
	var req apiv2.PromocodeCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	var expiresAt time.Time
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	id, err := s.subs.CreatePromocodeUC.Create(r.Context(), usecase.CreatePromocodeInput{
		ServiceID:    req.ServiceId,
		Value:        req.Value,
		PlanID:       req.PlanId,
		ExpiresAt:    expiresAt,
		Discount:     req.Discount,
		MaxUses:      req.MaxUses,
		DurationDays: intPtrVal(req.DurationDays),
	})
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to create promocode: "+err.Error())
		return
	}
	utils.MakeResponse(w, http.StatusCreated, apiv2.Promocode{
		PromocodeId: intPtr(int(id)),
		ServiceId:   req.ServiceId,
		Value:       req.Value,
		PlanId:      req.PlanId,
		Discount:    req.Discount,
		MaxUses:     req.MaxUses,
	})
}

func (s *ServerV2) GetPromocodeById(w http.ResponseWriter, r *http.Request, promocodeId int) {
	promo, err := s.subs.GetPromocodeUC.ByID(r.Context(), domain.PromocodeID(promocodeId))
	if err != nil {
		errResp(w, http.StatusNotFound, "promocode not found")
		return
	}
	utils.MakeResponse(w, http.StatusOK, promoToV2(promo))
}

func (s *ServerV2) UpdatePromocode(w http.ResponseWriter, r *http.Request, promocodeId int) {
	var req apiv2.PromocodeCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp(w, http.StatusBadRequest, "invalid json")
		return
	}

	cur, err := s.subs.GetPromocodeUC.ByID(r.Context(), domain.PromocodeID(promocodeId))
	if err != nil {
		errResp(w, http.StatusNotFound, "promocode not found")
		return
	}

	expiresAt := cur.ExpiresAt
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	err = s.subs.UpdatePromocodeUC.Update(r.Context(), usecase.UpdatePromocodeInput{
		ID:           domain.PromocodeID(promocodeId),
		ServiceID:    req.ServiceId,
		Value:        req.Value,
		PlanID:       req.PlanId,
		ExpiresAt:    expiresAt,
		Discount:     req.Discount,
		MaxUses:      req.MaxUses,
		CurUses:      cur.CurUses,
		Status:       cur.Status,
		DurationDays: intPtrVal(req.DurationDays),
	})
	if err != nil {
		errResp(w, http.StatusNotFound, "promocode not found")
		return
	}

	promo, _ := s.subs.GetPromocodeUC.ByID(r.Context(), domain.PromocodeID(promocodeId))
	utils.MakeResponse(w, http.StatusOK, promoToV2(promo))
}

func (s *ServerV2) DeletePromocode(w http.ResponseWriter, r *http.Request, promocodeId int) {
	if err := s.subs.DeletePromocodeUC.Delete(r.Context(), domain.PromocodeID(promocodeId)); err != nil {
		errResp(w, http.StatusNotFound, "promocode not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- Analytics ----------

func (s *ServerV2) GetTotalCosts(w http.ResponseWriter, r *http.Request, params apiv2.GetTotalCostsParams) {
	stDate, err := parseMonth(params.StartDate)
	if err != nil {
		errResp(w, http.StatusBadRequest, "invalid start_date")
		return
	}
	var enDate time.Time
	if params.EndDate != nil {
		enDate, err = parseMonth(*params.EndDate)
		if err != nil {
			errResp(w, http.StatusBadRequest, "invalid end_date")
			return
		}
	}

	sum, subs, err := s.subs.TotalCostsUC.TotalCosts(r.Context(), usecase.SubsFilterDTO{
		StartDate: stDate,
		EndDate:   enDate,
		UserID:    params.UserId,
	})
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to get total costs: "+err.Error())
		return
	}

	subIDs := make([]int, 0, len(subs))
	subsV2 := make([]apiv2.Subscription, 0, len(subs))
	for _, sub := range subs {
		subIDs = append(subIDs, int(sub.SubId))
		subsV2 = append(subsV2, subDTOToV2(usecase.SubToDTO(sub)))
	}

	utils.MakeResponse(w, http.StatusOK, apiv2.TotalCosts{
		TotalSum:      sum,
		SubIds:        subIDs,
		Subscriptions: subsV2,
	})
}

func (s *ServerV2) GetStatsOverview(w http.ResponseWriter, r *http.Request) {
	out, err := s.subs.StatsOverviewUC.GetOverview(r.Context())
	if err != nil {
		errResp(w, http.StatusInternalServerError, "failed to get stats: "+err.Error())
		return
	}

	res := apiv2.StatsOverview{
		TotalSubscriptions:   &out.TotalSubscriptions,
		ActiveSubscriptions:  &out.ActiveSubscriptions,
		TotalPromocodes:      &out.TotalPromocodes,
		ActivePromocodes:     &out.ActivePromocodes,
		TotalRevenue:         &out.TotalRevenue,
		MostPopularService:   &out.MostPopularService,
	}
	avg := float32(out.AvgSubscriptionPrice)
	res.AvgSubscriptionPrice = &avg

	stats := make([]apiv2.ServiceStats, 0, len(out.ServiceStats))
	for _, ss := range out.ServiceStats {
		name := ss.ServiceName
		subs := ss.TotalSubscriptions
		rev := ss.TotalRevenue
		avgCost := float32(ss.AvgSubscriptionCost)
		stats = append(stats, apiv2.ServiceStats{
			ServiceName:         &name,
			TotalSubscriptions:  &subs,
			TotalRevenue:        &rev,
			AvgSubscriptionCost: &avgCost,
		})
	}
	res.ServiceStats = &stats

	utils.MakeResponse(w, http.StatusOK, res)
}

// ---------- мелкие хелперы (указатели) ----------

func intPtr(v int) *int    { return &v }
func intPtrVal(v *int) int { if v == nil { return 0 }; return *v }
func stringPtr(s *string) string { if s == nil { return "" }; return *s }
