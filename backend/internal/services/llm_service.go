package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sopeal/AskYourFeed/internal/db"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var llmTracer = otel.Tracer("llm_service")

// Common errors
var (
	ErrLLMUnavailable    = errors.New("LLM service is unavailable")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

// LLMService handles interactions with Language Model APIs
type LLMService struct {
	openRouterClient *OpenRouterClient
	model            string
}

// NewLLMService creates a new LLMService instance
func NewLLMService(openRouterClient *OpenRouterClient) *LLMService {
	return &LLMService{
		openRouterClient: openRouterClient,
		model:            "google/gemini-2.5-flash",
	}
}

// GenerateAnswer generates an answer to a question based on provided posts
// Returns the generated answer and selected source posts
func (s *LLMService) GenerateAnswer(ctx context.Context, question string, posts []db.PostWithAuthor) (string, []int64, error) {
	ctx, span := llmTracer.Start(ctx, "GenerateAnswer")
	defer span.End()

	span.SetAttributes(
		attribute.Int("post_count", len(posts)),
		attribute.Int("question_length", len(question)),
	)

	// Set timeout for LLM API call - increased to 90 seconds for complex queries
	ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// Format posts for LLM context
	formattedPosts := s.formatPostsForLLM(posts)

	// Construct prompt
	systemPrompt := s.buildSystemPrompt()
	userPrompt := s.buildUserPrompt(question, formattedPosts)

	span.SetAttributes(
		attribute.Int("formatted_posts_length", len(formattedPosts)),
		attribute.Int("user_prompt_length", len(userPrompt)),
	)

	// For now, we'll simulate the response
	answer, sourcePostIDs, err := s.callLLMAPI(ctx, systemPrompt, userPrompt, posts)
	if err != nil {
		span.RecordError(err)
		return "", nil, err
	}

	span.SetAttributes(attribute.Int("source_count", len(sourcePostIDs)))

	return answer, sourcePostIDs, nil
}

// formatPostsForLLM formats posts chronologically with metadata
func (s *LLMService) formatPostsForLLM(posts []db.PostWithAuthor) string {
	if len(posts) == 0 {
		return ""
	}

	var formatted string
	for i, post := range posts {
		displayName := post.Handle
		if post.DisplayName != nil && *post.DisplayName != "" {
			displayName = *post.DisplayName
		}

		formatted += fmt.Sprintf(
			"[Post %d]\nAuthor: %s (@%s)\nPublished: %s\nURL: %s\nContent: %s\n\n",
			i+1,
			displayName,
			post.Handle,
			post.PublishedAt.Format(time.RFC3339),
			post.URL,
			post.Text,
		)
	}

	return formatted
}

// buildSystemPrompt constructs the system prompt for the LLM
func (s *LLMService) buildSystemPrompt() string {
	return `You are an AI assistant that analyzes social media feed posts. Your role is to answer user questions based ONLY on the feed posts provided below.

Important constraints:
- Do not browse the web or use external knowledge
- Only reference information from the provided posts
- Generate structured answers with bullet points when appropriate
- Be concise and factual
- If the posts don't contain relevant information, state this clearly
- Always cite which posts you're referencing in your answer
- Answer in the same language as the user question.`
}

// buildUserPrompt constructs the user prompt with question and posts
func (s *LLMService) buildUserPrompt(question string, formattedPosts string) string {
	return fmt.Sprintf(`Here are the user's feed posts:

%s

User's question: %s

Please answer the question based on the posts above. Structure your answer with bullet points if there are multiple topics. Be specific and cite relevant posts.`, formattedPosts, question)
}

// callLLMAPI calls OpenRouter API to generate an answer
func (s *LLMService) callLLMAPI(ctx context.Context, systemPrompt, userPrompt string, posts []db.PostWithAuthor) (string, []int64, error) {
	if s.openRouterClient == nil {
		return "", nil, ErrLLMUnavailable
	}

	// Prepare the request with system and user messages using SDK types
	req := openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
	}

	// Make the API call
	answer, err := s.openRouterClient.makeCompletionRequest(ctx, req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to call OpenRouter API: %w", err)
	}

	return answer, []int64{}, nil
}
