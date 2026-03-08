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
	ErrInvalidAvatarType  = errors.New("avatar must be a valid png/jpeg image")
	ErrAvatarTooLarge     = errors.New("avatar file is too large (max 5MB)")
)

const (
	hashCost = 12
)

type storage interface {
	GetUserByID(ctx context.Context, userID int) (dto.User, error)
	PatchUpdateUser(ctx context.Context, opts dto.PatchUpdateUserOpts) error
	PatchUpdateAvatar(ctx context.Context, opts dto.PatchUpdateAvatarOpts) error
}

type userAvatarUploader interface {
	UploadAvatar(ctx context.Context, opts dto.UploadAvatarOpts) (string, error)
}

type Usecase struct {
	storage            storage
	userAvatarUploader userAvatarUploader

	logger *logrus.Logger
}

func NewUsecase(
	storage storage,
	userAvatarUploader userAvatarUploader,

	logger *logrus.Logger,
) *Usecase {
	return &Usecase{
		storage:            storage,
		userAvatarUploader: userAvatarUploader,
		logger:             logger,
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

func (u *Usecase) PatchUpdateAvatar(ctx context.Context, opts dto.UploadAvatarOpts) error {
	// 1. Загружаем аватар в minio.
	avatarUrl, err := u.userAvatarUploader.UploadAvatar(ctx, opts)
	if err != nil {
		if errors.Is(err, wrappers.ErrInvalidAvatarType) {
			return fmt.Errorf("patch update avatar: %v", ErrInvalidAvatarType)
		}
		if errors.Is(err, wrappers.ErrAvatarTooLarge) {
			return fmt.Errorf("patch update avatar: %v", ErrAvatarTooLarge)
		}
		return fmt.Errorf("patch update avatar: %v", err)
	}

	// 2. Обновляем ссылку на аватар в бд.
	err = u.storage.PatchUpdateAvatar(ctx, dto.PatchUpdateAvatarOpts{
		UserID:    opts.UserID,
		AvatarURL: avatarUrl,
	})
	if err != nil {
		if errors.Is(err, wrappers.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("patch update avatar: %v", err)
	}

	return nil
}
