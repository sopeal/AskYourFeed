package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"
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
	client *openai.Client
}

// NewOpenRouterClient creates a new OpenRouter API client
func NewOpenRouterClient(apiKey string, httpClient *http.Client) *OpenRouterClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for LLM responses and vision/transcription
		}
	}

	// Create OpenAI client configuration for OpenRouter
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = OpenRouterBaseURL
	config.HTTPClient = httpClient

	client := openai.NewClientWithConfig(config)

	return &OpenRouterClient{
		client: client,
	}
}

// DescribeImage generates a text description of an image using vision model
func (c *OpenRouterClient) DescribeImage(ctx context.Context, imageURL string) (string, error) {
	ctx, span := openRouterTracer.Start(ctx, "DescribeImage")
	defer span.End()

	span.SetAttributes(attribute.String("image_url", imageURL))

	// Prepare the request using SDK
	req := openai.ChatCompletionRequest{
		Model: "openai/gpt-4o-mini", // Using GPT-4o-mini for vision (cost-effective)
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "Describe this image in detail. Focus on the main content, text, and any important visual elements. Keep it concise but informative.",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
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

// makeCompletionRequest makes a chat completion request to OpenRouter using the SDK
func (c *OpenRouterClient) makeCompletionRequest(ctx context.Context, req openai.ChatCompletionRequest) (string, error) {
	ctx, span := openRouterTracer.Start(ctx, "makeCompletionRequest")
	defer span.End()

	span.SetAttributes(attribute.String("model", req.Model))

	// Make request using SDK
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Extract content from first choice
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := resp.Choices[0].Message.Content
	span.SetAttributes(attribute.Int("content_length", len(content)))

	return content, nil
}
