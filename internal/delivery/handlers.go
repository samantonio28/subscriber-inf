package delivery

import (
	"context"
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
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

type CostsFilter struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Filter    struct {
		UserId      string `json:"user_id"`
		ServiceName string `json:"service_name"`
	} `json:"filter"`
}

func NewSubsHandler(repo domain.SubscriptionRepository, logger *logger.LogrusLogger) (*SubsHandler, error) {
	createSubUC, err := usecase.NewCreateSubUC(repo, logger)
	if err != nil {
		return nil, err
	}
	deleteSubUC, err := usecase.NewDeleteSubUC(repo, logger)
	if err != nil {
		return nil, err
	}
	getSubUC, err := usecase.NewGetSubUC(repo, logger)
	if err != nil {
		return nil, err
	}
	getSubsUC, err := usecase.NewGetSubsUC(repo, logger)
	if err != nil {
		return nil, err
	}
	totalCostsUC, err := usecase.NewTotalCostsUC(repo, logger)
	if err != nil {
		return nil, err
	}
	updateSubUC, err := usecase.NewUpdateSubUC(repo, logger)
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
		StartDate:   stDate,
		EndDate:     enDate,
	}
	return subDTO, nil
}

func (h *SubsHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var err error
	var req HandlingSub
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
	}

	subDTO, err := SerializeSub(req)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad with serializing sub: " + err.Error(),
		})
		return
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
	err = h.DeleteSubUC.DeleteSub(context.Background(), subId)
	if err != nil {
		if err.Error() == "no subs deleted" {
			utils.MakeResponse(w, http.StatusNotFound, map[string]string{
				"message": "bad deleting sub:" + err.Error(),
			})
			return
		}
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad deleting sub:" + err.Error(),
		})
		return
	}
	utils.MakeResponse(w, http.StatusNoContent, map[string]string{
		"message": "nice",
	})
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
	sub, err := h.GetSubUC.SubById(context.Background(), subId)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad getting sub: " + err.Error(),
		})
		return
	}

	stDate := utils.DateString(sub.StartDate)
	enDate := utils.DateString(sub.EndDate)

	var hSub HandlingSub = HandlingSub{
		ServiceName: sub.ServiceName,
		Price:       sub.Price,
		UserId:      sub.UserId.String(),
		StartDate:   stDate,
		EndDate:     enDate,
	}
	utils.MakeResponse(w, http.StatusOK, hSub)
}

func (h *SubsHandler) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserId string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid json",
		})
		return
	}
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "invalid user id: " + err.Error(),
		})
		return
	}
	subs, err := h.GetSubsUC.SubsByUserId(context.Background(), userId)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad getting subs: " + err.Error(),
		})
		return
	}
	hSubs := make([]HandlingSub, 0, len(subs))
	for _, s := range subs {
		stDate := utils.DateString(s.StartDate)
		enDate := utils.DateString(s.EndDate)

		var hSub HandlingSub = HandlingSub{
			ServiceName: s.ServiceName,
			Price:       s.Price,
			UserId:      s.UserId.String(),
			StartDate:   stDate,
			EndDate:     enDate,
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
	stDate, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad start date: " + err.Error(),
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
		enDate, _ = utils.ParseMonthYear(ZeroDateString)
	}

	if req.Filter.ServiceName == "" {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "service name mustn't be empty",
		})
		return
	}
	var uID uuid.UUID
	if req.Filter.UserId != "" {
		uID, err = uuid.Parse(req.Filter.UserId)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "can't parse uuid:" + err.Error(),
			})
			return
		}
	} else {
		uID = uuid.Nil
	}
	filter := usecase.SubsFilterDTO{
		StartDate:   stDate,
		EndDate:     enDate,
		UserID:      uID,
		ServiceName: req.Filter.ServiceName,
	}
	sum, subIds, err := h.TotalCostsUC.TotalCosts(context.Background(), filter)
	if err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad getting total costs: " + err.Error(),
		})
		return
	}
	var ans struct {
		TotalSum int   `json:"total_sum"`
		SubIds   []int `json:"sub_ids"`
	}
	ans.TotalSum = sum
	ans.SubIds = subIds
	utils.MakeResponse(w, http.StatusOK, ans)
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

	if req.ServiceName == "" {
		req.ServiceName = " "
	}

	var uID uuid.UUID
	if req.UserId != "" {
		uID, err = uuid.Parse(req.UserId)
		if err != nil {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "can't parse uuid: " + err.Error(),
			})
			return
		}
	} else {
		uID = uuid.Nil
	}

	stDate, err := utils.ParseMonthYear(req.StartDate)
	if err != nil {
		if err.Error() == "empty date" {
			stDate, _ = utils.ParseMonthYear(ZeroDateString)
		} else {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "bad parsing start date: " + err.Error(),
			})
			return
		}
	}
	var enDate time.Time
	if req.EndDate != "" {
		enDate, err = utils.ParseMonthYear(req.EndDate)
		if err != nil && err.Error() != "empty date" {
			utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
				"message": "bad parsing end date: " + err.Error(),
			})
			return
		}
	} else {
		enDate, _ = utils.ParseMonthYear(ZeroDateString)
	}

	subDTO := usecase.SubscriptionDTO{
		SubId:       0,
		UserId:      uID,
		ServiceName: req.ServiceName,
		Price:       req.Price,
		StartDate:   stDate,
		EndDate:     enDate,
	}

	if err := h.UpdateSubUC.UpdateSub(context.Background(), subId, subDTO); err != nil {
		utils.MakeResponse(w, http.StatusBadRequest, map[string]string{
			"message": "bad updating sub: " + err.Error(),
		})
		return
	}
	utils.MakeResponse(w, http.StatusOK, map[string]string{
		"message": "subscription updated",
	})
}
