package delivery

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/api"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type SubsServer struct {
	CreateSubUC  usecase.CreateSubUC
	DeleteSubUC  usecase.DeleteSubUC
	GetSubUC     usecase.GetSubUC
	GetSubsUC    usecase.GetSubsUC
	TotalCostsUC usecase.TotalCostsUC
	UpdateSubUC  usecase.UpdateSubUC
	// Promocode usecases
	CreatePromocodeUC usecase.CreatePromocodeUC
	DeletePromocodeUC usecase.DeletePromocodeUC
	GetPromocodeUC    usecase.GetPromocodeUC
	UpdatePromocodeUC usecase.UpdatePromocodeUC
	ApplyPromocodeUC  usecase.ApplyPromocodeUC
	// Subscription plan usecases
	CreateSubscriptionPlanUC usecase.CreateSubscriptionPlanUC
	GetSubscriptionPlanUC    usecase.GetSubscriptionPlanUC
	UpdateSubscriptionPlanUC usecase.UpdateSubscriptionPlanUC
	DeleteSubscriptionPlanUC usecase.DeleteSubscriptionPlanUC
	// Stats overview usecase
	StatsOverviewUC usecase.StatsOverviewUC
	// Filtering usecases
	GetFilteredPromocodesUC        usecase.GetFilteredPromocodesUC
	GetFilteredSubscriptionPlansUC usecase.GetFilteredSubscriptionPlansUC
	logger                         *logger.LogrusLogger
}

func NewSubsServer(
	repo domain.SubscriptionRepository,
	promoRepo domain.PromocodeRepository,
	planRepo domain.SubscriptionPlanRepository,
	statsService domain.StatsService,
	logger *logger.LogrusLogger,
) (*SubsServer, error) {
	createSubUC, err := usecase.NewCreateSubUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreateSubUC: %w", err)
	}
	deleteSubUC, err := usecase.NewDeleteSubUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DeleteSubUC: %w", err)
	}
	getSubUC, err := usecase.NewGetSubUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetSubUC: %w", err)
	}
	getSubsUC, err := usecase.NewGetSubsUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetSubsUC: %w", err)
	}
	totalCostsUC, err := usecase.NewTotalCostsUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create TotalCostsUC: %w", err)
	}
	updateSubUC, err := usecase.NewUpdateSubUC(repo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create UpdateSubUC: %w", err)
	}

	// Promocode usecases
	createPromoUC, err := usecase.NewCreatePromocodeUC(promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreatePromocodeUC: %w", err)
	}
	deletePromoUC, err := usecase.NewDeletePromocodeUC(promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DeletePromocodeUC: %w", err)
	}
	getPromoUC, err := usecase.NewGetPromocodeUC(promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetPromocodeUC: %w", err)
	}
	updatePromoUC, err := usecase.NewUpdatePromocodeUC(promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create UpdatePromocodeUC: %w", err)
	}
	applyPromoUC, err := usecase.NewApplyPromocodeUC(repo, promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create ApplyPromocodeUC: %w", err)
	}
	statsOverviewUC, err := usecase.NewStatsOverviewUC(repo, promoRepo, statsService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create StatsOverviewUC: %w", err)
	}

	// Filtering usecases
	getFilteredPromosUC, err := usecase.NewGetFilteredPromocodesUC(promoRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetFilteredPromocodesUC: %w", err)
	}
	getFilteredPlansUC, err := usecase.NewGetFilteredSubscriptionPlansUC(planRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetFilteredSubscriptionPlansUC: %w", err)
	}

	// Subscription plan usecases
	createPlanUC, err := usecase.NewCreateSubscriptionPlanUC(planRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create CreateSubscriptionPlanUC: %w", err)
	}
	getPlanUC, err := usecase.NewGetSubscriptionPlanUC(planRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create GetSubscriptionPlanUC: %w", err)
	}
	updatePlanUC, err := usecase.NewUpdateSubscriptionPlanUC(planRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create UpdateSubscriptionPlanUC: %w", err)
	}
	deletePlanUC, err := usecase.NewDeleteSubscriptionPlanUC(planRepo, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create DeleteSubscriptionPlanUC: %w", err)
	}

	return &SubsServer{
		CreateSubUC:  *createSubUC,
		DeleteSubUC:  *deleteSubUC,
		GetSubUC:     *getSubUC,
		GetSubsUC:    *getSubsUC,
		TotalCostsUC: *totalCostsUC,
		UpdateSubUC:  *updateSubUC,
		// Promocode
		CreatePromocodeUC: *createPromoUC,
		DeletePromocodeUC: *deletePromoUC,
		GetPromocodeUC:    *getPromoUC,
		UpdatePromocodeUC: *updatePromoUC,
		ApplyPromocodeUC:  *applyPromoUC,
		// Subscription plan
		CreateSubscriptionPlanUC: *createPlanUC,
		GetSubscriptionPlanUC:    *getPlanUC,
		UpdateSubscriptionPlanUC: *updatePlanUC,
		DeleteSubscriptionPlanUC: *deletePlanUC,
		// Stats overview
		StatsOverviewUC: *statsOverviewUC,
		// Filtering
		GetFilteredPromocodesUC:        *getFilteredPromosUC,
		GetFilteredSubscriptionPlansUC: *getFilteredPlansUC,
		logger:                         logger,
	}, nil
}

func (s *SubsServer) GetSubscriptions(w http.ResponseWriter, r *http.Request, params api.GetSubscriptionsParams) {
	userId := params.Uuid

	dtos, err := s.GetSubsUC.SubsByUserId(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscriptions not found",
			})
			return
		}
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get subscriptions: " + err.Error(),
		})
		return
	}

	type subscriptionResponse struct {
		SubId       int64   `json:"sub_id"`
		ServiceName string  `json:"service_name"`
		Price       int64   `json:"price"`
		UserId      string  `json:"user_id"`
		SubType     string  `json:"sub_type"`
		StartDate   string  `json:"start_date"`
		EndDate     *string `json:"end_date,omitempty"`
	}

	subs := make([]subscriptionResponse, 0, len(dtos))
	for _, dto := range dtos {
		resp := subscriptionResponse{
			SubId:       int64(dto.SubId),
			ServiceName: dto.ServiceName,
			Price:       int64(dto.Price),
			UserId:      dto.UserId.String(),
			SubType:     dto.SubType,
			StartDate:   dto.StartDate.Format("01-2006"),
		}
		if !dto.EndDate.IsZero() {
			endDate := dto.EndDate.Format("01-2006")
			resp.EndDate = &endDate
		}
		subs = append(subs, resp)
	}

	utils.MakeResponse(w, http.StatusOK, subs)
}

func (s *SubsServer) PostSubscriptions(w http.ResponseWriter, r *http.Request) {
	var req api.Subscription
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}

	var uID uuid.UUID
	if req.UserId != uuid.Nil {
		uID = req.UserId
	} else {
		uID = uuid.Nil
	}

	stDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid start date: " + err.Error(),
		})
		return
	}

	var enDate time.Time
	if req.EndDate != nil {
		enDate, err = time.Parse("01-2006", *req.EndDate)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "invalid end date: " + err.Error(),
			})
			return
		}
	} else {
		enDate = time.Time{}
	}

	dto := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      uID,
		ServiceName: req.ServiceName,
		Price:       int(req.Price),
		SubType:     string(*req.SubType),
		StartDate:   stDate,
		EndDate:     enDate,
	}

	subId, err := s.CreateSubUC.NewSub(r.Context(), dto)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to create subscription: " + err.Error(),
		})
		return
	}
	utils.MakeResponse(w, http.StatusCreated, map[string]string{
		"message": fmt.Sprintf("new sub_id: %d", subId),
	})
}

func (s *SubsServer) DeleteSubscriptionsId(w http.ResponseWriter, r *http.Request, id int) {
	err := s.DeleteSubUC.DeleteSub(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNoSubscriptionDeleted) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscription not found",
			})
			return
		}
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to delete subscription: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusNoContent, nil)
}

// PutPromocodesId updates an existing promocode
func (s *SubsServer) PutPromocodesId(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		ServiceID    int     `json:"service_id"`
		Value        string  `json:"value"`
		PlanID       *int    `json:"plan_id,omitempty"`
		SubID        *int    `json:"sub_id,omitempty"`
		ExpiresAt    *string `json:"expires_at,omitempty"`
		Discount     int     `json:"discount"`
		MaxUses      int     `json:"max_uses"`
		CurUses      int     `json:"cur_uses"`
		Status       string  `json:"status"`
		DurationDays int     `json:"duration_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid JSON",
		})
		return
	}

	var expiresAt time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "invalid expires_at format, use RFC3339",
			})
			return
		}
		expiresAt = t
	}

	status, err := domain.NewPromocodeStatus(req.Status)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid status: " + err.Error(),
		})
		return
	}

	input := usecase.UpdatePromocodeInput{
		ID:           domain.PromocodeID(id),
		ServiceID:    req.ServiceID,
		Value:        req.Value,
		PlanID:       req.PlanID,
		SubID:        req.SubID,
		ExpiresAt:    expiresAt,
		Discount:     req.Discount,
		MaxUses:      req.MaxUses,
		CurUses:      req.CurUses,
		Status:       status,
		DurationDays: req.DurationDays,
	}

	err = s.UpdatePromocodeUC.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "promocode not found",
			})
			return
		}
		s.logger.Error("failed to update promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to update promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "promocode updated successfully",
	})
}

// PostSubscriptionsIdApplyPromocode applies a promocode to a subscription
func (s *SubsServer) PostSubscriptionsIdApplyPromocode(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		Promocode string `json:"promocode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid JSON",
		})
		return
	}

	input := usecase.ApplyPromocodeInput{
		SubscriptionID: id,
		PromocodeValue: req.Promocode,
	}

	output, err := s.ApplyPromocodeUC.Apply(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscription not found",
			})
			return
		}
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "promocode not found",
			})
			return
		}
		if errors.Is(err, domain.ErrPromocodeNotActive) ||
			errors.Is(err, domain.ErrPromocodeExpired) ||
			errors.Is(err, domain.ErrPromocodeMaxUsesReached) ||
			errors.Is(err, domain.ErrPromocodeNotApplicable) {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": err.Error(),
			})
			return
		}
		s.logger.Error("failed to apply promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to apply promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]any{
		"message":          output.Message,
		"discount_applied": output.DiscountApplied,
		"new_price":        output.NewPrice,
	})
}

// GetStatsOverview returns system overview statistics
func (s *SubsServer) GetStatsOverview(w http.ResponseWriter, r *http.Request) {
	output, err := s.StatsOverviewUC.GetOverview(r.Context())
	if err != nil {
		s.logger.Error("failed to get stats overview", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get stats overview: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, output)
}

func (s *SubsServer) GetSubscriptionsId(w http.ResponseWriter, r *http.Request, id int) {
	dto, err := s.GetSubUC.SubById(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscription not found",
			})
			return
		}
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get subscription: " + err.Error(),
		})
		return
	}

	sub := api.Subscription{
		ServiceName: dto.ServiceName,
		Price:       int64(dto.Price),
		UserId:      dto.UserId,
		SubType:     (*api.SubscriptionType)(&dto.SubType),
		StartDate:   dto.StartDate.Format("01-2006"),
	}
	if !dto.EndDate.IsZero() {
		endDate := dto.EndDate.Format("01-2006")
		sub.EndDate = &endDate
	}

	utils.MakeResponse(w, http.StatusOK, sub)
}

func (s *SubsServer) PutSubscriptionsId(w http.ResponseWriter, r *http.Request, id int) {
	var req api.Subscription
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}

	// Convert from API model to DTO
	var uID uuid.UUID
	if req.UserId != uuid.Nil {
		uID = req.UserId
	} else {
		uID = uuid.Nil
	}

	stDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid start date: " + err.Error(),
		})
		return
	}

	var enDate time.Time
	if req.EndDate != nil {
		enDate, err = time.Parse("01-2006", *req.EndDate)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "invalid end date: " + err.Error(),
			})
			return
		}
	} else {
		// Use zero date as default
		enDate = time.Time{}
	}

	dto := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      uID,
		ServiceName: req.ServiceName,
		Price:       int(req.Price),
		SubType:     string(*req.SubType),
		StartDate:   stDate,
		EndDate:     enDate,
	}

	err = s.UpdateSubUC.UpdateSub(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscription not found",
			})
			return
		}
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to update subscription: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "subscription updated successfully",
	})
}

func (s *SubsServer) PostTotalCosts(w http.ResponseWriter, r *http.Request) {
	var req api.PostTotalCostsJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}

	startDate, err := time.Parse("01-2006", req.StartDate)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid start date: " + err.Error(),
		})
		return
	}

	endDate := time.Now()
	if req.EndDate != nil {
		endDate, err = time.Parse("01-2006", *req.EndDate)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "invalid end date: " + err.Error(),
			})
			return
		}
	}

	var userId uuid.UUID
	if req.Filter.UserId != nil {
		userId = *req.Filter.UserId
	}

	var subType string
	if req.Filter.SubType != nil {
		subType = string(*req.Filter.SubType)
	}

	dto := usecase.SubsFilterDTO{
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      userId,
		ServiceName: req.Filter.ServiceName,
		SubType:     subType,
	}

	sum, subs, err := s.TotalCostsUC.TotalCosts(r.Context(), dto)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscriptions not found",
			})
			return
		}
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get total costs: " + err.Error(),
		})
		return
	}

	// Extract sub IDs
	int64SubIds := make([]int64, len(subs))
	subscriptions := make([]map[string]any, len(subs))
	for i, sub := range subs {
		int64SubIds[i] = int64(sub.SubId)
		// Calculate duration months and total cost for period (simplified)
		durationMonths := 0
		totalCostForPeriod := 0
		// TODO: implement proper calculation
		subscriptions[i] = map[string]any{
			"sub_id":                int64(sub.SubId),
			"service_name":          sub.ServiceName,
			"price":                 int64(sub.Price),
			"user_id":               sub.UserID,
			"sub_type":              sub.SubType.String(),
			"start_date":            sub.StartDate.Format("01-2006"),
			"end_date":              sub.EndDate.Format("01-2006"),
			"duration_months":       durationMonths,
			"total_cost_for_period": totalCostForPeriod,
		}
	}

	response := map[string]any{
		"total_sum":     int64(sum),
		"sub_ids":       int64SubIds,
		"subscriptions": subscriptions,
	}
	utils.MakeResponse(w, http.StatusOK, response)
}

// Promocode handlers

func (s *SubsServer) PostPromocodes(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServiceID    int     `json:"service_id"`
		Value        string  `json:"value"`
		PlanID       *int    `json:"plan_id,omitempty"`
		SubID        *int    `json:"sub_id,omitempty"`
		ExpiresAt    *string `json:"expires_at,omitempty"`
		Discount     int     `json:"discount"`
		MaxUses      int     `json:"max_uses"`
		DurationDays int     `json:"duration_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	var expiresAt time.Time
	if req.ExpiresAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid expires_at format, use RFC3339",
			})
			return
		}
		expiresAt = t
	}

	input := usecase.CreatePromocodeInput{
		ServiceID:    req.ServiceID,
		Value:        req.Value,
		PlanID:       req.PlanID,
		SubID:        req.SubID,
		ExpiresAt:    expiresAt,
		Discount:     req.Discount,
		MaxUses:      req.MaxUses,
		DurationDays: req.DurationDays,
	}

	id, err := s.CreatePromocodeUC.Create(r.Context(), input)
	if err != nil {
		s.logger.Error("failed to create promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusCreated, map[string]any{
		"promocode_id": id,
		"message":      "promocode created successfully",
	})
}

func (s *SubsServer) GetPromocodes(w http.ResponseWriter, r *http.Request, params api.GetPromocodesParams) {
	// Преобразовать параметры API в фильтр юзкейса
	filter := usecase.PromocodeFilter{}

	if params.ServiceId != nil {
		filter.ServiceID = params.ServiceId
	}
	if params.PlanId != nil {
		filter.PlanID = params.PlanId
	}
	if params.SubId != nil {
		filter.SubID = params.SubId
	}
	if params.Status != nil {
		status, err := domain.NewPromocodeStatus(string(*params.Status))
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid status value",
			})
			return
		}
		filter.Status = &status
	}
	if params.DiscountMin != nil {
		filter.DiscountMin = params.DiscountMin
	}
	if params.DiscountMax != nil {
		filter.DiscountMax = params.DiscountMax
	}
	if params.ExpiresBefore != nil {
		filter.ExpiresBefore = params.ExpiresBefore
	}
	if params.ExpiresAfter != nil {
		filter.ExpiresAfter = params.ExpiresAfter
	}
	if params.CodeContains != nil {
		filter.CodeContains = params.CodeContains
	}

	promocodes, err := s.GetFilteredPromocodesUC.GetFiltered(r.Context(), filter)
	if err != nil {
		s.logger.Error("failed to get filtered promocodes", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get promocodes: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, promocodes)
}

func (s *SubsServer) GetPromocodesId(w http.ResponseWriter, r *http.Request, id int) {
	promocode, err := s.GetPromocodeUC.ByID(r.Context(), domain.PromocodeID(id))
	if err != nil {
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"error": "promocode not found",
			})
			return
		}
		s.logger.Error("failed to get promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, promocode)
}

func (s *SubsServer) DeletePromocodesId(w http.ResponseWriter, r *http.Request, id int) {
	err := s.DeletePromocodeUC.Delete(r.Context(), domain.PromocodeID(id))
	if err != nil {
		if errors.Is(err, domain.ErrPromocodeNotFound) || errors.Is(err, domain.ErrNoPromocodesDeleted) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "promocode not found",
			})
			return
		}
		s.logger.Error("failed to delete promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to delete promocode: " + err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *SubsServer) GetPromocodesCodeCode(w http.ResponseWriter, r *http.Request, code string) {
	promocode, err := s.GetPromocodeUC.ByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"error": "promocode not found",
			})
			return
		}
		s.logger.Error("failed to get promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, promocode)
}

// Subscription plan handlers

func (s *SubsServer) PostSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ServiceID    int    `json:"service_id"`
		Name         string `json:"name"`
		DurationDays int    `json:"duration_days"`
		Price        int64  `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	input := usecase.CreateSubscriptionPlanInput{
		ServiceID:    req.ServiceID,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        int(req.Price),
	}

	id, err := s.CreateSubscriptionPlanUC.Create(r.Context(), input)
	if err != nil {
		s.logger.Error("failed to create subscription plan", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create subscription plan: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusCreated, map[string]any{
		"plan_id": id,
		"message": "subscription plan created successfully",
	})
}

func (s *SubsServer) GetSubscriptionPlans(w http.ResponseWriter, r *http.Request, params api.GetSubscriptionPlansParams) {
	// Преобразовать параметры API в фильтр юзкейса
	filter := usecase.SubscriptionPlanFilter{}

	if params.ServiceId != nil {
		filter.ServiceID = params.ServiceId
	}
	if params.NameContains != nil {
		filter.NameContains = params.NameContains
	}
	if params.PriceMin != nil {
		filter.PriceMin = params.PriceMin
	}
	if params.PriceMax != nil {
		filter.PriceMax = params.PriceMax
	}
	if params.DurationDaysMin != nil {
		filter.DurationDaysMin = params.DurationDaysMin
	}
	if params.DurationDaysMax != nil {
		filter.DurationDaysMax = params.DurationDaysMax
	}

	plans, err := s.GetFilteredSubscriptionPlansUC.GetFiltered(r.Context(), filter)
	if err != nil {
		s.logger.Error("failed to get filtered subscription plans", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get subscription plans: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, plans)
}

func (s *SubsServer) GetSubscriptionPlansId(w http.ResponseWriter, r *http.Request, id int) {
	plan, err := s.GetSubscriptionPlanUC.ByID(r.Context(), domain.PlanID(id))
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionPlanNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"error": "subscription plan not found",
			})
			return
		}
		s.logger.Error("failed to get subscription plan", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to get subscription plan: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, plan)
}

func (s *SubsServer) PutSubscriptionPlansId(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		ServiceID    int    `json:"service_id"`
		Name         string `json:"name"`
		DurationDays int    `json:"duration_days"`
		Price        int64  `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"error": "invalid JSON",
		})
		return
	}

	input := usecase.UpdateSubscriptionPlanInput{
		ID:           domain.PlanID(id),
		ServiceID:    req.ServiceID,
		Name:         req.Name,
		DurationDays: req.DurationDays,
		Price:        int(req.Price),
	}

	err := s.UpdateSubscriptionPlanUC.Update(r.Context(), input)
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionPlanNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"error": "subscription plan not found",
			})
			return
		}
		s.logger.Error("failed to update subscription plan", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to update subscription plan: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "subscription plan updated successfully",
	})
}

func (s *SubsServer) DeleteSubscriptionPlansId(w http.ResponseWriter, r *http.Request, id int) {
	err := s.DeleteSubscriptionPlanUC.Delete(r.Context(), domain.PlanID(id))
	if err != nil {
		if errors.Is(err, domain.ErrSubscriptionPlanNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "subscription plan not found",
			})
			return
		}
		s.logger.Error("failed to delete subscription plan", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to delete subscription plan: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusNoContent, nil)
}
