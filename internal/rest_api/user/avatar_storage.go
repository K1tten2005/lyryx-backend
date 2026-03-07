package user

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const maxAvatarSizeBytes int64 = 5 * 1024 * 1024

var (
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg/gif image")
	ErrAvatarTooLarge    = errors.New("avatar file is too large (max 5MB)")
)

type MinIOAvatarStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewMinIOAvatarStorage(client *minio.Client, bucket string, publicBaseURL string) *MinIOAvatarStorage {
	return &MinIOAvatarStorage{
		client:        client,
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}
}

func (s *MinIOAvatarStorage) UploadAvatar(ctx context.Context, userID int, avatarFile *multipart.FileHeader) (string, error) {
	src, err := avatarFile.Open()
	if err != nil {
		return "", fmt.Errorf("open avatar file: %w", err)
	}
	defer src.Close()

	avatarBytes, err := io.ReadAll(io.LimitReader(src, maxAvatarSizeBytes+1))
	if err != nil {
		return "", fmt.Errorf("read avatar file: %w", err)
	}

	if int64(len(avatarBytes)) > maxAvatarSizeBytes {
		return "", ErrAvatarTooLarge
	}

	contentType, extension, err := detectAvatarFormat(avatarBytes)
	if err != nil {
		return "", err
	}

	objectName, err := buildAvatarObjectName(userID, extension)
	if err != nil {
		return "", fmt.Errorf("build avatar object name: %w", err)
	}

	_, err = s.client.PutObject(
		ctx,
		s.bucket,
		objectName,
		bytes.NewReader(avatarBytes),
		int64(len(avatarBytes)),
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, s.bucket, objectName), nil
}

func detectAvatarFormat(fileBytes []byte) (string, string, error) {
	if _, _, err := image.DecodeConfig(bytes.NewReader(fileBytes)); err != nil {
		return "", "", ErrInvalidAvatarType
	}

	contentType := http.DetectContentType(fileBytes)
	switch contentType {
	case "image/png":
		return contentType, ".png", nil
	case "image/jpeg":
		return contentType, ".jpg", nil
	case "image/gif":
		return contentType, ".gif", nil
	}

	return "", "", ErrInvalidAvatarType
}

func buildAvatarObjectName(userID int, ext string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("random read: %w", err)
	}

	randomPart := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("avatars/%d/%d_%s%s", userID, time.Now().Unix(), randomPart, ext), nil
}
