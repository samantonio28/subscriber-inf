package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
	"github.com/samantonio28/subscriber-inf/pkg/utils"
)

type PromocodeServer struct {
	createPromocodeUC *usecase.CreatePromocodeUC
	deletePromocodeUC *usecase.DeletePromocodeUC
	getPromocodeUC    *usecase.GetPromocodeUC
	logger            *logger.LogrusLogger
}

func NewPromocodeServer(
	createUC *usecase.CreatePromocodeUC,
	deleteUC *usecase.DeletePromocodeUC,
	getUC *usecase.GetPromocodeUC,
	logger *logger.LogrusLogger,
) *PromocodeServer {
	return &PromocodeServer{
		createPromocodeUC: createUC,
		deletePromocodeUC: deleteUC,
		getPromocodeUC:    getUC,
		logger:            logger,
	}
}

// CreatePromocode handles POST /promocodes
func (s *PromocodeServer) CreatePromocode(w http.ResponseWriter, r *http.Request) {
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

	// Convert expires_at string to time.Time (optional)
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

	id, err := s.createPromocodeUC.Create(r.Context(), input)
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

// GetPromocodeByID handles GET /promocodes/{id}
func (s *PromocodeServer) GetPromocodeByID(w http.ResponseWriter, r *http.Request, id int) {
	promocode, err := s.getPromocodeUC.ByID(r.Context(), domain.PromocodeID(id))
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

// GetPromocodeByCode handles GET /promocodes/code/{code}
func (s *PromocodeServer) GetPromocodeByCode(w http.ResponseWriter, r *http.Request, code string) {
	promocode, err := s.getPromocodeUC.ByCode(r.Context(), code)
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

// DeletePromocode handles DELETE /promocodes/{id}
func (s *PromocodeServer) DeletePromocode(w http.ResponseWriter, r *http.Request, id int) {
	err := s.deletePromocodeUC.Delete(r.Context(), domain.PromocodeID(id))
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

// GetPromocodes handles GET /promocodes with optional filtering
func (s *PromocodeServer) GetPromocodes(w http.ResponseWriter, r *http.Request, serviceID *int, status *string) {
	// For simplicity, return not implemented; can be extended later
	utils.MakeResponse(w, http.StatusNotImplemented, map[string]string{
		"message": "filtered list not implemented yet",
	})
}
