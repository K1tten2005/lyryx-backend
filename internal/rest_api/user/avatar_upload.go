package user

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	presignedUploadURLTTL = time.Minute * 15
	s3Region              = "us-east-1"
	s3Service             = "s3"
)

type AvatarUploadService struct {
	endpoint       string
	accessKey      string
	secretKey      string
	bucketName     string
	publicBaseURL  string
	useSSL         bool
	presignExpires time.Duration
}

func NewAvatarUploadService(endpoint, accessKey, secretKey, bucketName, publicBaseURL string, useSSL bool) (*AvatarUploadService, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}
	if strings.TrimSpace(accessKey) == "" {
		return nil, fmt.Errorf("access key is empty")
	}
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("secret key is empty")
	}
	if strings.TrimSpace(bucketName) == "" {
		return nil, fmt.Errorf("bucket name is empty")
	}

	base := strings.TrimSpace(publicBaseURL)
	if base == "" {
		scheme := "http"
		if useSSL {
			scheme = "https"
		}
		base = fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucketName)
	}

	return &AvatarUploadService{
		endpoint:       endpoint,
		accessKey:      accessKey,
		secretKey:      secretKey,
		bucketName:     bucketName,
		publicBaseURL:  strings.TrimRight(base, "/"),
		useSSL:         useSSL,
		presignExpires: presignedUploadURLTTL,
	}, nil
}

func (s *AvatarUploadService) CreatePresignedUploadURL(_ context.Context, userID int, ext string, contentType string) (UploadURLInfo, error) {
	normalizedExt := normalizeExt(ext)
	objectKey, err := generateObjectKey(userID, normalizedExt)
	if err != nil {
		return UploadURLInfo{}, fmt.Errorf("generate object key: %v", err)
	}

	signedURL, err := s.presignPutURL(objectKey)
	if err != nil {
		return UploadURLInfo{}, fmt.Errorf("presign put url: %v", err)
	}

	avatarURL := fmt.Sprintf("%s/%s", s.publicBaseURL, objectKey)

	return UploadURLInfo{
		UploadURL:    signedURL,
		AvatarURL:    avatarURL,
		ObjectKey:    objectKey,
		ExpiresInSec: int(s.presignExpires.Seconds()),
		Method:       "PUT",
		Headers: map[string]string{
			"Content-Type": contentType,
		},
	}, nil
}

func (s *AvatarUploadService) presignPutURL(objectKey string) (string, error) {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}

	canonicalURI := "/" + s.bucketName + "/" + pathEncode(objectKey)
	host := s.endpoint
	expires := int(s.presignExpires.Seconds())

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, s3Region, s3Service)
	query := map[string]string{
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    s.accessKey + "/" + credentialScope,
		"X-Amz-Date":          amzDate,
		"X-Amz-Expires":       fmt.Sprintf("%d", expires),
		"X-Amz-SignedHeaders": "host",
	}

	canonicalQuery := canonicalQueryString(query)
	canonicalHeaders := "host:" + host + "\n"
	signedHeaders := "host"
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := strings.Join([]string{
		"PUT",
		canonicalURI,
		canonicalQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	hashedCanonicalRequest := sha256Hex(canonicalRequest)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hashedCanonicalRequest,
	}, "\n")

	signingKey := getSignatureKey(s.secretKey, dateStamp, s3Region, s3Service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	return fmt.Sprintf("%s://%s%s?%s&X-Amz-Signature=%s", scheme, host, canonicalURI, canonicalQuery, signature), nil
}

func canonicalQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}

	return strings.Join(parts, "&")
}

func pathEncode(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func getSignatureKey(secret, dateStamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	return hmacSHA256(kService, "aws4_request")
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	_, _ = h.Write([]byte(data))
	return h.Sum(nil)
}

func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

func normalizeExt(ext string) string {
	trimmed := strings.TrimSpace(strings.ToLower(ext))
	if trimmed == "" {
		return ".jpg"
	}
	if !strings.HasPrefix(trimmed, ".") {
		trimmed = "." + trimmed
	}
	if filepath.Ext(trimmed) == "" {
		return ".jpg"
	}
	return trimmed
}

func generateObjectKey(userID int, ext string) (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("rand read: %v", err)
	}

	suffix := hex.EncodeToString(buf)
	return fmt.Sprintf("users/%d/%d_%s%s", userID, time.Now().UTC().Unix(), suffix, ext), nil
}
