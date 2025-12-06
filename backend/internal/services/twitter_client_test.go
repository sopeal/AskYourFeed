package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserTweets(t *testing.T) {
	// Read the mock response from test.json
	mockResponse := `{
  "status" : "success",
  "code" : 0,
  "msg" : "success",
  "data" : {
    "pin_tweet" : null,
    "tweets" : [ {
      "type" : "tweet",
      "id" : "1995408537720094802",
      "url" : "https://x.com/OMalinkiewicz/status/1995408537720094802",
      "twitterUrl" : "https://twitter.com/OMalinkiewicz/status/1995408537720094802",
      "text" : "KIM JEST PIOTR KURCZEWSKI Z DC24 ASI?\nhttps://t.co/pwcpIl65w2\n\nMoi drodzy,\n\nWielu z Was pyta kim jest PIOTR KURCZEWSKI z DC24 ASI?\n\nTrudno prześledzić historię PIOTRA KURCZEWSKIEGO (https://t.co/rrn7Kt3Za2) przed 2004 rokiem, kiedy to zaczął prowadzić interesy w Polsce. Niektórzy mówią, że w latach 90. pracował on w Lybrand and Coopers (obecnie PWC) w Londynie, inni, że w tym samym Londynie w latach 1997-2001 pracował on też w ENRON Europe i ENRON Capital & Trade, spółce zależnej od słynnego amerykańskiego ENRON i że właśnie w ENRON Capital & Trade zbił PIOTR KURCZEWSKI swój pierwszy wielki kapitał.\n\nFaktem natomiast jest, że PIOTR KURCZEWSKI ma decydujący wpływ w takich spółkach sektora finansowego w Polsce jak:\n1. https://t.co/vXmzDtGwQK – przelewy elektroniczne (https://t.co/vecoZSG5Qk),\n2. https://t.co/NLUIAJH5zv – zakup biletów online (https://t.co/yxO4vkSNzJ),\n3. https://t.co/NevRS1I7Ya – serwis anonimowych elektronicznych przelewów pieniężnych (https://t.co/Y0TGk4GvQn),\n4. https://t.co/NCuDFYaRQM – konsumenckie zakupy na kredyt (https://t.co/o3cFmHQbmb),\n5. https://t.co/QH3xFliZB9 - największy krajowy konkurent serwisu wymiany walut https://t.co/jRLBzlAQJB (https://t.co/6iu98sdywL).\n\nFundusz PIOTRA KURCZEWSKIEGO DC24 ASI (https://t.co/O1yrOHcEJj; https://t.co/irWJkBNqbI) jest oficjalnym partnerem NCBiR Investment Fund ASI, funduszu inwestycyjnego Narodowego Centrum Badań i Rozwoju (https://t.co/Hrp8vXWPqm)\n\nDwie najbardziej zaufane osoby PIOTRA KURCZEWSKIEGO to:\n1. ANNA SZYMAŃSKA-PIPER (https://t.co/DzoBYcVYep) i\n2. ARKADIUSZ KRZEMIŃSKI (https://t.co/5ymobRvkQb), którzy zasiadają w zarządach i radach nadzorczych wszystkich spółek, w których PIOTR KURCZEWSKI pośrednio bądź bezpośrednio ma znaczący wpływ, tym w spółce P24 DOTCARD (https://t.co/xXvnEv7OTg), jedynego udziałowcy spółki PayPro (https://t.co/vecoZSG5Qk), właściciela serwisu https://t.co/vXmzDtGwQK.\n\nPrawną obsługą interesów Kurczewskiego zajmują się:\n1. GESSEL-KOZIOROWSKI KANCELARIA RADCÓW PRAWNYCH I ADWOKATÓW (https://t.co/JRdzdprNQ8), a mianowicie partner kancelarii\n2. MICHAŁ BOCHOWICZ (https://t.co/ilLFFbsIhQ; https://t.co/sFcJf1qeol), który sporządza wszystkie umowy pożyczkowe i inwestycyjne poprzez które Kurczewski kontroluje i przejmuje inne biznesy, w tym Columbus Energy i Saule. MICHAŁ BOCHOWICZ zasiadał również w radzie nadzorczej Saule.\n\nMICHAŁ BOCHOWICZ jest wśród beneficjentów rzeczywistych:\n1. GESSEL KOZIOROWSKI (https://t.co/rpNVWY7mc1),\n2. GESSEL, KOZIOROWSKI KANCELARIA RADCÓW PRAWNYCH I ADWOKATÓW (https://t.co/VRusyVYQjf) oraz\n3. GESSEL TRUST SERVICES (https://t.co/uGQSAqobMu).\n\nBędę wdzięczna wszystkim, którzy pomogą nam w zrozumieniu kim tak naprawdę jest PIOTR KURCZEWSKI.\n\nOlga Malinkiewicz\n\nWESPRZYJ WALKĘ OLGI MALINKIEWICZ O JEJ WYNALAZEK NA WHYDONATE. COM: https://t.co/pwcpIl65w2",
      "source" : "Twitter for iPhone",
      "retweetCount" : 142,
      "replyCount" : 51,
      "likeCount" : 1021,
      "quoteCount" : 6,
      "viewCount" : 78724,
      "createdAt" : "Mon Dec 01 08:24:02 +0000 2025",
      "lang" : "pl",
      "bookmarkCount" : 96,
      "isReply" : false,
      "inReplyToId" : null,
      "conversationId" : "1995408537720094802",
      "author" : {
        "type" : "user",
        "userName" : "OMalinkiewicz",
        "url" : "https://x.com/OMalinkiewicz",
        "id" : "1970073644421488640",
        "name" : "Olga Malinkiewicz, PhD",
        "isVerified" : false,
        "isBlueVerified" : false,
        "profilePicture" : "https://pbs.twimg.com/profile_images/1970143945318350848/rc02osf9_normal.jpg",
        "coverPicture" : "https://pbs.twimg.com/profile_banners/1970073644421488640/1758553939",
        "description" : "",
        "location" : "Wroclaw, Poland",
        "followers" : 10751,
        "following" : 26,
        "canDm" : true,
        "createdAt" : "Mon Sep 22 10:33:35 +0000 2025",
        "favouritesCount" : 3,
        "statusesCount" : 41
      },
      "quoted_tweet" : null,
      "retweeted_tweet" : null
    }, {
      "type" : "tweet",
      "id" : "1995403861901615156",
      "url" : "https://x.com/OMalinkiewicz/status/1995403861901615156",
      "twitterUrl" : "https://twitter.com/OMalinkiewicz/status/1995403861901615156",
      "text" : "WHO IS PIOTR KURCZEWSKI OF DC24 ASI?\nhttps://t.co/pwcpIl65w2\n\nDear friends,\n\nMany of you are asking me who is PIOTR KURCZEWSKI of DC24 ASI?\n\nIt's difficult to trace PIOTR KURCZEWSKI's history (https://t.co/rrn7Kt3Za2) before 2004, when he began doing business in Poland. Some say he worked at Lybrand and Coopers (now PWC) in London in the 1990s. Others say that from 1997 to 2001 in London he also worked at ENRON Europe and ENRON Capital & Trade, a subsidiary of the infamous American company ENRON, and that it was at ENRON Capital & Trade that PIOTR KURCZEWSKI made his first big capital.\n\nThe fact is, however, that PIOTR KURCZEWSKI has a decisive influence in such financial sector companies in Poland as:\n1. https://t.co/vXmzDtGwQK – electronic transfers (https://t.co/vecoZSG5Qk),\n2. https://t.co/NLUIAJH5zv – online ticket purchases (https://t.co/yxO4vkSNzJ),\n3. https://t.co/NevRS1I7Ya – anonymous electronic money transfer service (https://t.co/Y0TGk4GvQn),\n4. https://t.co/NCuDFYaRQM – consumer credit purchases (https://t.co/o3cFmHQbmb), and\n5. 5. https://t.co/QH3xFliZB9 – the largest domestic competitor to the currency exchange service https://t.co/jRLBzlAQJB (https://t.co/6iu98sdywL).\n\nPIOTR KURCZEWSKI's DC24 ASI Fund (https://t.co/O1yrOHcEJj; https://t.co/irWJkBNqbI) is an official partner of the NCBiR Investment Fund ASI, an investment fund of the National Centre for Research and Development (https://t.co/Hrp8vXWPqm).\n\nTwo of PIOTR KURCZEWSKI's most trusted individuals are:\n1. ANNA SZYMANSKA-PIPER (https://t.co/DzoBYcVYep) and 2. ARKADIUSZ KRZEMINSKI (https://t.co/5ymobRvkQb), who\nserve on the management and supervisory boards of all companies in which PIOTR KURCZEWSKI directly or indirectly has significant influence, including P24 DOTCARD (https://t.co/xXvnEv7OTg), the sole shareholder of PayPro (https://t.co/vecoZSG5Qk), the owner of the https://t.co/vXmzDtGwQK service.\n\nKurczewski's legal interests are handled by:\n1. GESSEL-KOZIOROWSKI law firm (https://t.co/JRdzdprNQ8) and\n2. MICHAL BOCHOWICZ (https://t.co/ilLFFbsIhQ, https://t.co/sFcJf1qeol), a partner in the law firm who drafts all loan and investment agreements through which PIOTR KURCZEWSKI controls and takes over other businesses, including Columbus Energy and Saule. MICHAL BOCHOWICZ also served on the supervisory board of Saule.\n\nMICHAL BOCHOWICZ is among the beneficial owners of:\n1. GESSEL KOZIOROWSKI (https://t.co/rpNVWY7mc1),\n2. GESSEL, KOZIOROWSKI LEGAL COUNSEL AND ADVOCATS (https://t.co/VRusyVYQjf), and\n3. GESSEL TRUST SERVICES (https://t.co/uGQSAqobMu).\n\nI would be grateful to anyone who helps us understand who PIOTR KURCZEWSKI really is.\n\nOlga Malinkiewicz\n\nSUPPORT OLGA MALINKIEWICZ'S FIGHT FOR HER INVENTION OF WHYDONATE. COM: https://t.co/pwcpIl65w2",
      "source" : "Twitter for iPhone",
      "retweetCount" : 24,
      "replyCount" : 11,
      "likeCount" : 209,
      "quoteCount" : 0,
      "viewCount" : 12926,
      "createdAt" : "Mon Dec 01 08:05:28 +0000 2025",
      "lang" : "en",
      "bookmarkCount" : 6,
      "isReply" : false,
      "inReplyToId" : null,
      "conversationId" : "1995403861901615156",
      "author" : {
        "type" : "user",
        "userName" : "OMalinkiewicz",
        "url" : "https://x.com/OMalinkiewicz",
        "id" : "1970073644421488640",
        "name" : "Olga Malinkiewicz, PhD",
        "isVerified" : false,
        "isBlueVerified" : false,
        "profilePicture" : "https://pbs.twimg.com/profile_images/1970143945318350848/rc02osf9_normal.jpg",
        "coverPicture" : "https://pbs.twimg.com/profile_banners/1970073644421488640/1758553939",
        "description" : "",
        "location" : "Wroclaw, Poland",
        "followers" : 10751,
        "following" : 26,
        "canDm" : true,
        "createdAt" : "Mon Sep 22 10:33:35 +0000 2025",
        "favouritesCount" : 3,
        "statusesCount" : 41
      },
      "quoted_tweet" : null,
      "retweeted_tweet" : null
    } ]
  },
  "has_next_page" : true,
  "next_cursor" : "DAADDAABCgABG7EdfjCWgFIKAAIbnyQUqpeg1wAIAAIAAAACCAADAAAAAAgABAAAAAAKAAUbtybgQMAnEAoABhu3JuBAv9jwAAA"
}`

	tests := []struct {
		name           string
		username       string
		cursor         string
		mockStatusCode int
		mockResponse   string
		wantErr        bool
		validateResult func(t *testing.T, resp *TweetResponse)
	}{
		{
			name:           "successful request without cursor",
			username:       "OMalinkiewicz",
			cursor:         "",
			mockStatusCode: http.StatusOK,
			mockResponse:   mockResponse,
			wantErr:        false,
			validateResult: func(t *testing.T, resp *TweetResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "success", resp.Status)
				assert.True(t, resp.HasNextPage)
				assert.Equal(t, "DAADDAABCgABG7EdfjCWgFIKAAIbnyQUqpeg1wAIAAIAAAACCAADAAAAAAgABAAAAAAKAAUbtybgQMAnEAoABhu3JuBAv9jwAAA", resp.NextCursor)
				assert.Len(t, resp.Tweets, 2)

				// Validate first tweet
				firstTweet := resp.Tweets[0]
				assert.Equal(t, "1995408537720094802", firstTweet.ID)
				assert.Equal(t, "tweet", firstTweet.Type)
				assert.Contains(t, firstTweet.Text, "KIM JEST PIOTR KURCZEWSKI Z DC24 ASI?")
				assert.Equal(t, "https://x.com/OMalinkiewicz/status/1995408537720094802", firstTweet.URL)
				assert.Equal(t, "Twitter for iPhone", firstTweet.Source)
				assert.Equal(t, 142, firstTweet.RetweetCount)
				assert.Equal(t, 51, firstTweet.ReplyCount)
				assert.Equal(t, 1021, firstTweet.LikeCount)
				assert.Equal(t, 6, firstTweet.QuoteCount)
				assert.Equal(t, 78724, firstTweet.ViewCount)
				assert.Equal(t, 96, firstTweet.BookmarkCount)
				assert.Equal(t, "pl", firstTweet.Lang)
				assert.False(t, firstTweet.IsReply)
				assert.Equal(t, "1995408537720094802", firstTweet.ConversationId)
				assert.Nil(t, firstTweet.QuotedTweet)
				assert.Nil(t, firstTweet.RetweetedTweet)

				// Validate author
				author := firstTweet.Author
				assert.Equal(t, "OMalinkiewicz", author.UserName)
				assert.Equal(t, "1970073644421488640", author.ID)
				assert.Equal(t, "Olga Malinkiewicz, PhD", author.Name)
				assert.Equal(t, "Wroclaw, Poland", author.Location)
				assert.Equal(t, 10751, author.Followers)
				assert.Equal(t, 26, author.Following)
				assert.False(t, author.IsBlueVerified)
				assert.True(t, author.CanDm)

				// Validate second tweet
				secondTweet := resp.Tweets[1]
				assert.Equal(t, "1995403861901615156", secondTweet.ID)
				assert.Contains(t, secondTweet.Text, "WHO IS PIOTR KURCZEWSKI OF DC24 ASI?")
				assert.Equal(t, 24, secondTweet.RetweetCount)
				assert.Equal(t, 11, secondTweet.ReplyCount)
				assert.Equal(t, 209, secondTweet.LikeCount)
				assert.Equal(t, "en", secondTweet.Lang)
			},
		},
		{
			name:           "successful request with cursor",
			username:       "OMalinkiewicz",
			cursor:         "DAADDAABCgABG7EdfjCWgFIKAAIbnyQUqpeg1wAIAAIAAAACCAADAAAAAAgABAAAAAAKAAUbtybgQMAnEAoABhu3JuBAv9jwAAA",
			mockStatusCode: http.StatusOK,
			mockResponse:   mockResponse,
			wantErr:        false,
			validateResult: func(t *testing.T, resp *TweetResponse) {
				assert.NotNil(t, resp)
				assert.Len(t, resp.Tweets, 2)
			},
		},
		{
			name:           "API error - 404 not found",
			username:       "nonexistent",
			cursor:         "",
			mockStatusCode: http.StatusNotFound,
			mockResponse:   `{"error": "User not found"}`,
			wantErr:        true,
			validateResult: nil,
		},
		{
			name:           "API error - 500 internal server error",
			username:       "OMalinkiewicz",
			cursor:         "",
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   `{"error": "Internal server error"}`,
			wantErr:        true,
			validateResult: nil,
		},
		{
			name:           "invalid JSON response",
			username:       "OMalinkiewicz",
			cursor:         "",
			mockStatusCode: http.StatusOK,
			mockResponse:   `{invalid json}`,
			wantErr:        true,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/twitter/user/last_tweets", r.URL.Path)
				assert.Equal(t, "test-api-key", r.Header.Get("x-api-key"))
				assert.Equal(t, "AskYourFeed/1.0", r.Header.Get("User-Agent"))

				// Verify query parameters
				query := r.URL.Query()
				assert.Equal(t, tt.username, query.Get("userName"))
				if tt.cursor != "" {
					assert.Equal(t, tt.cursor, query.Get("cursor"))
				}

				// Send mock response
				w.WriteHeader(tt.mockStatusCode)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// Create Twitter client with mock server URL
			client := NewTwitterClient("test-api-key", server.Client())
			client.baseURL = server.URL

			// Call the function
			ctx := context.Background()
			resp, err := client.GetUserTweets(ctx, tt.username, tt.cursor)

			// Validate error
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				// Run custom validation if provided
				if tt.validateResult != nil {
					tt.validateResult(t, resp)
				}
			}
		})
	}
}

func TestGetUserTweets_ContextCancellation(t *testing.T) {
	// Create a mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow response
		select {
		case <-r.Context().Done():
			return
		}
	}))
	defer server.Close()

	client := NewTwitterClient("test-api-key", server.Client())
	client.baseURL = server.URL

	// Create a context that is already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Call the function with cancelled context
	resp, err := client.GetUserTweets(ctx, "testuser", "")

	// Should return an error
	assert.Error(t, err)
	assert.Nil(t, resp)
}

func TestGetUserTweets_ResponseStructure(t *testing.T) {
	// Test that the response structure matches the expected format
	mockData := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"tweets": []map[string]interface{}{
				{
					"id":   "123456789",
					"text": "Test tweet",
					"author": map[string]interface{}{
						"id":       "987654321",
						"userName": "testuser",
						"name":     "Test User",
					},
					"retweetCount":   10,
					"replyCount":     5,
					"likeCount":      100,
					"quoteCount":     2,
					"viewCount":      1000,
					"bookmarkCount":  15,
					"createdAt":      "Mon Dec 01 08:24:02 +0000 2025",
					"lang":           "en",
					"isReply":        false,
					"conversationId": "123456789",
				},
			},
		},
		"has_next_page": false,
		"next_cursor":   "",
	}

	mockJSON, err := json.Marshal(mockData)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(mockJSON)
	}))
	defer server.Close()

	client := NewTwitterClient("test-api-key", server.Client())
	client.baseURL = server.URL

	resp, err := client.GetUserTweets(context.Background(), "testuser", "")

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Tweets, 1)
	assert.Equal(t, "123456789", resp.Tweets[0].ID)
	assert.Equal(t, "Test tweet", resp.Tweets[0].Text)
	assert.False(t, resp.HasNextPage)
}
