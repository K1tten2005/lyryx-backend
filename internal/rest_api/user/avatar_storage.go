package user

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
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

func (s *MinIOAvatarStorage) EnsureBucketPublic(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket: %w", err)
		}
	}

	policy, err := buildPublicReadPolicy(s.bucket)
	if err != nil {
		return fmt.Errorf("build public read policy: %w", err)
	}

	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}

	return nil
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
		minio.PutObjectOptions{
			ContentType: contentType},
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
	case "image/webp":
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
	return fmt.Sprintf("%d/%d_%s%s", userID, time.Now().Unix(), randomPart, ext), nil
}

func buildPublicReadPolicy(bucket string) (string, error) {
	type statement struct {
		Action    []string `json:"Action"`
		Effect    string   `json:"Effect"`
		Principal string   `json:"Principal"`
		Resource  []string `json:"Resource"`
		Sid       string   `json:"Sid"`
	}

	type policyDocument struct {
		Version   string      `json:"Version"`
		Statement []statement `json:"Statement"`
	}

	policy := policyDocument{
		Version: "2012-10-17",
		Statement: []statement{{
			Sid:       "PublicReadForAvatars",
			Effect:    "Allow",
			Principal: "*",
			Action:    []string{"s3:GetObject"},
			Resource:  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucket)},
		}},
	}

	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("marshal policy: %w", err)
	}

	return string(policyBytes), nil
}
