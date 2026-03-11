package storage

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
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

const maxAvatarSizeBytes int64 = 5 * 1024 * 1024

var (
	ErrInvalidAvatarType = errors.New("avatar must be a valid png/jpeg image")
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
		return fmt.Errorf("check bucket exists: %v", err)
	}

	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket: %v", err)
		}
	}

	policy, err := buildPublicReadPolicy(s.bucket)
	if err != nil {
		return fmt.Errorf("build public read policy: %v", err)
	}

	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("set bucket policy: %v", err)
	}

	return nil
}

func (s *MinIOAvatarStorage) UploadAvatar(ctx context.Context, filter UploadAvatarFilter) (string, error) {
	src, err := filter.AvatarFile.Open()
	if err != nil {
		return "", fmt.Errorf("open avatar file: %v", err)
	}
	defer src.Close()

	avatarBytes, err := io.ReadAll(io.LimitReader(src, maxAvatarSizeBytes+1))
	if err != nil {
		return "", fmt.Errorf("read avatar file: %v", err)
	}

	if int64(len(avatarBytes)) > maxAvatarSizeBytes {
		return "", fmt.Errorf("upload avatar: %w", ErrAvatarTooLarge)
	}

	contentType, extension, err := detectAvatarFormat(avatarBytes)
	if err != nil {
		return "", err
	}

	objectName, err := buildAvatarObjectName(filter.ArtistID, extension)
	if err != nil {
		return "", fmt.Errorf("build avatar object name: %v", err)
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
		return "", "", fmt.Errorf("decode avatar format: %w", ErrInvalidAvatarType)
	}

	contentType := http.DetectContentType(fileBytes)
	switch contentType {
	case "image/png":
		return contentType, ".png", nil
	case "image/jpeg":
		return contentType, ".jpg", nil
	}

	return "", "", fmt.Errorf("decode avatar format: %w", ErrInvalidAvatarType)
}

func buildAvatarObjectName(artistID int, ext string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("random read: %v", err)
	}

	randomPart := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d/%d_%s%s", artistID, time.Now().Unix(), randomPart, ext), nil
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
		return "", fmt.Errorf("marshal policy: %v", err)
	}

	return string(policyBytes), nil
}
