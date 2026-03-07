package user

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildPublicReadPolicy(t *testing.T) {
	policy, err := buildPublicReadPolicy("avatars")
	if err != nil {
		t.Fatalf("buildPublicReadPolicy() error = %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal([]byte(policy), &decoded); err != nil {
		t.Fatalf("policy is not valid json: %v", err)
	}

	if !strings.Contains(policy, "\"s3:GetObject\"") {
		t.Fatalf("policy should contain s3:GetObject action, got: %s", policy)
	}

	if !strings.Contains(policy, "arn:aws:s3:::avatars/*") {
		t.Fatalf("policy should contain avatar resource ARN, got: %s", policy)
	}
}
