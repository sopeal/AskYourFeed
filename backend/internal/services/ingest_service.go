package services

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"github.com/sopeal/AskYourFeed/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ingestionServiceTracer = otel.Tracer("ingestion_service")

const (
	// MaxFollowingLimit is the maximum number of followed users to sync (MVP limit)
	MaxFollowingLimit = 150
	
	// MaxRetries is the maximum number of retries for rate-limited requests
	MaxRetries = 3
	
	// BaseBackoffDelay is the base delay for exponential backoff (in seconds)
	BaseBackoffDelay = 2
)

// IngestService handles the actual ingestion of Twitter data
type IngestService struct {
	twitterClient    *TwitterClient
	openRouterClient *OpenRouterClient
	ingestRepo       *repositories.IngestRepository
	followingRepo    *repositories.FollowingRepository
	postRepo         *repositories.PostRepository
	authorRepo       *repositories.AuthorRepository
	userRepo         repositories.UserRepository
}

// NewIngestService creates a new IngestService instance
func NewIngestService(
	twitterClient *TwitterClient,
	openRouterClient *OpenRouterClient,
	ingestRepo *repositories.IngestRepository,
	followingRepo *repositories.FollowingRepository,
	postRepo *repositories.PostRepository,
	authorRepo *repositories.AuthorRepository,
	userRepo repositories.UserRepository,
) *IngestService {
	return &IngestService{
		twitterClient:    twitterClient,
		openRouterClient: openRouterClient,
		ingestRepo:       ingestRepo,
		followingRepo:    followingRepo,
		postRepo:         postRepo,
		authorRepo:       authorRepo,
		userRepo:         userRepo,
	}
}

// IngestUserData performs a complete ingestion for a user with backfill support
func (s *IngestService) IngestUserData(ctx context.Context, userID uuid.UUID, backfillHours int) error {
	ctx, span := ingestionServiceTracer.Start(ctx, "IngestUserData")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.Int("backfill_hours", backfillHours),
	)

	// Check if there's already a running ingest
	currentRun, err := s.ingestRepo.GetCurrentRun(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check current run: %w", err)
	}

	if currentRun != nil {
		return fmt.Errorf("ingestion already running for user %s", userID.String())
	}

	// Get user's X username
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", userID.String())
	}

	// Create a new ingest run
	runID := ulid.Make().String()
	sinceID := int64(1000000000) // Default starting point
	err = s.ingestRepo.CreateIngestRun(ctx, userID, runID, sinceID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create ingest run: %w", err)
	}

	span.SetAttributes(attribute.String("run_id", runID))

	// Track rate limiting metrics
	rateLimitHits := 0
	retried := 0

	// Perform the ingestion
	totalFetched, rateLimitHits, retried, err := s.performIngestion(ctx, userID, user.XUsername, runID, backfillHours)
	if err != nil {
		// Mark run as failed
		errText := err.Error()
		
		// Determine status based on error type
		status := "error"
		if strings.Contains(errText, "429") || strings.Contains(errText, "rate limit") {
			status = "rate_limited"
		}
		
		s.ingestRepo.CompleteIngestRun(ctx, runID, status, totalFetched, &errText)
		span.RecordError(err)
		return fmt.Errorf("ingestion failed: %w", err)
	}

	// Update run with rate limit metrics before completing
	if rateLimitHits > 0 || retried > 0 {
		// Note: We'd need to add a method to update these metrics
		logger.Info("ingestion completed with rate limiting",
			"run_id", runID,
			"rate_limit_hits", rateLimitHits,
			"retried", retried)
	}

	// Mark run as completed
	err = s.ingestRepo.CompleteIngestRun(ctx, runID, "ok", totalFetched, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to complete ingest run: %w", err)
	}

	span.SetAttributes(
		attribute.Int("total_fetched", totalFetched),
		attribute.Int("rate_limit_hits", rateLimitHits),
		attribute.Int("retried", retried),
	)
	
	logger.Info("ingestion completed successfully",
		"user_id", userID,
		"run_id", runID,
		"total_fetched", totalFetched)
	
	return nil
}

// performIngestion executes the actual ingestion logic
func (s *IngestService) performIngestion(ctx context.Context, userID uuid.UUID, xUsername string, runID string, backfillHours int) (int, int, int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "performIngestion")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("x_username", xUsername),
		attribute.String("run_id", runID),
		attribute.Int("backfill_hours", backfillHours),
	)

	totalFetched := 0
	totalRateLimitHits := 0
	totalRetried := 0

	// Step 1: Update following list (max 150 users)
	followingFetched, rateLimitHits, retried, err := s.ingestFollowing(ctx, userID, xUsername, runID)
	if err != nil {
		span.RecordError(err)
		return totalFetched, totalRateLimitHits, totalRetried, fmt.Errorf("failed to ingest following: %w", err)
	}
	totalFetched += followingFetched
	totalRateLimitHits += rateLimitHits
	totalRetried += retried

	// Step 2: Ingest tweets from followed users
	tweetsFetched, rateLimitHits, retried, err := s.ingestTweets(ctx, userID, runID, backfillHours)
	if err != nil {
		span.RecordError(err)
		return totalFetched, totalRateLimitHits, totalRetried, fmt.Errorf("failed to ingest tweets: %w", err)
	}
	totalFetched += tweetsFetched
	totalRateLimitHits += rateLimitHits
	totalRetried += retried

	span.SetAttributes(
		attribute.Int("total_fetched", totalFetched),
		attribute.Int("total_rate_limit_hits", totalRateLimitHits),
		attribute.Int("total_retried", totalRetried),
	)
	
	return totalFetched, totalRateLimitHits, totalRetried, nil
}

// ingestFollowing updates the user's following list (max 150 users)
func (s *IngestService) ingestFollowing(ctx context.Context, userID uuid.UUID, xUsername string, runID string) (int, int, int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("run_id", runID),
		attribute.String("x_username", xUsername),
	)

	cursor := ""
	fetched := 0
	rateLimitHits := 0
	retried := 0

	for {
		// Check if we've reached the limit
		if fetched >= MaxFollowingLimit {
			logger.Info("reached following limit",
				"user_id", userID,
				"limit", MaxFollowingLimit,
				"fetched", fetched)
			break
		}

		// Get following from Twitter API with retry logic
		resp, hits, retries, err := s.getFollowingsWithRetry(ctx, xUsername, cursor)
		if err != nil {
			span.RecordError(err)
			return fetched, rateLimitHits, retried, fmt.Errorf("failed to get user followings: %w", err)
		}
		rateLimitHits += hits
		retried += retries

		// Process each following
		for _, user := range resp.Users {
			if fetched >= MaxFollowingLimit {
				break
			}

			authorID, err := s.ensureAuthorExists(ctx, &user)
			if err != nil {
				span.RecordError(err)
				logger.Warn("failed to ensure author exists, skipping",
					"error", err,
					"handle", user.UserName)
				continue
			}

			// Update or insert following relationship with last_checked_at
			err = s.followingRepo.UpsertFollowing(ctx, userID, authorID, time.Now())
			if err != nil {
				span.RecordError(err)
				logger.Warn("failed to upsert following, skipping",
					"error", err,
					"author_id", authorID)
				continue
			}

			fetched++
		}

		// Check if we have more pages and haven't reached limit
		if !resp.HasNextPage || fetched >= MaxFollowingLimit {
			break
		}

		cursor = resp.NextCursor

		// Add a small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	span.SetAttributes(
		attribute.Int("following_fetched", fetched),
		attribute.Int("rate_limit_hits", rateLimitHits),
		attribute.Int("retried", retried),
	)
	
	logger.Info("following ingestion completed",
		"user_id", userID,
		"fetched", fetched,
		"rate_limit_hits", rateLimitHits)
	
	return fetched, rateLimitHits, retried, nil
}

// ingestTweets ingests tweets from followed users
func (s *IngestService) ingestTweets(ctx context.Context, userID uuid.UUID, runID string, backfillHours int) (int, int, int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestTweets")
	defer span.End()

	span.SetAttributes(
		attribute.String("run_id", runID),
		attribute.Int("backfill_hours", backfillHours),
	)

	// Get all followed authors
	following, err := s.followingRepo.GetFollowing(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return 0, 0, 0, fmt.Errorf("failed to get following list: %w", err)
	}

	fetched := 0
	rateLimitHits := 0
	retried := 0
	
	// Calculate backfill cutoff time
	backfillCutoff := time.Now().Add(-time.Duration(backfillHours) * time.Hour)
	isBackfill := backfillHours > 0

	for _, follow := range following {
		// Get author details
		author, err := s.authorRepo.GetAuthor(ctx, follow.XAuthorID)
		if err != nil {
			span.RecordError(err)
			logger.Warn("failed to get author details, skipping",
				"error", err,
				"author_id", follow.XAuthorID,
				"user_id", userID)
			continue
		}

		if author == nil || author.Handle == "" {
			logger.Debug("skipping author with no handle",
				"author_id", follow.XAuthorID,
				"user_id", userID)
			continue
		}

		// Get tweets for this author
		authorTweetsFetched, hits, retries, err := s.ingestTweetsForAuthor(
			ctx, userID, author.Handle, author.XAuthorID, backfillCutoff, isBackfill)
		if err != nil {
			span.RecordError(err)
			logger.Warn("failed to ingest tweets for author, continuing with others",
				"error", err,
				"author_handle", author.Handle,
				"author_id", follow.XAuthorID,
				"user_id", userID)
			continue
		}

		fetched += authorTweetsFetched
		rateLimitHits += hits
		retried += retries

		// Update progress
		err = s.ingestRepo.UpdateIngestRunProgress(ctx, runID, fetched)
		if err != nil {
			span.RecordError(err)
			logger.Warn("failed to update progress",
				"error", err,
				"run_id", runID)
		}

		// Add delay between authors to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	span.SetAttributes(
		attribute.Int("tweets_fetched", fetched),
		attribute.Int("rate_limit_hits", rateLimitHits),
		attribute.Int("retried", retried),
	)
	
	logger.Info("tweets ingestion completed",
		"user_id", userID,
		"fetched", fetched,
		"authors_processed", len(following),
		"rate_limit_hits", rateLimitHits)
	
	return fetched, rateLimitHits, retried, nil
}

// ingestTweetsForAuthor ingests tweets for a specific author with pagination and temporal filtering
func (s *IngestService) ingestTweetsForAuthor(
	ctx context.Context,
	userID uuid.UUID,
	authorHandle string,
	authorID int64,
	backfillCutoff time.Time,
	isBackfill bool,
) (int, int, int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestTweetsForAuthor")
	defer span.End()

	span.SetAttributes(
		attribute.String("author_handle", authorHandle),
		attribute.Int64("author_id", authorID),
		attribute.Bool("is_backfill", isBackfill),
	)

	fetched := 0
	rateLimitHits := 0
	retried := 0
	cursor := ""
	shouldContinue := true
	latestSeenAt := time.Time{}

	for shouldContinue {
		// Get tweets from Twitter API with retry logic
		resp, hits, retries, err := s.getTweetsWithRetry(ctx, authorHandle, cursor)
		if err != nil {
			span.RecordError(err)
			return fetched, rateLimitHits, retried, fmt.Errorf("failed to get user tweets: %w", err)
		}
		rateLimitHits += hits
		retried += retries

		// Process each tweet
		tweetsInPage := 0
		for _, tweet := range resp.Tweets {
			// Only ingest original posts (filters out retweets, quote tweets, but allows self-replies)
			if !s.twitterClient.IsOriginalPost(tweet) {
				continue
			}

			// Parse tweet timestamp for temporal filtering
			tweetTime, err := time.Parse(time.RFC3339, tweet.CreatedAt)
			if err != nil {
				logger.Warn("failed to parse tweet timestamp, skipping",
					"error", err,
					"tweet_id", tweet.ID,
					"created_at", tweet.CreatedAt)
				continue
			}

			// Temporal filtering: skip tweets older than backfill cutoff
			if isBackfill && tweetTime.Before(backfillCutoff) {
				logger.Debug("reached backfill cutoff, stopping pagination",
					"author_handle", authorHandle,
					"tweet_time", tweetTime,
					"cutoff", backfillCutoff)
				shouldContinue = false
				break
			}

			// Track latest seen timestamp for author update
			if latestSeenAt.IsZero() || tweetTime.After(latestSeenAt) {
				latestSeenAt = tweetTime
			}

			// Convert to DTO
			tweetDTO := s.twitterClient.ConvertToDTO(tweet)

			// Process media (images and videos) if OpenRouter client is available
			if s.openRouterClient != nil {
				err := s.processMedia(ctx, &tweet, tweetDTO)
				if err != nil {
					logger.Warn("failed to process media, continuing without media descriptions",
						"error", err,
						"tweet_id", tweet.ID,
						"author_handle", authorHandle)
					// Continue processing the tweet without media descriptions
				}
			}

			// Check if tweet already exists
			exists, err := s.postRepo.PostExists(ctx, userID, tweetDTO.ID)
			if err != nil {
				logger.Warn("failed to check if post exists, skipping",
					"error", err,
					"post_id", tweetDTO.ID,
					"author_handle", authorHandle)
				continue
			}

			if exists {
				continue // Skip existing tweets
			}

			// Insert the tweet
			err = s.postRepo.InsertPost(ctx, userID, tweetDTO)
			if err != nil {
				logger.Error("failed to insert post",
					err,
					"post_id", tweetDTO.ID,
					"author_handle", authorHandle,
					"user_id", userID)
				continue
			}

			fetched++
			tweetsInPage++
		}

		// For regular ingest (not backfill), only fetch first page
		if !isBackfill {
			logger.Debug("regular ingest: stopping after first page",
				"author_handle", authorHandle,
				"tweets_fetched", tweetsInPage)
			break
		}

		// Check if we have more pages
		if !resp.HasNextPage {
			shouldContinue = false
			break
		}

		cursor = resp.NextCursor

		// Add delay between pages
		time.Sleep(100 * time.Millisecond)
	}

	// Update author's last_seen_at if we found any tweets
	if !latestSeenAt.IsZero() {
		err := s.authorRepo.UpdateAuthorLastSeen(ctx, authorID, latestSeenAt)
		if err != nil {
			logger.Warn("failed to update author last_seen_at",
				"error", err,
				"author_id", authorID)
		}
	}

	span.SetAttributes(
		attribute.Int("tweets_fetched", fetched),
		attribute.Int("rate_limit_hits", rateLimitHits),
		attribute.Int("retried", retried),
	)

	return fetched, rateLimitHits, retried, nil
}

// ensureAuthorExists ensures an author exists in the database
func (s *IngestService) ensureAuthorExists(ctx context.Context, user *UserData) (int64, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ensureAuthorExists")
	defer span.End()

	span.SetAttributes(attribute.String("user_handle", user.UserName))

	// Check if author already exists
	existing, err := s.authorRepo.GetAuthorByHandle(ctx, user.UserName)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to check existing author: %w", err)
	}

	if existing != nil {
		return existing.XAuthorID, nil
	}

	// Convert to DTO and insert
	userDTO := s.twitterClient.ConvertUserToDTO(*user)
	authorID, err := s.authorRepo.InsertAuthor(ctx, userDTO)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to insert author: %w", err)
	}

	return authorID, nil
}

// getFollowingsWithRetry gets user followings with exponential backoff retry logic
func (s *IngestService) getFollowingsWithRetry(ctx context.Context, username string, cursor string) (*FollowingResponse, int, int, error) {
	rateLimitHits := 0
	retried := 0

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		resp, err := s.twitterClient.GetUserFollowings(ctx, username, cursor)
		if err == nil {
			return resp, rateLimitHits, retried, nil
		}

		// Check if it's a rate limit error (429)
		if strings.Contains(err.Error(), "429") {
			rateLimitHits++
			
			if attempt < MaxRetries {
				retried++
				backoffDelay := time.Duration(math.Pow(float64(BaseBackoffDelay), float64(attempt+1))) * time.Second
				logger.Warn("rate limited, retrying with exponential backoff",
					"attempt", attempt+1,
					"max_retries", MaxRetries,
					"backoff_delay", backoffDelay,
					"username", username)
				
				time.Sleep(backoffDelay)
				continue
			}
		}

		// Non-rate-limit error or max retries exceeded
		return nil, rateLimitHits, retried, err
	}

	return nil, rateLimitHits, retried, fmt.Errorf("max retries exceeded for user followings")
}

// getTweetsWithRetry gets user tweets with exponential backoff retry logic
func (s *IngestService) getTweetsWithRetry(ctx context.Context, username string, cursor string) (*TweetResponse, int, int, error) {
	rateLimitHits := 0
	retried := 0

	for attempt := 0; attempt <= MaxRetries; attempt++ {
		resp, err := s.twitterClient.GetUserTweets(ctx, username, cursor)
		if err == nil {
			return resp, rateLimitHits, retried, nil
		}

		// Check if it's a rate limit error (429)
		if strings.Contains(err.Error(), "429") {
			rateLimitHits++
			
			if attempt < MaxRetries {
				retried++
				backoffDelay := time.Duration(math.Pow(float64(BaseBackoffDelay), float64(attempt+1))) * time.Second
				logger.Warn("rate limited, retrying with exponential backoff",
					"attempt", attempt+1,
					"max_retries", MaxRetries,
					"backoff_delay", backoffDelay,
					"username", username)
				
				time.Sleep(backoffDelay)
				continue
			}
		}

		// Non-rate-limit error or max retries exceeded
		return nil, rateLimitHits, retried, err
	}

	return nil, rateLimitHits, retried, fmt.Errorf("max retries exceeded for user tweets")
}

// processMedia processes media (images and videos) in a tweet and appends descriptions to the text
func (s *IngestService) processMedia(ctx context.Context, tweet *TweetData, tweetDTO *dto.TweetDTO) error {
	ctx, span := ingestionServiceTracer.Start(ctx, "processMedia")
	defer span.End()

	if tweet.Media == nil {
		return nil // No media to process
	}

	var mediaDescriptions []string

	// Process images (max 4)
	if len(tweet.Media.Photos) > 0 {
		span.SetAttributes(attribute.Int("photo_count", len(tweet.Media.Photos)))
		
		// Collect image URLs (limit to MaxImagesPerPost)
		imageURLs := make([]string, 0, len(tweet.Media.Photos))
		for i, photo := range tweet.Media.Photos {
			if i >= MaxImagesPerPost {
				logger.Warn("tweet has more than max images, processing only first",
					"total_images", len(tweet.Media.Photos),
					"max_images", MaxImagesPerPost,
					"tweet_id", tweet.ID)
				break
			}
			imageURLs = append(imageURLs, photo.URL)
		}

		// Describe images using OpenRouter
		if len(imageURLs) > 0 {
			descriptions, err := s.openRouterClient.DescribeImages(ctx, imageURLs)
			if err != nil {
				span.RecordError(err)
				logger.Warn("failed to describe images",
					"error", err,
					"tweet_id", tweet.ID,
					"image_count", len(imageURLs))
				// Don't return error, continue without image descriptions
			} else {
				for i, desc := range descriptions {
					mediaDescriptions = append(mediaDescriptions, fmt.Sprintf("[Image %d: %s]", i+1, desc))
				}
				span.SetAttributes(attribute.Int("images_described", len(descriptions)))
			}
		}
	}

	// Process videos
	if len(tweet.Media.Videos) > 0 {
		span.SetAttributes(attribute.Int("video_count", len(tweet.Media.Videos)))
		
		for i, video := range tweet.Media.Videos {
			// Calculate duration in seconds
			durationSeconds := video.DurationMs / 1000
			
			// Estimate size (we don't have exact size from API, so we skip size check)
			// In a real implementation, you might need to fetch the video to check size
			sizeBytes := int64(0) // Unknown size
			
			// Try to transcribe video
			transcription, err := s.openRouterClient.TranscribeVideo(ctx, video.URL, durationSeconds, sizeBytes)
			if err != nil {
				// Check if it's a limit error
				if strings.Contains(err.Error(), "exceeds limit") {
					logger.Warn("video exceeds limits, skipping",
						"error", err,
						"tweet_id", tweet.ID,
						"video_index", i,
						"duration_seconds", durationSeconds)
					// Continue to next video
					continue
				} else if strings.Contains(err.Error(), "not yet implemented") {
					logger.Debug("video transcription not implemented, skipping",
						"tweet_id", tweet.ID,
						"video_index", i)
					// Continue to next video
					continue
				} else {
					logger.Warn("failed to transcribe video",
						"error", err,
						"tweet_id", tweet.ID,
						"video_index", i)
					// Continue to next video
					continue
				}
			}
			
			if transcription != "" {
				mediaDescriptions = append(mediaDescriptions, fmt.Sprintf("[Video %d transcription: %s]", i+1, transcription))
				span.SetAttributes(attribute.Int("videos_transcribed", i+1))
			}
		}
	}

	// Append media descriptions to tweet text
	if len(mediaDescriptions) > 0 {
		originalText := tweetDTO.Text
		mediaText := strings.Join(mediaDescriptions, " ")
		tweetDTO.Text = originalText + "\n\n" + mediaText
		
		span.SetAttributes(
			attribute.Int("media_descriptions_count", len(mediaDescriptions)),
			attribute.Int("original_text_length", len(originalText)),
			attribute.Int("final_text_length", len(tweetDTO.Text)),
		)
		
		logger.Info("media processed and added to tweet text",
			"tweet_id", tweet.ID,
			"media_count", len(mediaDescriptions),
			"original_length", len(originalText),
			"final_length", len(tweetDTO.Text))
	}

	return nil
}
