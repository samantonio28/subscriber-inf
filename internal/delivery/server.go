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
	// Subscription plan usecases
	CreateSubscriptionPlanUC usecase.CreateSubscriptionPlanUC
	GetSubscriptionPlanUC    usecase.GetSubscriptionPlanUC
	UpdateSubscriptionPlanUC usecase.UpdateSubscriptionPlanUC
	DeleteSubscriptionPlanUC usecase.DeleteSubscriptionPlanUC
	logger                   *logger.LogrusLogger
}

func NewSubsServer(
	repo domain.SubscriptionRepository,
	promoRepo domain.PromocodeRepository,
	planRepo domain.SubscriptionPlanRepository,
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
		// Subscription plan
		CreateSubscriptionPlanUC: *createPlanUC,
		GetSubscriptionPlanUC:    *getPlanUC,
		UpdateSubscriptionPlanUC: *updatePlanUC,
		DeleteSubscriptionPlanUC: *deletePlanUC,
		logger:                   logger,
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

	subs := make([]api.Subscription, 0, len(dtos))
	for _, dto := range dtos {
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
		subs = append(subs, sub)
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

	sum, subIds, err := s.TotalCostsUC.TotalCosts(r.Context(), dto)
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

	int64SubIds := make([]int64, len(subIds))
	for i, id := range subIds {
		int64SubIds[i] = int64(id)
	}

	response := map[string]any{
		"total_sum": int64(sum),
		"sub_ids":   int64SubIds,
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
	// For simplicity, return not implemented; can be extended later
	utils.MakeResponse(w, http.StatusNotImplemented, map[string]string{
		"message": "filtered list not implemented yet",
	})
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
		if errors.Is(err, domain.ErrPromocodeNotFound) {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"error": "promocode not found",
			})
			return
		}
		s.logger.Error("failed to delete promocode", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete promocode: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusNoContent, nil)
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
	// For simplicity, return not implemented; can be extended later
	utils.MakeResponse(w, http.StatusNotImplemented, map[string]string{
		"message": "filtered list not implemented yet",
	})
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
				"error": "subscription plan not found",
			})
			return
		}
		s.logger.Error("failed to delete subscription plan", "error", err)
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to delete subscription plan: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusNoContent, nil)
}
