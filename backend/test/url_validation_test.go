package test

import (
	"testing"

	"github.com/sopeal/AskYourFeed/internal/services"
)

// TestURLValidation tests that the Twitter client properly validates and normalizes URLs
func TestURLValidation(t *testing.T) {
	client := services.NewTwitterClient("test-key", nil)

	tests := []struct {
		name        string
		tweetData   services.TweetData
		expectedURL string
	}{
		{
			name: "Valid x.com URL",
			tweetData: services.TweetData{
				ID:  "1995915875489382540",
				URL: "https://x.com/MorawieckiM/status/1995915875489382540",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1995915875489382540",
			},
			expectedURL: "https://x.com/MorawieckiM/status/1995915875489382540",
		},
		{
			name: "Valid twitter.com URL",
			tweetData: services.TweetData{
				ID:  "1995915875489382540",
				URL: "https://twitter.com/MorawieckiM/status/1995915875489382540",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1995915875489382540",
			},
			expectedURL: "https://twitter.com/MorawieckiM/status/1995915875489382540",
		},
		{
			name: "Empty URL - should construct from author and ID",
			tweetData: services.TweetData{
				ID:  "1995915875489382540",
				URL: "",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1995915875489382540",
			},
			expectedURL: "https://twitter.com/MorawieckiM/status/1995915875489382540",
		},
		{
			name: "Invalid URL - should construct from author and ID",
			tweetData: services.TweetData{
				ID:  "1995915875489382540",
				URL: "https://invalid.com/something",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1995915875489382540",
			},
			expectedURL: "https://twitter.com/MorawieckiM/status/1995915875489382540",
		},
		{
			name: "URL with query parameters - should remove them",
			tweetData: services.TweetData{
				ID:  "1998849925971673364",
				URL: "https://x.com/MorawieckiM/status/1998849925971673364?s=20&t=abc123",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1998849925971673364",
			},
			expectedURL: "https://x.com/MorawieckiM/status/1998849925971673364",
		},
		{
			name: "URL with fragment - should remove it",
			tweetData: services.TweetData{
				ID:  "1998849925971673364",
				URL: "https://twitter.com/MorawieckiM/status/1998849925971673364#reply",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1998849925971673364",
			},
			expectedURL: "https://twitter.com/MorawieckiM/status/1998849925971673364",
		},
		{
			name: "URL with both query and fragment - should remove both",
			tweetData: services.TweetData{
				ID:  "1998849925971673364",
				URL: "https://x.com/MorawieckiM/status/1998849925971673364?s=20#reply",
				Author: services.UserData{
					ID:       "939053934232195072",
					UserName: "MorawieckiM",
				},
				Text:           "Test tweet",
				CreatedAt:      "Mon Dec 02 18:00:01 +0000 2025",
				ConversationId: "1998849925971673364",
			},
			expectedURL: "https://x.com/MorawieckiM/status/1998849925971673364",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := client.ConvertToDTO(tt.tweetData)
			if dto.URL != tt.expectedURL {
				t.Errorf("Expected URL %s, got %s", tt.expectedURL, dto.URL)
			}
		})
	}
}
