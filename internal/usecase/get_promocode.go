package usecase

import (
	"context"

	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type GetPromocodeUC struct {
	promocodeRepo domain.PromocodeRepository
	promoCache    domain.PromocodeCache
	logger        logger.Logger
}

func NewGetPromocodeUC(promocodeRepo domain.PromocodeRepository, promoCache domain.PromocodeCache, logger logger.Logger) (*GetPromocodeUC, error) {
	if promocodeRepo == nil {
		return nil, domain.ErrInvalidSubRepo
	}
	if promoCache == nil {
		return nil, domain.ErrInvalidCache
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &GetPromocodeUC{promocodeRepo: promocodeRepo, promoCache: promoCache, logger: logger}, nil
}

func (uc *GetPromocodeUC) ByID(ctx context.Context, id domain.PromocodeID) (domain.Promocode, error) {
	uc.logger.Debug("fetching promocode by ID", "promocode_id", id, "source", "database")
	promocode, err := uc.promocodeRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("failed to fetch promocode by ID", "promocode_id", id, "error", err, "source", "database")
		return domain.Promocode{}, err
	}
	uc.logger.Info("promocode fetched successfully", "promocode_id", id, "value", promocode.Value, "source", "database")
	return promocode, nil
}

func (uc *GetPromocodeUC) ByCode(ctx context.Context, code string) (domain.Promocode, error) {
	// Если есть кэш, пытаемся получить из него
	if uc.promoCache != nil {
		uc.logger.Debug("attempting to get promocode from cache", "code", code, "cache_key", code)
		promo, err := uc.promoCache.GetPromocode(ctx, code)
		if err == nil {
			uc.logger.Info("promocode cache hit", "code", code, "promocode_id", promo.PromocodeID, "source", "cache")
			return promo, nil
		}
		uc.logger.Debug("promocode cache miss", "code", code, "error", err)
	}
	// Если в кэше нет или кэш отсутствует, идём в репозиторий
	uc.logger.Debug("fetching promocode by code from database", "code", code, "source", "database")
	promocode, err := uc.promocodeRepo.GetByCode(ctx, code)
	if err != nil {
		uc.logger.Error("failed to fetch promocode by code", "code", code, "error", err, "source", "database")
		return domain.Promocode{}, err
	}
	// Сохраняем в кэш, если он есть
	if uc.promoCache != nil {
		uc.logger.Debug("caching promocode", "code", code, "promocode_id", promocode.PromocodeID, "ttl", "default")
		err = uc.promoCache.SetPromocode(ctx, code, promocode, 0)
		if err != nil {
			uc.logger.Warn("promocode was not saved in cache, continuing", "code", code, "error", err)
		} else {
			uc.logger.Info("promocode retrieved from db and cached", "code", code, "promocode_id", promocode.PromocodeID, "source", "database")
		}
	} else {
		uc.logger.Info("promocode fetched successfully", "code", code, "promocode_id", promocode.PromocodeID, "source", "database")
	}
	return promocode, nil
}
