package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/dto"
	"github.com/K1tten2005/lyryx-backend/internal/usecases/user/wrappers"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameTaken      = errors.New("username already exists")
)

const (
	hashCost = 12
)

type storage interface {
	GetUserByID(ctx context.Context, userID int) (dto.User, error)
	PatchUpdateUser(ctx context.Context, opts dto.PatchUpdateUserOpts) error
	PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error
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

func (u *Usecase) PatchUpdateUser(ctx context.Context, opts dto.PatchUpdateUserOpts) error {
	if opts.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*opts.Password), hashCost)
		if err != nil {
			return fmt.Errorf("hash password: %v", err)
		}

		hashedPasswordStr := string(hashedPassword)
		opts.Password = &hashedPasswordStr
	}

	err := u.storage.PatchUpdateUser(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return ErrUserNotFound
		}
		if errors.Is(err, wrappers.ErrEmailAlreadyExists) {
			return ErrEmailAlreadyExists
		}
		if errors.Is(err, wrappers.ErrUsernameTaken) {
			return ErrUsernameTaken
		}
		return fmt.Errorf("patch update user: %v", err)
	}

	return nil
}

func (u *Usecase) PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error {
	err := u.storage.PatchUpdateAvatar(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("patch update avatar: %v", err)
	}

	return nil
}
