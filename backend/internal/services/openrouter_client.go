package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sopeal/AskYourFeed/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var openRouterTracer = otel.Tracer("openrouter_client")

const (
	// OpenRouterBaseURL is the base URL for OpenRouter API
	OpenRouterBaseURL = "https://openrouter.ai/api/v1"

	// MaxImageSize is the maximum image size to process (25 MB)
	MaxImageSize = 25 * 1024 * 1024

	// MaxVideoSize is the maximum video size to process (25 MB)
	MaxVideoSize = 25 * 1024 * 1024

	// MaxVideoDuration is the maximum video duration in seconds (90 seconds)
	MaxVideoDuration = 90

	// MaxImagesPerPost is the maximum number of images to process per post
	MaxImagesPerPost = 4
)

// OpenRouterClient handles communication with OpenRouter API for vision and transcription
type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewOpenRouterClient creates a new OpenRouter API client
func NewOpenRouterClient(apiKey string, httpClient *http.Client) *OpenRouterClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 60 * time.Second, // Longer timeout for vision/transcription
		}
	}
	return &OpenRouterClient{
		apiKey:     apiKey,
		baseURL:    OpenRouterBaseURL,
		httpClient: httpClient,
	}
}

// ChatCompletionRequest represents a request to OpenRouter chat completion API
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Message represents a chat message
type Message struct {
	Role    string        `json:"role"`
	Content []ContentPart `json:"content"`
}

// ContentPart represents a part of message content (text or image)
type ContentPart struct {
	Type     string    `json:"type"`                // "text" or "image_url"
	Text     string    `json:"text,omitempty"`      // For text content
	ImageURL *ImageURL `json:"image_url,omitempty"` // For image content
}

// ImageURL represents an image URL in the message
type ImageURL struct {
	URL string `json:"url"`
}

// ChatCompletionResponse represents the response from OpenRouter
type ChatCompletionResponse struct {
	ID      string    `json:"id"`
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a completion choice
type Choice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	FinishReason string `json:"finish_reason"`
}

// APIError represents an error from OpenRouter API
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// DescribeImage generates a text description of an image using vision model
func (c *OpenRouterClient) DescribeImage(ctx context.Context, imageURL string) (string, error) {
	ctx, span := openRouterTracer.Start(ctx, "DescribeImage")
	defer span.End()

	span.SetAttributes(attribute.String("image_url", imageURL))

	// Prepare the request
	req := ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Using GPT-4o-mini for vision (cost-effective)
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "text",
						Text: "Describe this image in detail. Focus on the main content, text, and any important visual elements. Keep it concise but informative.",
					},
					{
						Type: "image_url",
						ImageURL: &ImageURL{
							URL: imageURL,
						},
					},
				},
			},
		},
	}

	description, err := c.makeCompletionRequest(ctx, req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to describe image: %w", err)
	}

	span.SetAttributes(attribute.Int("description_length", len(description)))
	return description, nil
}

// DescribeImages generates text descriptions for multiple images
func (c *OpenRouterClient) DescribeImages(ctx context.Context, imageURLs []string) ([]string, error) {
	ctx, span := openRouterTracer.Start(ctx, "DescribeImages")
	defer span.End()

	span.SetAttributes(attribute.Int("image_count", len(imageURLs)))

	// Limit to MaxImagesPerPost
	if len(imageURLs) > MaxImagesPerPost {
		logger.Warn("too many images, limiting to max",
			"total", len(imageURLs),
			"max", MaxImagesPerPost)
		imageURLs = imageURLs[:MaxImagesPerPost]
	}

	descriptions := make([]string, 0, len(imageURLs))

	for i, imageURL := range imageURLs {
		description, err := c.DescribeImage(ctx, imageURL)
		if err != nil {
			logger.Warn("failed to describe image, skipping",
				"error", err,
				"image_index", i,
				"image_url", imageURL)
			continue
		}
		descriptions = append(descriptions, description)
	}

	span.SetAttributes(attribute.Int("successful_descriptions", len(descriptions)))
	return descriptions, nil
}

// TranscribeVideo transcribes video content to text
// Note: OpenRouter doesn't directly support video transcription via API
// This is a placeholder for future implementation or alternative service
func (c *OpenRouterClient) TranscribeVideo(ctx context.Context, videoURL string, durationSeconds int, sizeBytes int64) (string, error) {
	ctx, span := openRouterTracer.Start(ctx, "TranscribeVideo")
	defer span.End()

	span.SetAttributes(
		attribute.String("video_url", videoURL),
		attribute.Int("duration_seconds", durationSeconds),
		attribute.Int64("size_bytes", sizeBytes),
	)

	// Check video limits
	if durationSeconds > MaxVideoDuration {
		logger.Warn("video exceeds duration limit, skipping",
			"duration", durationSeconds,
			"max_duration", MaxVideoDuration,
			"video_url", videoURL)
		return "", fmt.Errorf("video duration %ds exceeds limit of %ds", durationSeconds, MaxVideoDuration)
	}

	if sizeBytes > MaxVideoSize {
		logger.Warn("video exceeds size limit, skipping",
			"size_mb", sizeBytes/(1024*1024),
			"max_size_mb", MaxVideoSize/(1024*1024),
			"video_url", videoURL)
		return "", fmt.Errorf("video size %d MB exceeds limit of %d MB", sizeBytes/(1024*1024), MaxVideoSize/(1024*1024))
	}

	// TODO: Implement video transcription
	// Options:
	// 1. Use OpenAI Whisper API directly (not through OpenRouter)
	// 2. Use another transcription service
	// 3. Extract frames and describe them as images

	logger.Info("video transcription not yet implemented, skipping",
		"video_url", videoURL)

	return "", fmt.Errorf("video transcription not yet implemented")
}

// makeCompletionRequest makes a chat completion request to OpenRouter
func (c *OpenRouterClient) makeCompletionRequest(ctx context.Context, req ChatCompletionRequest) (string, error) {
	ctx, span := openRouterTracer.Start(ctx, "makeCompletionRequest")
	defer span.End()

	span.SetAttributes(attribute.String("model", req.Model))

	// Marshal request body
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("HTTP-Referer", "https://askyourfeed.com") // Optional: for OpenRouter analytics
	httpReq.Header.Set("X-Title", "AskYourFeed")                  // Optional: for OpenRouter analytics

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		span.SetAttributes(attribute.Int("status_code", resp.StatusCode))
		return "", fmt.Errorf("API error: %d, body: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var completionResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API error
	if completionResp.Error != nil {
		return "", fmt.Errorf("OpenRouter API error: %s (type: %s, code: %s)",
			completionResp.Error.Message,
			completionResp.Error.Type,
			completionResp.Error.Code)
	}

	// Extract content from first choice
	if len(completionResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := completionResp.Choices[0].Message.Content
	span.SetAttributes(attribute.Int("content_length", len(content)))

	return content, nil
}
