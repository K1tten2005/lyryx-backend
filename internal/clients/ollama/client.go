package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	modelName  string
	baseURL    string
	httpClient http.Client
}

func NewOllamaClient(modelName, baseURL string) *Client {
	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	return &Client{
		modelName: modelName,
		baseURL:   baseURL,
		httpClient: http.Client{
			Transport: tr,
		},
	}
}

func (c *Client) PostGenerate(ctx context.Context, userRequest string) (PostGenerateOut, error) {
	// 1. Формируем тело запроса
	reqBody := GenerateRequest{
		Model:  c.modelName,
		Prompt: userRequest,
		Stream: false,
	}

	// 2. Сериализуем в JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return PostGenerateOut{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 3. Создаём HTTP-запрос
	url := c.baseURL + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return PostGenerateOut{}, fmt.Errorf("failed to create request: %w", err)
	}

	// 4. Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 5. Выполняем запрос
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PostGenerateOut{}, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 6. Проверяем статус код
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return PostGenerateOut{}, fmt.Errorf("ollama API error [%d]: %s", resp.StatusCode, string(body))
	}

	// 7. Читаем и парсим ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PostGenerateOut{}, fmt.Errorf("failed to read response: %w", err)
	}

	var result GenerateResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return PostGenerateOut{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return PostGenerateOut{Response: result.Response}, nil
}
