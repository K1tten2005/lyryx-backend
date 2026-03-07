package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/wrappers"
	"github.com/sirupsen/logrus"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type storage interface {
	GetUserByID(ctx context.Context, userID int) (dto.User, error)
}

type Usecase struct {
	storage storage

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage: storage,
		logger:  logger,
	}
}

func (u *Usecase) GetUserByID(ctx context.Context, userID int) (dto.User, error) {
	user, err := u.storage.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return dto.User{}, ErrUserNotFound
		}
		return dto.User{}, fmt.Errorf("get user by id: %v", err)
	}

	return user, nil
}
