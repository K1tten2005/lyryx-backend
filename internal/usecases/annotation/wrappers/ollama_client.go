package wrappers

import (
	"context"
	"fmt"

	ollamaClientPkg "github.com/K1tten2005/lyryx-backend/internal/clients/ollama"
)

type ollamaClient interface {
	PostGenerate(ctx context.Context, userRequest string) (ollamaClientPkg.PostGenerateOut, error)
}

type OllamaAnnotationGetter struct {
	ollamaClient ollamaClient
}

func NewOllamaAnnotationGetter(ollamaClient ollamaClient) *OllamaAnnotationGetter {
	return &OllamaAnnotationGetter{
		ollamaClient: ollamaClient,
	}
}

func (og *OllamaAnnotationGetter) GetAiAnnotation(ctx context.Context, prompt string) (string, error) {
	resp, err := og.ollamaClient.PostGenerate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("post generate: %v", err)
	}
	return resp.Response, nil
}
