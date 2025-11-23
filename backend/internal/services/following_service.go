package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sopeal/AskYourFeed/internal/dto"
	"github.com/sopeal/AskYourFeed/internal/repositories"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var followingServiceTracer = otel.Tracer("following_service")

// FollowingService handles business logic for following-related operations
type FollowingService struct {
	followingRepo *repositories.FollowingRepository
}

// NewFollowingService creates a new FollowingService instance
func NewFollowingService(followingRepo *repositories.FollowingRepository) *FollowingService {
	return &FollowingService{
		followingRepo: followingRepo,
	}
}

// GetFollowing retrieves paginated list of authors the user follows
func (s *FollowingService) GetFollowing(ctx context.Context, userID uuid.UUID) (*dto.FollowingListResponseDTO, error) {
	ctx, span := followingServiceTracer.Start(ctx, "GetFollowing")
	defer span.End()

	span.SetAttributes(
		attribute.String("user_id", userID.String()),
	)

	// Fetch one extra item to determine if there are more pages
	items, err := s.followingRepo.GetFollowing(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get following list: %w", err)
	}

	// Convert repository items to DTOs
	dtoItems := make([]dto.FollowingItemDTO, len(items))
	for i, item := range items {
		dtoItems[i] = dto.FollowingItemDTO{
			XAuthorID:     item.XAuthorID,
			Handle:        item.Handle,
			DisplayName:   convertStringPtr(item.DisplayName),
			LastSeenAt:    item.LastSeenAt,
			LastCheckedAt: item.LastCheckedAt,
		}
	}

	// Prepare response
	response := &dto.FollowingListResponseDTO{
		Items: dtoItems,
	}

	span.SetAttributes(
		attribute.Int("items_returned", len(dtoItems)),
	)

	return response, nil
}

// convertStringPtr converts *string to string, returning empty string if nil
func convertStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
