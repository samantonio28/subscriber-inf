package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/samantonio28/subscriber-inf/internal/domain"
	"github.com/samantonio28/subscriber-inf/internal/logger"
)

type CreateUserUC struct {
	userR  domain.UserRepository
	logger logger.Logger
}

func NewCreateUserUC(userR domain.UserRepository, logger logger.Logger) (*CreateUserUC, error) {
	if userR == nil {
		return nil, domain.ErrInvalidUserRepo
	}
	if logger == nil {
		return nil, domain.ErrInvalidLogger
	}
	return &CreateUserUC{userR: userR, logger: logger}, nil
}

func (u *CreateUserUC) NewUser(ctx context.Context, input UserDTO) (uuid.UUID, error) {
	u.logger.Debug("creating new user", "input", input)
	user, err := DTOToUser(input)
	if err != nil {
		u.logger.Error("failed to convert DTO to user", "error", err, "input", input)
		return uuid.Nil, err
	}
	if user.UserID == uuid.Nil {
		u.logger.Info("generating new user ID")
		user.UserID = uuid.New()
	}
	err = u.userR.StoreUser(ctx, user)
	if err != nil {
		u.logger.Error("failed to store user", "error", err, "user", user)
		return uuid.Nil, err
	}
	u.logger.Info("user created successfully", "user_id", user.UserID, "referral_code", user.ReferralCode)
	if input.ReferralCode != nil && *input.ReferralCode != "" {
		inviter, err := u.userR.GetUserByReferralCode(ctx, *input.ReferralCode)
		if err != nil {
			u.logger.Error("failed to find inviter by referral code", "referral_code", *input.ReferralCode, "error", err)
		} else {
			err = u.userR.StoreReferral(ctx, inviter.UserID, user.UserID)
			if err != nil {
				u.logger.Error("failed to store referral", "referrer_id", inviter.UserID, "referred_id", user.UserID, "error", err)
			} else {
				u.logger.Info("referral link created", "referrer_id", inviter.UserID, "referred_id", user.UserID)
			}
		}
	}
	return user.UserID, nil
}