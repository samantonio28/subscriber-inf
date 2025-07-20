package delivery

import (
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/usecase"
)

type SubsHandler struct {
	CreateSubUC  usecase.CreateSubUC
	DeleteSubUC  usecase.DeleteSubUC
	GetSubUC     usecase.GetSubUC
	GetSubsUC    usecase.GetSubsUC
	TotalCostsUC usecase.TotalCostsUC
	UpdateSubUC  usecase.UpdateSubUC
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
