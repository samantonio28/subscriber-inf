package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type SubsHandler struct {
	CreateSubUC  usecase.CreateSubUC
	DeleteSubUC  usecase.DeleteSubUC
	GetSubUC     usecase.GetSubUC
	GetSubsUC    usecase.GetSubsUC
	TotalCostsUC usecase.TotalCostsUC
	UpdateSubUC  usecase.UpdateSubUC
}

type HandlingSub struct {
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	UserId      string `json:"user_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

func NewSubsHandler(repo domain.SubscriptionRepository) (*SubsHandler, error) {
	createSubUC, err := usecase.NewCreateSubUC(repo)
	if err != nil {
		return nil, err
	}
	deleteSubUC, err := usecase.NewDeleteSubUC(repo)
	if err != nil {
		return nil, err
	}
	getSubUC, err := usecase.NewGetSubUC(repo)
	if err != nil {
		return nil, err
	}
	getSubsUC, err := usecase.NewGetSubsUC(repo)
	if err != nil {
		return nil, err
	}
	totalCostsUC, err := usecase.NewTotalCostsUC(repo)
	if err != nil {
		return nil, err
	}
	updateSubUC, err := usecase.NewUpdateSubUC(repo)
	if err != nil {
		return nil, err
	}
	return &SubsHandler{
		CreateSubUC:  *createSubUC,
		DeleteSubUC:  *deleteSubUC,
		GetSubUC:     *getSubUC,
		GetSubsUC:    *getSubsUC,
		TotalCostsUC: *totalCostsUC,
		UpdateSubUC:  *updateSubUC,
	}, nil
}

func (h *SubsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req HandlingSub
	var err error
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}
	if req.ServiceName == "" {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "service name mustn't be empty",
		})
		return
	}
	var uID uuid.UUID
	uID, err = uuid.Parse(req.UserId)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "can't parse uuid",
		})
		return
	}
	if req.Price < 0 {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "price is zero or positive",
		})
		return
	}
	stDate, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "parsing start date: " + err.Error(),
		})
		return
	}
	var enDate time.Time
	if req.EndDate != "" {
		enDate, err = utils.ParseMonthYear(req.EndDate)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "parsing end date: " + err.Error(),
			})
			return
		}
	} else {
		enDate, _ = utils.ParseMonthYear("01-0001")
	}

	subDTO := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      uID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   stDate,
		EndDate:     enDate,
	}
	subId, err := h.CreateSubUC.NewSub(context.Background(), subDTO)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad with creating new sub: " + err.Error(),
		})
		return
	}
	utils.MakeResponse(w, http.StatusCreated, map[string]string{
		"message": fmt.Sprintf("new sub_id: %d", subId),
	})
}
