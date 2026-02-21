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
	logger       *logger.LogrusLogger
}

func NewSubsServer(repo domain.SubscriptionRepository, logger *logger.LogrusLogger) (*SubsServer, error) {
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
	return &SubsServer{
		CreateSubUC:  *createSubUC,
		DeleteSubUC:  *deleteSubUC,
		GetSubUC:     *getSubUC,
		GetSubsUC:    *getSubsUC,
		TotalCostsUC: *totalCostsUC,
		UpdateSubUC:  *updateSubUC,
		logger:       logger,
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
		if errors.Is(err, fmt.Errorf("no subs deleted")) {
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
