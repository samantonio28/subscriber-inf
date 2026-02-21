package delivery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

var ZeroDateString = "01-0001"

type SubsHandler struct {
	CreateSubUC  usecase.CreateSubUC
	DeleteSubUC  usecase.DeleteSubUC
	GetSubUC     usecase.GetSubUC
	GetSubsUC    usecase.GetSubsUC
	TotalCostsUC usecase.TotalCostsUC
	UpdateSubUC  usecase.UpdateSubUC
	logger       *logger.LogrusLogger
}

type HandlingSub struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserId      string `json:"user_id"`
	SubType     string `json:"sub_type"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type CostsFilter struct {
	StartDate string  `json:"start_date"`
	EndDate   *string `json:"end_date,omitempty"`
	Filter    struct {
		ServiceName string  `json:"service_name"`
		UserId      *string `json:"user_id,omitempty"`
		SubType     *string `json:"sub_type,omitempty"`
	} `json:"filter"`
}

func NewSubsHandler(repo domain.SubscriptionRepository, logger *logger.LogrusLogger) (*SubsHandler, error) {
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
	return &SubsHandler{
		CreateSubUC:  *createSubUC,
		DeleteSubUC:  *deleteSubUC,
		GetSubUC:     *getSubUC,
		GetSubsUC:    *getSubsUC,
		TotalCostsUC: *totalCostsUC,
		UpdateSubUC:  *updateSubUC,
		logger:       logger,
	}, nil
}

func SerializeSub(req HandlingSub) (usecase.SubscriptionDTO, error) {
	var err error
	if req.ServiceName == "" {
		return usecase.SubscriptionDTO{}, fmt.Errorf("service name mustn't be empty")
	}
	if req.Price < 0 {
		return usecase.SubscriptionDTO{}, fmt.Errorf("price must be zero or positive")
	}

	var uID uuid.UUID
	if req.UserId != "" {
		uID, err = uuid.Parse(req.UserId)
		if err != nil {
			return usecase.SubscriptionDTO{}, fmt.Errorf("can't parse uuid: %v", err)
		}
	} else {
		uID = uuid.Nil
	}

	stDate, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		return usecase.SubscriptionDTO{}, fmt.Errorf("parsing start date: %v", err)
	}
	var enDate time.Time
	if req.EndDate != "" {
		enDate, err = utils.ParseMonthYear(req.EndDate)
		if err != nil {
			return usecase.SubscriptionDTO{}, fmt.Errorf("parsing end date: %v", err)
		}
	} else {
		enDate, _ = utils.ParseMonthYear(ZeroDateString)
	}

	subDTO := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      uID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		SubType:     req.SubType,
		StartDate:   stDate,
		EndDate:     enDate,
	}
	return subDTO, nil
}

func (h *SubsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req HandlingSub
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}

	dto, err := SerializeSub(req)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid subscription data: " + err.Error(),
		})
		return
	}

	subId, err := h.CreateSubUC.NewSub(r.Context(), dto)
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

func (h *SubsHandler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIdSt, ok := vars["id"]
	if !ok {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "has no valid id in query",
		})
		return
	}
	subId, err := strconv.Atoi(subIdSt)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "has no valid id in query:" + err.Error(),
		})
		return
	}

	err = h.DeleteSubUC.DeleteSub(r.Context(), subId)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to delete subscription: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusNoContent, nil)
}

func (h *SubsHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subIdSt, ok := vars["id"]
	if !ok {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "has no valid id in query",
		})
		return
	}
	subId, err := strconv.Atoi(subIdSt)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad sub id: " + err.Error(),
		})
		return
	}

	dto, err := h.GetSubUC.SubById(r.Context(), subId)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get subscription: " + err.Error(),
		})
		return
	}

	hSub := HandlingSub{
		ServiceName: dto.ServiceName,
		Price:       dto.Price,
		UserId:      dto.UserId.String(),
		SubType:     dto.SubType,
		StartDate:   dto.StartDate.Format("01-2006"),
		EndDate:     dto.EndDate.Format("01-2006"),
	}

	utils.MakeResponse(w, http.StatusOK, hSub)
}

func (h *SubsHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	userId, err := uuid.Parse(r.URL.Query().Get("uuid"))
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid user id: " + err.Error(),
		})
		return
	}

	dtos, err := h.GetSubsUC.SubsByUserId(r.Context(), userId)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get subscriptions: " + err.Error(),
		})
		return
	}

	hSubs := make([]HandlingSub, 0, len(dtos))
	for _, dto := range dtos {
		hSub := HandlingSub{
			ServiceName: dto.ServiceName,
			Price:       dto.Price,
			UserId:      dto.UserId.String(),
			SubType:     dto.SubType,
			StartDate:   dto.StartDate.Format("01-2006"),
			EndDate:     dto.EndDate.Format("01-2006"),
		}
		hSubs = append(hSubs, hSub)
	}

	utils.MakeResponse(w, http.StatusOK, hSubs)
}

func (h *SubsHandler) GetTotalCosts(w http.ResponseWriter, r *http.Request) {
	var req CostsFilter
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
		userId, err = uuid.Parse(*req.Filter.UserId)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "invalid user id: " + err.Error(),
			})
			return
		}
	}

	var subType string
	if req.Filter.SubType != nil {
		subType = *req.Filter.SubType
	}

	dto := usecase.SubsFilterDTO{
		StartDate:   startDate,
		EndDate:     endDate,
		UserID:      userId,
		ServiceName: req.Filter.ServiceName,
		SubType:     subType,
	}

	sum, subIds, err := h.TotalCostsUC.TotalCosts(r.Context(), dto)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to get total costs: " + err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"total_sum": sum,
		"sub_ids":   subIds,
	}
	utils.MakeResponse(w, http.StatusOK, response)
}

func (h *SubsHandler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	var req HandlingSub
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}
	vars := mux.Vars(r)
	subId, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid sub id: " + err.Error(),
		})
		return
	}

	dto, err := SerializeSub(req)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid subscription data: " + err.Error(),
		})
		return
	}

	err = h.UpdateSubUC.UpdateSub(r.Context(), subId, dto)
	if err != nil {
		utils.MakeResponse(w, http.StatusInternalServerError, map[string]string{
			"message": "failed to update subscription: " + err.Error(),
		})
		return
	}

	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "subscription updated successfully",
	})
}
