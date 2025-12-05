package services

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	// Configuration for LLM API (e.g., API key, model, endpoint)
	// In a real implementation, these would be injected
}

// NewLLMService creates a new LLMService instance
func NewLLMService() *LLMService {
	return &LLMService{}
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

	// Set timeout for LLM API call
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
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

	// TODO: In production, this would call actual LLM API (OpenAI, Anthropic, etc.)
	// For now, we'll simulate the response
	answer, sourcePostIDs, err := s.callLLMAPI(ctx, systemPrompt, userPrompt, posts)
	if err != nil {
		span.RecordError(err)
		return "", nil, err
	}

	// Ensure minimum 3 sources if available
	sourcePostIDs = s.ensureMinimumSources(sourcePostIDs, posts, 3)

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
- Always cite which posts you're referencing in your answer`
}

// buildUserPrompt constructs the user prompt with question and posts
func (s *LLMService) buildUserPrompt(question string, formattedPosts string) string {
	return fmt.Sprintf(`Here are the user's feed posts:

%s

User's question: %s

Please answer the question based on the posts above. Structure your answer with bullet points if there are multiple topics. Be specific and cite relevant posts.`, formattedPosts, question)
}

// callLLMAPI simulates calling the actual LLM API
// In production, this would use OpenAI SDK, Anthropic SDK, or HTTP client
func (s *LLMService) callLLMAPI(ctx context.Context, systemPrompt, userPrompt string, posts []db.PostWithAuthor) (string, []int64, error) {
	// TODO: Replace with actual LLM API call
	// Example for OpenAI:
	// client := openai.NewClient(apiKey)
	// resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
	//     Model: "gpt-4-turbo",
	//     Messages: []openai.ChatCompletionMessage{
	//         {Role: "system", Content: systemPrompt},
	//         {Role: "user", Content: userPrompt},
	//     },
	// })

	// For now, simulate a response
	answer := `Na podstawie analizy postów z wybranego okresu, oto główne tematy dyskusji:

• Rozwój sztucznej inteligencji i jej zastosowania w różnych dziedzinach
• Nowe regulacje dotyczące prywatności danych
• Innowacje technologiczne w sektorze finansowym
• Dyskusje na temat zrównoważonego rozwoju i zmian klimatycznych

Źródła tych informacji pochodzą z kilku najnowszych postów w Twoim feedzie.`

	// Select diverse source posts (first, middle, last few posts as sources)
	var sourceIDs []int64
	if len(posts) > 0 {
		// Select posts strategically to provide good coverage
		indices := s.selectSourceIndices(len(posts))
		for _, idx := range indices {
			sourceIDs = append(sourceIDs, posts[idx].XPostID)
		}
	}

	return answer, sourceIDs, nil
}

// selectSourceIndices selects indices of posts to use as sources
// Aims for diverse temporal coverage across the post collection
func (s *LLMService) selectSourceIndices(postCount int) []int {
	if postCount == 0 {
		return []int{}
	}

	if postCount <= 3 {
		// Return all posts if we have 3 or fewer
		indices := make([]int, postCount)
		for i := 0; i < postCount; i++ {
			indices[i] = i
		}
		return indices
	}

	// For more than 3 posts, select:
	// - First post (earliest)
	// - Middle post(s)
	// - Last post (most recent)
	indices := []int{0} // First post

	// Middle post(s)
	mid := postCount / 2
	indices = append(indices, mid)

	// If we have many posts, add another middle point
	if postCount > 6 {
		quarterMark := postCount / 4
		threeQuarterMark := (postCount * 3) / 4
		indices = append(indices, quarterMark, threeQuarterMark)
	}

	// Last post
	indices = append(indices, postCount-1)

	// Remove duplicates and sort
	seen := make(map[int]bool)
	uniqueIndices := []int{}
	for _, idx := range indices {
		if !seen[idx] {
			seen[idx] = true
			uniqueIndices = append(uniqueIndices, idx)
		}
	}

	return uniqueIndices
}

// ensureMinimumSources ensures at least minCount sources are returned if available
func (s *LLMService) ensureMinimumSources(sourceIDs []int64, posts []db.PostWithAuthor, minCount int) []int64 {
	if len(sourceIDs) >= minCount || len(posts) <= minCount {
		return sourceIDs
	}

	// If we don't have enough sources but have enough posts, add more
	existingIDs := make(map[int64]bool)
	for _, id := range sourceIDs {
		existingIDs[id] = true
	}

	for _, post := range posts {
		if !existingIDs[post.XPostID] {
			sourceIDs = append(sourceIDs, post.XPostID)
			existingIDs[post.XPostID] = true

			if len(sourceIDs) >= minCount {
				break
			}
		}
	}

	return sourceIDs
}
