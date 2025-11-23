package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ingestionServiceTracer = otel.Tracer("ingestion_service")

// IngestService handles the actual ingestion of Twitter data
type IngestService struct {
	twitterClient *TwitterClient
	ingestRepo    *repositories.IngestRepository
	followingRepo *repositories.FollowingRepository
	postRepo      *repositories.PostRepository
	authorRepo    *repositories.AuthorRepository
}

// NewIngestionService creates a new IngestService instance
func NewIngestionService(
	twitterClient *TwitterClient,
	ingestRepo *repositories.IngestRepository,
	followingRepo *repositories.FollowingRepository,
	postRepo *repositories.PostRepository,
	authorRepo *repositories.AuthorRepository,
) *IngestService {
	return &IngestService{
		twitterClient: twitterClient,
		ingestRepo:    ingestRepo,
		followingRepo: followingRepo,
		postRepo:      postRepo,
		authorRepo:    authorRepo,
	}
}

// IngestUserData performs a complete ingestion for a user
func (s *IngestService) IngestUserData(ctx context.Context, userID uuid.UUID) error {
	ctx, span := ingestionServiceTracer.Start(ctx, "IngestUserData")
	defer span.End()

	span.SetAttributes(attribute.String("user_id", userID.String()))

	// Check if there's already a running ingest
	currentRun, err := s.ingestRepo.GetCurrentRun(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to check current run: %w", err)
	}

	if currentRun != nil {
		return fmt.Errorf("ingestion already running for user %s", userID.String())
	}

	// Get the last cursor from the most recent completed run
	lastCursor := ""
	lastRun, err := s.ingestRepo.GetRecentRuns(ctx, userID, 1)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to get last run: %w", err)
	}

	if len(lastRun) > 0 {
		lastCursor = lastRun[0].Cursor
	}

	// Create a new ingest run
	runID := ulid.Make().String()
	err = s.ingestRepo.CreateIngestRun(ctx, userID, runID, lastCursor)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create ingest run: %w", err)
	}

	span.SetAttributes(attribute.String("run_id", runID))

	// Perform the ingestion
	totalFetched, err := s.performIngestion(ctx, userID, runID, lastCursor)
	if err != nil {
		// Mark run as failed
		errText := err.Error()
		s.ingestRepo.CompleteIngestRun(ctx, runID, "error", totalFetched, &errText)
		span.RecordError(err)
		return fmt.Errorf("ingestion failed: %w", err)
	}

	// Mark run as completed
	err = s.ingestRepo.CompleteIngestRun(ctx, runID, "ok", totalFetched, nil)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to complete ingest run: %w", err)
	}

	span.SetAttributes(attribute.Int("total_fetched", totalFetched))
	return nil
}

// performIngestion executes the actual ingestion logic
func (s *IngestService) performIngestion(ctx context.Context, userID uuid.UUID, runID string, startCursor string) (int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "performIngestion")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
		attribute.String("run_id", runID),
		attribute.String("start_cursor", startCursor),
	)

	totalFetched := 0

	// Step 1: Update following list
	followingFetched, err := s.ingestFollowing(ctx, userID, runID)
	if err != nil {
		span.RecordError(err)
		return totalFetched, fmt.Errorf("failed to ingest following: %w", err)
	}
	totalFetched += followingFetched

	// Step 2: Ingest tweets from followed users
	tweetsFetched, err := s.ingestTweets(ctx, userID, runID, startCursor)
	if err != nil {
		span.RecordError(err)
		return totalFetched, fmt.Errorf("failed to ingest tweets: %w", err)
	}
	totalFetched += tweetsFetched

	span.SetAttributes(attribute.Int("total_fetched", totalFetched))
	return totalFetched, nil
}

// ingestFollowing updates the user's following list
func (s *IngestService) ingestFollowing(ctx context.Context, userID uuid.UUID, runID string) (int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestFollowing")
	defer span.End()

	span.SetAttributes(attribute.String("run_id", runID))

	// Get user's handle (this would come from session/auth context in real implementation)
	// For now, we'll assume we have a way to get it
	userHandle := "testuser" // TODO: Get from user session

	cursor := ""
	fetched := 0

	for {
		// Get following from Twitter API
		resp, err := s.twitterClient.GetUserFollowings(ctx, userHandle, cursor)
		if err != nil {
			span.RecordError(err)
			return fetched, fmt.Errorf("failed to get user followings: %w", err)
		}

		// Process each following
		for _, user := range resp.Users {
			authorID, err := s.ensureAuthorExists(ctx, &user)
			if err != nil {
				span.RecordError(err)
				return fetched, fmt.Errorf("failed to ensure author exists: %w", err)
			}

			// Update or insert following relationship
			err = s.followingRepo.UpsertFollowing(ctx, userID, authorID, time.Now())
			if err != nil {
				span.RecordError(err)
				return fetched, fmt.Errorf("failed to upsert following: %w", err)
			}

			fetched++
		}

		// Check if we have more pages
		if !resp.HasNextPage {
			break
		}

		cursor = resp.NextCursor

		// Add a small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	span.SetAttributes(attribute.Int("following_fetched", fetched))
	return fetched, nil
}

// ingestTweets ingests tweets from followed users
func (s *IngestService) ingestTweets(ctx context.Context, userID uuid.UUID, runID string, startCursor string) (int, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestTweets")
	defer span.End()

	span.SetAttributes(
		attribute.String("run_id", runID),
		attribute.String("start_cursor", startCursor),
	)

	// Get all followed authors
	following, err := s.followingRepo.GetFollowing(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to get following list: %w", err)
	}

	fetched := 0
	cursor := startCursor

	for _, follow := range following {
		// Get author details
		author, err := s.authorRepo.GetAuthor(ctx, follow.XAuthorID)
		if err != nil {
			span.RecordError(err)
			continue // Skip this author if we can't get details
		}

		if author == nil || author.Handle == "" {
			continue // Skip if no handle
		}

		// Get tweets for this author
		authorTweetsFetched, nextCursor, err := s.ingestTweetsForAuthor(ctx, userID, author.Handle, cursor)
		if err != nil {
			span.RecordError(err)
			// Log error but continue with other authors
			continue
		}

		fetched += authorTweetsFetched

		// Update cursor for next run
		if nextCursor != "" {
			cursor = nextCursor
			err = s.ingestRepo.UpdateIngestRunCursor(ctx, runID, cursor, fetched)
			if err != nil {
				span.RecordError(err)
				return fetched, fmt.Errorf("failed to update cursor: %w", err)
			}
		}

		// Add delay between authors to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	span.SetAttributes(attribute.Int("tweets_fetched", fetched))
	return fetched, nil
}

// ingestTweetsForAuthor ingests tweets for a specific author
func (s *IngestService) ingestTweetsForAuthor(ctx context.Context, userID uuid.UUID, authorHandle string, cursor string) (int, string, error) {
	ctx, span := ingestionServiceTracer.Start(ctx, "ingestTweetsForAuthor")
	defer span.End()

	span.SetAttributes(
		attribute.String("author_handle", authorHandle),
		attribute.String("cursor", cursor),
	)

	fetched := 0
	nextCursor := cursor

	// Get tweets from Twitter API
	resp, err := s.twitterClient.GetUserTweets(ctx, authorHandle, cursor)
	if err != nil {
		span.RecordError(err)
		return fetched, nextCursor, fmt.Errorf("failed to get user tweets: %w", err)
	}

	// Process each tweet
	for _, tweet := range resp.Tweets {
		// Only ingest original posts (not retweets or quotes)
		if !s.twitterClient.IsOriginalPost(tweet) {
			continue
		}

		// Convert to DTO
		tweetDTO := s.twitterClient.ConvertToDTO(tweet)

		// Check if tweet already exists
		exists, err := s.postRepo.PostExists(ctx, userID, tweetDTO.ID)
		if err != nil {
			span.RecordError(err)
			continue // Skip this tweet
		}

		if exists {
			continue // Skip existing tweets
		}

		// Insert the tweet
		err = s.postRepo.InsertPost(ctx, userID, tweetDTO)
		if err != nil {
			span.RecordError(err)
			continue // Skip this tweet
		}

		fetched++
	}

	if resp.HasNextPage {
		nextCursor = resp.NextCursor
	} else {
		nextCursor = ""
	}

	span.SetAttributes(
		attribute.Int("tweets_fetched", fetched),
		attribute.String("next_cursor", nextCursor),
	)

	return fetched, nextCursor, nil
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
