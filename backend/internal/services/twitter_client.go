package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sopeal/AskYourFeed/internal/dto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var twitterClientTracer = otel.Tracer("twitter_client")

// TwitterClient handles communication with twitterapi.io
type TwitterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewTwitterClient creates a new Twitter API client
func NewTwitterClient(apiKey string) *TwitterClient {
	return &TwitterClient{
		apiKey:  apiKey,
		baseURL: "https://api.twitterapi.io",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UserResponse represents the response from user info endpoint
type UserResponse struct {
	Data   UserData `json:"data"`
	Status string   `json:"status"`
	Msg    string   `json:"msg"`
}

// UserData represents user information
type UserData struct {
	Type           string `json:"type"`
	ID             string `json:"id"`
	UserName       string `json:"userName"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	ProfilePicture string `json:"profilePicture"`
	CoverPicture   string `json:"coverPicture"`
	Description    string `json:"description"`
	Location       string `json:"location"`
	IsBlueVerified bool   `json:"isBlueVerified"`
	Followers      int    `json:"followers"`
	Following      int    `json:"following"`
	CanDm          bool   `json:"canDm"`
	CreatedAt      string `json:"createdAt"`
	FavouritesCount int   `json:"favouritesCount"`
	StatusesCount  int    `json:"statusesCount"`
}

// FollowingResponse represents the response from user followings endpoint
type FollowingResponse struct {
	Users       []UserData `json:"users"`
	HasNextPage bool       `json:"has_next_page"`
	NextCursor  string     `json:"next_cursor"`
	Status      string     `json:"status"`
}

// TweetResponse represents the response from tweet endpoints
type TweetResponse struct {
	Tweets      []TweetData `json:"tweets"`
	HasNextPage bool        `json:"has_next_page"`
	NextCursor  string      `json:"next_cursor"`
	Status      string      `json:"status"`
}

// TweetData represents tweet information
type TweetData struct {
	Type           string     `json:"type"`
	ID             string     `json:"id"`
	URL            string     `json:"url"`
	Text           string     `json:"text"`
	Source         string     `json:"source"`
	RetweetCount   int        `json:"retweetCount"`
	ReplyCount     int        `json:"replyCount"`
	LikeCount      int        `json:"likeCount"`
	QuoteCount     int        `json:"quoteCount"`
	ViewCount      int        `json:"viewCount"`
	BookmarkCount  int        `json:"bookmarkCount"`
	CreatedAt      string     `json:"createdAt"`
	Lang           string     `json:"lang"`
	IsReply        bool       `json:"isReply"`
	InReplyToId    string     `json:"inReplyToId"`
	ConversationId string     `json:"conversationId"`
	Author         UserData   `json:"author"`
	QuotedTweet    *TweetData `json:"quoted_tweet,omitempty"`
	RetweetedTweet *TweetData `json:"retweeted_tweet,omitempty"`
}

// makeRequest performs HTTP request with authentication and error handling
func (c *TwitterClient) makeRequest(ctx context.Context, method, endpoint string, params url.Values) ([]byte, error) {
	ctx, span := twitterClientTracer.Start(ctx, "makeRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
	)

	reqURL := c.baseURL + endpoint
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("User-Agent", "AskYourFeed/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.SetAttributes(attribute.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("API error: %d, body: %s", resp.StatusCode, string(body))
	}

	span.SetAttributes(attribute.Int("response_size", len(body)))
	return body, nil
}

// GetUserInfo retrieves user information by username
func (c *TwitterClient) GetUserInfo(ctx context.Context, username string) (*UserData, error) {
	ctx, span := twitterClientTracer.Start(ctx, "GetUserInfo")
	defer span.End()

	span.SetAttributes(attribute.String("username", username))

	params := url.Values{}
	params.Set("userName", username)

	body, err := c.makeRequest(ctx, "GET", "/twitter/user/info", params)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Status != "success" {
		return nil, fmt.Errorf("API returned error status: %s, msg: %s", resp.Status, resp.Msg)
	}

	return &resp.Data, nil
}

// GetUserFollowings retrieves users that a given user follows
func (c *TwitterClient) GetUserFollowings(ctx context.Context, username string, cursor string) (*FollowingResponse, error) {
	ctx, span := twitterClientTracer.Start(ctx, "GetUserFollowings")
	defer span.End()

	span.SetAttributes(
		attribute.String("username", username),
		attribute.String("cursor", cursor),
	)

	params := url.Values{}
	params.Set("userName", username)
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.makeRequest(ctx, "GET", "/twitter/user/followings", params)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var resp FollowingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	span.SetAttributes(
		attribute.Int("users_count", len(resp.Users)),
		attribute.Bool("has_next_page", resp.HasNextPage),
	)

	return &resp, nil
}

// GetUserTweets retrieves recent tweets from a user
func (c *TwitterClient) GetUserTweets(ctx context.Context, username string, cursor string) (*TweetResponse, error) {
	ctx, span := twitterClientTracer.Start(ctx, "GetUserTweets")
	defer span.End()

	span.SetAttributes(
		attribute.String("username", username),
		attribute.String("cursor", cursor),
	)

	params := url.Values{}
	params.Set("userName", username)
	if cursor != "" {
		params.Set("cursor", cursor)
	}

	body, err := c.makeRequest(ctx, "GET", "/twitter/user/last_tweets", params)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	var resp TweetResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	span.SetAttributes(
		attribute.Int("tweets_count", len(resp.Tweets)),
		attribute.Bool("has_next_page", resp.HasNextPage),
	)

	return &resp, nil
}

// ConvertToDTO converts TweetData to TweetDTO
func (c *TwitterClient) ConvertToDTO(tweet TweetData) *dto.TweetDTO {
	// Parse created_at timestamp
	createdAt, err := time.Parse(time.RFC3339, tweet.CreatedAt)
	if err != nil {
		// Fallback to current time if parsing fails
		createdAt = time.Now()
	}

	// Convert author ID to int64
	authorID, err := strconv.ParseInt(tweet.Author.ID, 10, 64)
	if err != nil {
		authorID = 0
	}

	// Convert tweet ID to int64
	tweetID, err := strconv.ParseInt(tweet.ID, 10, 64)
	if err != nil {
		tweetID = 0
	}

	// Convert conversation ID to int64
	conversationID, err := strconv.ParseInt(tweet.ConversationId, 10, 64)
	if err != nil {
		conversationID = 0
	}

	return &dto.TweetDTO{
		ID:             tweetID,
		AuthorID:       authorID,
		Text:           tweet.Text,
		URL:            tweet.URL,
		PublishedAt:    createdAt,
		ConversationID: conversationID,
	}
}

// ConvertUserToDTO converts UserData to UserDTO
func (c *TwitterClient) ConvertUserToDTO(user UserData) *dto.UserDTO {
	// Parse created_at timestamp
	createdAt, err := time.Parse(time.RFC3339, user.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}

	// Convert user ID to int64
	userID, err := strconv.ParseInt(user.ID, 10, 64)
	if err != nil {
		userID = 0
	}

	return &dto.UserDTO{
		ID:          userID,
		Handle:      user.UserName,
		DisplayName: user.Name,
		LastSeenAt:  createdAt,
	}
}

// IsOriginalPost checks if a tweet is an original post (not a retweet or quote tweet)
func (c *TwitterClient) IsOriginalPost(tweet TweetData) bool {
	// Original posts don't have quoted_tweet or retweeted_tweet fields
	return tweet.QuotedTweet == nil && tweet.RetweetedTweet == nil
}
