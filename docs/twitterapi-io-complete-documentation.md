# twitterapi.io API – Complete Documentation for AI

## 1. Introduction

### What is twitterapi.io?

twitterapi.io is a third-party Twitter/X API service that provides comprehensive access to Twitter data without the complexity and cost of the official Twitter API. It offers a simpler, more cost-effective alternative for developers and researchers.

**Key Features:**
- **No Twitter App Required**: No need to create a Twitter developer account or app
- **No OAuth Complexity**: Simple API key authentication
- **Cost-Effective**: 96% cheaper than official Twitter API
  - $0.15/1k tweets
  - $0.18/1k user profiles
  - $0.15/1k followers
  - Minimum charge: $0.00015 per request (even if no data returned)
- **High Performance**: 
  - Average response time: 700ms
  - Supports up to 200 QPS (queries per second) per client
  - Proven stability with over 1,000,000 API calls
- **RESTful Design**: Standard OpenAPI specifications
- **Special Offers**: Discounted rates for students and research institutions

### Base URL

```
https://api.twitterapi.io
```

### API Version

The API follows RESTful conventions with versioned endpoints where applicable (v2 for action endpoints).

## 2. Authorization and Access

### Getting Your API Key

1. Visit the [TwitterApiIO Dashboard](https://twitterapi.io/dashboard)
2. Log in to your account
3. Your API key will be displayed on the dashboard homepage

### Authentication Method

All API requests require authentication using the `x-api-key` header (case-insensitive, can also use `X-API-Key`).

**Header Format:**
```
x-api-key: YOUR_API_KEY
```

**Example Requests:**

**cURL:**
```bash
curl --location 'https://api.twitterapi.io/twitter/user/followings?userName=KaitoEasyAPI' \
  --header 'x-api-key: your_api_key_here'
```

**Python:**
```python
import requests

url = 'https://api.twitterapi.io/twitter/user/followings?userName=KaitoEasyAPI'
headers = {'x-api-key': 'your_api_key_here'}
response = requests.get(url, headers=headers)
print(response.json())
```

**JavaScript (Node.js):**
```javascript
const axios = require('axios');

const url = 'https://api.twitterapi.io/twitter/user/followings?userName=KaitoEasyAPI';
const headers = { 'x-api-key': 'your_api_key_here' };

axios.get(url, { headers })
  .then(response => console.log(response.data))
  .catch(error => console.error(error));
```

**Go:**
```go
package main

import (
    "fmt"
    "io"
    "net/http"
)

func main() {
    url := "https://api.twitterapi.io/twitter/user/followings?userName=KaitoEasyAPI"
    
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("x-api-key", "your_api_key_here")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    fmt.Println(string(body))
}
```

## 3. Rate Limits and Pricing

### Pricing Structure

| Resource Type | Price per 1,000 Units |
|--------------|----------------------|
| Tweets | $0.15 |
| User Profiles | $0.18 |
| Followers | $0.15 |
| Minimum per Request | $0.00015 |

**Note**: The minimum charge applies even if the request returns no data.

### Rate Limits

- **Maximum QPS**: 200 queries per second per client
- **Average Response Time**: 700ms

### Handling Rate Limits

When you exceed rate limits, the API will return a `429 Too Many Requests` status code. Implement exponential backoff in your application:

```python
import time
import requests

def make_request_with_retry(url, headers, max_retries=3):
    for attempt in range(max_retries):
        response = requests.get(url, headers=headers)
        
        if response.status_code == 429:
            wait_time = (2 ** attempt) * 1  # Exponential backoff
            print(f"Rate limited. Waiting {wait_time} seconds...")
            time.sleep(wait_time)
            continue
            
        return response
    
    raise Exception("Max retries exceeded")
```

## 4. Data Models

### User Object

The User object represents a Twitter/X user profile.

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `type` | string | Always "user" | Yes |
| `id` | string | Unique user ID | Yes |
| `userName` | string | Username (handle without @) | Yes |
| `name` | string | Display name | Yes |
| `url` | string | Profile URL | Yes |
| `profilePicture` | string | Profile picture URL | Yes |
| `coverPicture` | string | Cover/banner image URL | No |
| `description` | string | Bio/description | No |
| `location` | string | User's location | No |
| `isBlueVerified` | boolean | Twitter Blue verification status | Yes |
| `verifiedType` | string | Type of verification | No |
| `followers` | integer | Follower count | Yes |
| `following` | integer | Following count | Yes |
| `canDm` | boolean | Whether user accepts DMs | Yes |
| `createdAt` | string | Account creation date | Yes |
| `favouritesCount` | integer | Number of likes | Yes |
| `hasCustomTimelines` | boolean | Has custom timelines | Yes |
| `isTranslator` | boolean | Is a translator | Yes |
| `mediaCount` | integer | Media items count | Yes |
| `statusesCount` | integer | Total tweets count | Yes |
| `withheldInCountries` | array[string] | Countries where content is withheld | No |
| `affiliatesHighlightedLabel` | object | Affiliate labels | No |
| `possiblySensitive` | boolean | Sensitive content flag | No |
| `pinnedTweetIds` | array[string] | IDs of pinned tweets | No |
| `isAutomated` | boolean | Is automated account | No |
| `automatedBy` | string | Automation source | No |
| `unavailable` | boolean | Account unavailable | No |
| `message` | string | Status message | No |
| `unavailableReason` | string | Reason for unavailability | No |
| `profile_bio` | object | Detailed bio with entities | No |

**Profile Bio Object:**
```json
{
  "description": "string",
  "entities": {
    "description": {
      "urls": [
        {
          "display_url": "string",
          "expanded_url": "string",
          "indices": [0, 23],
          "url": "string"
        }
      ]
    },
    "url": {
      "urls": [
        {
          "display_url": "string",
          "expanded_url": "string",
          "indices": [0, 23],
          "url": "string"
        }
      ]
    }
  }
}
```

### Tweet Object

The Tweet object represents a Twitter/X post.

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `type` | string | Always "tweet" | Yes |
| `id` | string | Unique tweet ID | Yes |
| `url` | string | Tweet URL | Yes |
| `text` | string | Tweet text content | Yes |
| `source` | string | Source application | Yes |
| `retweetCount` | integer | Number of retweets | Yes |
| `replyCount` | integer | Number of replies | Yes |
| `likeCount` | integer | Number of likes | Yes |
| `quoteCount` | integer | Number of quote tweets | Yes |
| `viewCount` | integer | Number of views | Yes |
| `bookmarkCount` | integer | Number of bookmarks | Yes |
| `createdAt` | string | Creation timestamp | Yes |
| `lang` | string | Language code | Yes |
| `isReply` | boolean | Is a reply tweet | Yes |
| `inReplyToId` | string | ID of tweet being replied to | No |
| `inReplyToUserId` | string | User ID being replied to | No |
| `inReplyToUsername` | string | Username being replied to | No |
| `conversationId` | string | Conversation thread ID | Yes |
| `displayTextRange` | array[integer] | Text display range | No |
| `author` | User object | Tweet author information | Yes |
| `entities` | object | Tweet entities (hashtags, URLs, mentions) | No |
| `quoted_tweet` | Tweet object | Quoted tweet (if quote tweet) | No |
| `retweeted_tweet` | Tweet object | Retweeted tweet (if retweet) | No |
| `isLimitedReply` | boolean | Has limited reply settings | No |

**Entities Object:**
```json
{
  "hashtags": [
    {
      "indices": [0, 10],
      "text": "hashtag"
    }
  ],
  "urls": [
    {
      "display_url": "example.com",
      "expanded_url": "https://example.com/full/url",
      "indices": [20, 43],
      "url": "https://t.co/shortened"
    }
  ],
  "user_mentions": [
    {
      "id_str": "123456789",
      "name": "Display Name",
      "screen_name": "username"
    }
  ]
}
```

### Community Object

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `id` | string | Community ID | Yes |
| `name` | string | Community name | Yes |
| `description` | string | Community description | No |
| `memberCount` | integer | Number of members | Yes |
| `moderatorCount` | integer | Number of moderators | Yes |
| `createdAt` | string | Creation timestamp | Yes |

### Pagination Object

Many endpoints return paginated results with these fields:

| Field | Type | Description |
|-------|------|-------------|
| `has_next_page` | boolean | Whether more results exist |
| `next_cursor` | string | Cursor token for next page |

## 5. Endpoints

### User Endpoints

#### GET /twitter/user/info

Get detailed information about a user by username.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `userName` | string | Yes | Username (without @) |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/info?userName=elonmusk' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "data": {
    "type": "user",
    "userName": "elonmusk",
    "url": "https://twitter.com/elonmusk",
    "id": "44196397",
    "name": "Elon Musk",
    "isBlueVerified": true,
    "verifiedType": "Business",
    "profilePicture": "https://pbs.twimg.com/profile_images/...",
    "coverPicture": "https://pbs.twimg.com/profile_banners/...",
    "description": "Tesla, SpaceX, Neuralink, xAI",
    "location": "Texas, USA",
    "followers": 150000000,
    "following": 500,
    "canDm": false,
    "createdAt": "2009-06-02T20:12:29.000Z",
    "favouritesCount": 50000,
    "statusesCount": 35000
  },
  "status": "success",
  "msg": ""
}
```

**Error Response (400):**
```json
{
  "status": "error",
  "msg": "User not found"
}
```

---

#### GET /twitter/user/batch_get_user_by_userids

Get information for multiple users by their user IDs.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `userIds` | string | Yes | Comma-separated user IDs |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/batch_get_user_by_userids?userIds=44196397,813286' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "44196397",
      "userName": "elonmusk",
      "name": "Elon Musk",
      "followers": 150000000
    },
    {
      "type": "user",
      "id": "813286",
      "userName": "BarackObama",
      "name": "Barack Obama",
      "followers": 130000000
    }
  ],
  "status": "success"
}
```

---

#### GET /twitter/user/last_tweets

Get the most recent tweets from a user.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `userName` | string | Yes | Username (without @) | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/last_tweets?userName=elonmusk' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1234567890",
      "text": "Tweet content here",
      "author": {
        "userName": "elonmusk",
        "name": "Elon Musk"
      },
      "createdAt": "2024-01-15T10:30:00.000Z",
      "likeCount": 50000,
      "retweetCount": 10000,
      "replyCount": 5000
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/user/followers

Get a user's followers list.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `userName` | string | Yes | Username (without @) | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/followers?userName=elonmusk&cursor=' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "123456",
      "userName": "follower1",
      "name": "Follower Name",
      "followers": 1000,
      "following": 500
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/user/followings

Get the list of users that a user follows.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `userName` | string | Yes | Username (without @) | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/followings?userName=elonmusk' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "789012",
      "userName": "following1",
      "name": "Following Name",
      "followers": 5000,
      "following": 200
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/user/mentions

Get tweets that mention a specific user.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `userName` | string | Yes | Username (without @) | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/mentions?userName=elonmusk' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1234567890",
      "text": "Hey @elonmusk, great work!",
      "author": {
        "userName": "someuser",
        "name": "Some User"
      },
      "createdAt": "2024-01-15T10:30:00.000Z"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/user/check_follow_relationship

Check if one user follows another.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `sourceUserName` | string | Yes | Source username |
| `targetUserName` | string | Yes | Target username |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/check_follow_relationship?sourceUserName=user1&targetUserName=user2' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "following": true,
  "followed_by": false,
  "status": "success"
}
```

---

#### GET /twitter/user/search

Search for users by keyword.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `keyword` | string | Yes | Search keyword | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/search?keyword=elon' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "44196397",
      "userName": "elonmusk",
      "name": "Elon Musk",
      "description": "Tesla, SpaceX...",
      "followers": 150000000
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/user/verified_followers

Get a user's verified followers.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `userName` | string | Yes | Username (without @) | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/user/verified_followers?userName=elonmusk' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "123456",
      "userName": "verifieduser",
      "name": "Verified User",
      "isBlueVerified": true,
      "verifiedType": "Business"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

### Tweet Endpoints

#### GET /twitter/tweets

Get tweets by their IDs.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tweet_ids` | string | Yes | Comma-separated tweet IDs (e.g., "1846987139428634858,1866332309399781537") |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweets?tweet_ids=1846987139428634858,1866332309399781537' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1846987139428634858",
      "url": "https://twitter.com/user/status/1846987139428634858",
      "text": "This is a tweet",
      "source": "Twitter Web App",
      "retweetCount": 100,
      "replyCount": 50,
      "likeCount": 500,
      "quoteCount": 25,
      "viewCount": 10000,
      "bookmarkCount": 75,
      "createdAt": "2024-01-15T10:30:00.000Z",
      "lang": "en",
      "isReply": false,
      "author": {
        "type": "user",
        "userName": "username",
        "name": "Display Name",
        "id": "123456789"
      },
      "entities": {
        "hashtags": [
          {
            "indices": [0, 10],
            "text": "example"
          }
        ],
        "urls": [],
        "user_mentions": []
      }
    }
  ],
  "status": "success",
  "message": ""
}
```

---

#### GET /twitter/tweet/replies

Get replies to a specific tweet.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `tweetId` | string | Yes | Tweet ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweet/replies?tweetId=1846987139428634858' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1846987139428634859",
      "text": "This is a reply",
      "isReply": true,
      "inReplyToId": "1846987139428634858",
      "author": {
        "userName": "replier",
        "name": "Replier Name"
      }
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/tweet/quotes

Get quote tweets of a specific tweet.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `tweetId` | string | Yes | Tweet ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweet/quotes?tweetId=1846987139428634858' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1846987139428634860",
      "text": "Quoting this tweet",
      "quoted_tweet": {
        "id": "1846987139428634858",
        "text": "Original tweet"
      }
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/tweet/retweeters

Get users who retweeted a specific tweet.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `tweetId` | string | Yes | Tweet ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweet/retweeters?tweetId=1846987139428634858' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "123456",
      "userName": "retweeter",
      "name": "Retweeter Name"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/tweet/thread_context

Get the full thread context for a tweet (parent tweets and replies).

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `tweetId` | string | Yes | Tweet ID |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweet/thread_context?tweetId=1846987139428634858' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1846987139428634857",
      "text": "First tweet in thread"
    },
    {
      "type": "tweet",
      "id": "1846987139428634858",
      "text": "Second tweet in thread",
      "inReplyToId": "1846987139428634857"
    }
  ],
  "status": "success"
}
```

---

#### GET /twitter/article

Get Twitter article/long-form content.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `articleId` | string | Yes | Article ID |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/article?articleId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "article": {
    "id": "123456",
    "title": "Article Title",
    "content": "Full article content...",
    "author": {
      "userName": "author",
      "name": "Author Name"
    },
    "createdAt": "2024-01-15T10:30:00.000Z"
  },
  "status": "success"
}
```

---

#### GET /twitter/tweet/advanced_search

Advanced tweet search with filters.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `query` | string | Yes | Search query (supports advanced operators) | - |
| `queryType` | string | Yes | "Latest" or "Top" | "Latest" |
| `cursor` | string | No | Pagination cursor | "" |

**Query Syntax Examples:**
- `"AI" OR "Twitter"` - Search for tweets containing AI or Twitter
- `from:elonmusk` - Tweets from specific user
- `since:2021-12-31_23:59:59_UTC` - Tweets since date
- `until:2024-01-01_00:00:00_UTC` - Tweets until date
- `"AI" from:elonmusk since:2023-01-01` - Combined filters

For more query examples, see: https://github.com/igorbrigadir/twitter-advanced-search

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/tweet/advanced_search?query=AI%20from:elonmusk&queryType=Latest' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1234567890",
      "text": "Tweet about AI",
      "author": {
        "userName": "elonmusk",
        "name": "Elon Musk"
      },
      "createdAt": "2024-01-15T10:30:00.000Z",
      "likeCount": 50000
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

### List Endpoints

#### GET /twitter/list/followers

Get followers of a Twitter list.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `listId` | string | Yes | List ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/list/followers?listId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "789012",
      "userName": "follower",
      "name": "Follower Name"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/list/members

Get members of a Twitter list.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `listId` | string | Yes | List ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/list/members?listId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "345678",
      "userName": "member",
      "name": "Member Name"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

### Community Endpoints

#### GET /twitter/community/info

Get information about a community by ID.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `communityId` | string | Yes | Community ID |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/community/info?communityId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "community": {
    "id": "123456",
    "name": "Community Name",
    "description": "Community description",
    "memberCount": 5000,
    "moderatorCount": 10,
    "createdAt": "2023-01-15T10:30:00.000Z"
  },
  "status": "success"
}
```

---

#### GET /twitter/community/members

Get members of a community.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `communityId` | string | Yes | Community ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/community/members?communityId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "789012",
      "userName": "member",
      "name": "Member Name"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/community/moderators

Get moderators of a community.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `communityId` | string | Yes | Community ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/community/moderators?communityId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "users": [
    {
      "type": "user",
      "id": "456789",
      "userName": "moderator",
      "name": "Moderator Name"
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/community/tweets

Get tweets from a specific community.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `communityId` | string | Yes | Community ID | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/community/tweets?communityId=123456' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1234567890",
      "text": "Community tweet",
      "author": {
        "userName": "member",
        "name": "Member Name"
      }
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

#### GET /twitter/community/search_tweets

Search tweets across all communities.

**Query Parameters:**

| Name | Type | Required | Description | Default |
|------|------|----------|-------------|---------|
| `query` | string | Yes | Search query | - |
| `cursor` | string | No | Pagination cursor | "" |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/community/search_tweets?query=AI' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "tweets": [
    {
      "type": "tweet",
      "id": "1234567890",
      "text": "Tweet about AI in community",
      "author": {
        "userName": "user",
        "name": "User Name"
      }
    }
  ],
  "has_next_page": true,
  "next_cursor": "DAABCgABGVz__-oKAAIZXNs..."
}
```

---

### Trend Endpoints

#### GET /twitter/trends

Get current trending topics.

**Query Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `location` | string | No | Location WOEID (default: worldwide) |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/trends' \
  --header 'x-api-key: your_api_key_here'
```

**Success Response (200):**
```json
{
  "trends": [
    {
      "name": "#TrendingTopic",
      "url": "https://twitter.com/search?q=%23TrendingTopic",
      "tweet_volume": 50000
    }
  ],
  "status": "success"
}
```

---

### My Account Endpoints

#### GET /twitter/my/info

Get information about the authenticated account.

**Headers:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `x-api-key` | string | Yes | API key |
| `login_cookies` | string | Yes | Login cookies from /twitter/user_login_v2 |

**Example Request:**
```bash
curl --request GET \
  --url 'https://api.twitterapi.io/twitter/my/info' \
  --header 'x-api-key: your_api_key_here' \
  --header 'login_cookies: your_login_cookies_here'
```

**Success Response (200):**
```json
{
  "data": {
    "type": "user",
    "id": "123456789",
    "userName": "myusername",
    "name": "My Name",
    "followers": 1000,
    "following": 500
  },
  "status": "success"
}
```

## 6. Common Patterns and Best Practices

### Pagination

Most list endpoints support cursor-based pagination:

```python
import requests

def fetch_all_followers(username, api_key):
    all_followers = []
    cursor = ""
    
    while True:
        url = f"https://api.twitterapi.io/twitter/user/followers"
        params = {"userName": username, "cursor": cursor}
        headers = {"x-api-key": api_key}
        
        response = requests.get(url, params=params, headers=headers)
        data = response.json()
        
        all_followers.extend(data.get("users", []))
        
        if not data.get("has_next_page", False):
            break
            
        cursor = data.get("next_cursor", "")
    
    return all_followers
```

### Error Handling

Always implement proper error handling:

```python
import requests
import time

def make_api_request(url, headers, max_retries=3):
    for attempt in range(max_retries):
        try:
            response = requests.get(url, headers=headers, timeout=30)
            
            if response.status_code == 200:
                return response.json()
            elif response.status_code == 429:
                # Rate limited
                wait_time = (2 ** attempt) * 2
                print(f"Rate limited. Waiting {wait_time}s...")
                time.sleep(wait_time)
            elif response.status_code == 400:
                # Bad request
                print(f"Bad request: {response.json()}")
                return None
            elif response.status_code >= 500:
                # Server error
                print(f"Server error. Retrying...")
                time.sleep(2 ** attempt)
            else:
                print(f"Unexpected status: {response.status_code}")
                return None
                
        except requests.exceptions.Timeout:
            print(f"Request timeout. Attempt {attempt + 1}/{max_retries}")
            time.sleep(2 ** attempt)
        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            return None
    
    return None
```

### Batch Operations

When fetching multiple users or tweets, use batch endpoints:

```python
# Instead of multiple single requests
user_ids = ["123", "456", "789"]
users = []
for user_id in user_ids:
    # DON'T DO THIS - inefficient
    response = get_user_by_id(user_id)
    users.append(response)

# Use batch endpoint instead
user_ids_str = ",".join(user_ids)
response = requests.get(
    f"https://api.twitterapi.io/twitter/user/batch_get_user_by_userids?userIds={user_ids_str}",
    headers={"x-api-key": api_key}
)
users = response.json().get("users", [])
```

### Advanced Search Queries

Use the advanced search endpoint with complex queries:

```python
# Search for tweets from specific users about AI
query = '"AI" OR "artificial intelligence" from:elonmusk OR from:sama'
response = requests.get(
    "https://api.twitterapi.io/twitter/tweet/advanced_search",
    params={"query": query, "queryType": "Latest"},
    headers={"x-api-key": api_key}
)

# Search with date range
query = 'bitcoin since:2024-01-01_00:00:00_UTC until:2024-12-31_23:59:59_UTC'
response = requests.get(
    "https://api.twitterapi.io/twitter/tweet/advanced_search",
    params={"query": query, "queryType": "Top"},
    headers={"x-api-key": api_key}
)
```

## 7. Common Issues and Troubleshooting

### Rate Limiting

**Issue**: Receiving 429 status codes
**Solution**: 
- Implement exponential backoff
- Reduce request frequency
- Consider upgrading your plan for higher limits

### Authentication Errors

**Issue**: 401 Unauthorized errors
**Solution**:
- Verify your API key is correct
- Ensure the `x-api-key` header is properly set
- Check that your API key hasn't expired

### Pagination Issues

**Issue**: Not receiving all results
**Solution**:
- Always check `has_next_page` field
- Use the `next_cursor` value for subsequent requests
- Don't assume a fixed number of pages

### Proxy Requirements for Action Endpoints

**Issue**: Action endpoints failing without proxy
**Solution**:
- Use high-quality residential proxies
- Avoid free proxies (they often don't work)
- Format: `http://username:password@ip:port`
- Recommended provider: https://www.webshare.io/?referral_code=4e0q1n00a504

### Empty Results

**Issue**: API returns success but empty data
**Solution**:
- Verify the username/ID exists
- Check if the account is private or suspended
- Ensure query parameters are correctly formatted
- Note: You're still charged the minimum fee even for empty results

### Timeout Errors

**Issue**: Requests timing out
**Solution**:
- Increase timeout value in your HTTP client
- The API average response time is 700ms, but some requests may take longer
- Implement retry logic with exponential backoff

## 8. HTTP Status Codes

| Status Code | Meaning | Action |
|-------------|---------|--------|
| 200 | Success | Request completed successfully |
| 400 | Bad Request | Check request parameters and format |
| 401 | Unauthorized | Verify API key is correct |
| 404 | Not Found | Resource doesn't exist |
| 429 | Too Many Requests | Implement rate limiting and backoff |
| 500 | Internal Server Error | Retry request after delay |
| 503 | Service Unavailable | Service temporarily down, retry later |

## 9. API Versioning and Changes

### Current Version

The API is currently in active development with v2 endpoints for action operations.

### Endpoint Versions

- **v1 (implicit)**: Read-only endpoints (user info, tweets, search, etc.)
- **v2**: Action endpoints (login, create tweet, follow, like, etc.)

### Deprecated Endpoints

The following endpoints are deprecated and should not be used for new development:

- `POST /twitter/login/email_or_username` - Use `/twitter/user_login_v2` instead
- `POST /twitter/login/2fa` - Integrated into v2 login flow
- `POST /twitter/upload_image` - Use `/twitter/upload_media_v2` instead
- `POST /twitter/create_tweet` - Use `/twitter/create_tweet_v2` instead
- `POST /twitter/like_tweet` - Use `/twitter/like_tweet_v2` instead
- `POST /twitter/retweet_tweet` - Use `/twitter/retweet_v2` instead

### Changelog

**Note**: For the most up-to-date changelog, visit https://docs.twitterapi.io

## 10. Code Examples

### Python Complete Example

```python
import requests
import time
from typing import List, Dict, Optional

class TwitterAPIClient:
    def __init__(self, api_key: str):
        self.api_key = api_key
        self.base_url = "https://api.twitterapi.io"
        self.headers = {"x-api-key": api_key}
    
    def _make_request(self, method: str, endpoint: str, **kwargs) -> Optional[Dict]:
        """Make API request with error handling and retries"""
        url = f"{self.base_url}{endpoint}"
        max_retries = 3
        
        for attempt in range(max_retries):
            try:
                response = requests.request(
                    method, url, headers=self.headers, timeout=30, **kwargs
                )
                
                if response.status_code == 200:
                    return response.json()
                elif response.status_code == 429:
                    wait_time = (2 ** attempt) * 2
                    print(f"Rate limited. Waiting {wait_time}s...")
                    time.sleep(wait_time)
                else:
                    print(f"Error {response.status_code}: {response.text}")
                    return None
                    
            except requests.exceptions.RequestException as e:
                print(f"Request failed: {e}")
                if attempt < max_retries - 1:
                    time.sleep(2 ** attempt)
        
        return None
    
    def get_user_info(self, username: str) -> Optional[Dict]:
        """Get user information by username"""
        return self._make_request(
            "GET", "/twitter/user/info", params={"userName": username}
        )
    
    def get_user_tweets(self, username: str, limit: int = 100) -> List[Dict]:
        """Get user's recent tweets with pagination"""
        tweets = []
        cursor = ""
        
        while len(tweets) < limit:
            data = self._make_request(
                "GET", "/twitter/user/last_tweets",
                params={"userName": username, "cursor": cursor}
            )
            
            if not data:
                break
            
            tweets.extend(data.get("tweets", []))
            
            if not data.get("has_next_page", False):
                break
            
            cursor = data.get("next_cursor", "")
        
        return tweets[:limit]
    
    def search_tweets(self, query: str, query_type: str = "Latest") -> List[Dict]:
        """Search tweets with advanced query"""
        return self._make_request(
            "GET", "/twitter/tweet/advanced_search",
            params={"query": query, "queryType": query_type}
        )
    
    def get_tweet_thread(self, tweet_id: str) -> List[Dict]:
        """Get full tweet thread context"""
        data = self._make_request(
            "GET", "/twitter/tweet/thread_context",
            params={"tweetId": tweet_id}
        )
        return data.get("tweets", []) if data else []

# Usage
client = TwitterAPIClient("your_api_key_here")

# Get user info
user = client.get_user_info("elonmusk")
print(f"User: {user['data']['name']}, Followers: {user['data']['followers']}")

# Get recent tweets
tweets = client.get_user_tweets("elonmusk", limit=50)
print(f"Retrieved {len(tweets)} tweets")

# Search tweets
results = client.search_tweets('"AI" from:elonmusk since:2024-01-01')
print(f"Found {len(results.get('tweets', []))} tweets")
```

### Go Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
)

type TwitterAPIClient struct {
    APIKey  string
    BaseURL string
    Client  *http.Client
}

type UserResponse struct {
    Data struct {
        UserName  string `json:"userName"`
        Name      string `json:"name"`
        Followers int    `json:"followers"`
        Following int    `json:"following"`
    } `json:"data"`
    Status string `json:"status"`
}

func NewTwitterAPIClient(apiKey string) *TwitterAPIClient {
    return &TwitterAPIClient{
        APIKey:  apiKey,
        BaseURL: "https://api.twitterapi.io",
        Client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *TwitterAPIClient) makeRequest(method, endpoint string, params url.Values) ([]byte, error) {
    reqURL := c.BaseURL + endpoint
    if len(params) > 0 {
        reqURL += "?" + params.Encode()
    }
    
    req, err := http.NewRequest(method, reqURL, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Add("x-api-key", c.APIKey)
    
    resp, err := c.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %d", resp.StatusCode)
    }
    
    return io.ReadAll(resp.Body)
}

func (c *TwitterAPIClient) GetUserInfo(username string) (*UserResponse, error) {
    params := url.Values{}
    params.Add("userName", username)
    
    body, err := c.makeRequest("GET", "/twitter/user/info", params)
    if err != nil {
        return nil, err
    }
    
    var userResp UserResponse
    if err := json.Unmarshal(body, &userResp); err != nil {
        return nil, err
    }
    
    return &userResp, nil
}

func main() {
    client := NewTwitterAPIClient("your_api_key_here")
    
    user, err := client.GetUserInfo("elonmusk")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("User: %s (@%s)\n", user.Data.Name, user.Data.UserName)
    fmt.Printf("Followers: %d, Following: %d\n", 
        user.Data.Followers, user.Data.Following)
}
```

### Node.js Complete Example

```javascript
const axios = require('axios');

class TwitterAPIClient {
    constructor(apiKey) {
        this.apiKey = apiKey;
        this.baseURL = 'https://api.twitterapi.io';
        this.headers = { 'x-api-key': apiKey };
    }

    async makeRequest(method, endpoint, params = {}) {
        try {
            const response = await axios({
                method,
                url: `${this.baseURL}${endpoint}`,
                headers: this.headers,
                params,
                timeout: 30000
            });
            return response.data;
        } catch (error) {
            if (error.response?.status === 429) {
                console.log('Rate limited. Please wait...');
                await new Promise(resolve => setTimeout(resolve, 2000));
                return this.makeRequest(method, endpoint, params);
            }
            throw error;
        }
    }

    async getUserInfo(username) {
        return this.makeRequest('GET', '/twitter/user/info', { userName: username });
    }

    async getUserTweets(username, limit = 100) {
        const tweets = [];
        let cursor = '';

        while (tweets.length < limit) {
            const data = await this.makeRequest('GET', '/twitter/user/last_tweets', {
                userName: username,
                cursor
            });

            tweets.push(...(data.tweets || []));

            if (!data.has_next_page) break;
            cursor = data.next_cursor;
        }

        return tweets.slice(0, limit);
    }

    async searchTweets(query, queryType = 'Latest') {
        return this.makeRequest('GET', '/twitter/tweet/advanced_search', {
            query,
            queryType
        });
    }
}

// Usage
(async () => {
    const client = new TwitterAPIClient('your_api_key_here');

    try {
        // Get user info
        const user = await client.getUserInfo('elonmusk');
        console.log(`User: ${user.data.name}, Followers: ${user.data.followers}`);

        // Get tweets
        const tweets = await client.getUserTweets('elonmusk', 50);
        console.log(`Retrieved ${tweets.length} tweets`);

        // Search
        const results = await client.searchTweets('"AI" from:elonmusk');
        console.log(`Found ${results.tweets?.length || 0} tweets`);
    } catch (error) {
        console.error('Error:', error.message);
    }
})();
```

## 11. Security Best Practices

1. **API Key Protection**
   - Never commit API keys to version control
   - Use environment variables or secure vaults
   - Rotate keys periodically

2. **Proxy Security**
   - Use reputable proxy providers
   - Avoid free proxies (security risk)
   - Rotate proxies for action endpoints

3. **Login Credentials**
   - Store login cookies securely
   - Implement cookie expiration handling
   - Never log sensitive credentials

4. **Rate Limiting**
   - Implement client-side rate limiting
   - Monitor usage to avoid unexpected charges
   - Use exponential backoff for retries

## 12. Performance Optimization

1. **Batch Requests**: Use batch endpoints when fetching multiple resources
2. **Caching**: Cache user profiles and tweets that don't change frequently
3. **Pagination**: Only fetch what you need, don't over-paginate
4. **Concurrent Requests**: Use async/parallel requests where appropriate (respect rate limits)
5. **Connection Pooling**: Reuse HTTP connections for better performance

## 13. Support and Resources

- **Documentation**: https://docs.twitterapi.io
- **Dashboard**: https://twitterapi.io/dashboard
- **Advanced Search Guide**: https://github.com/igorbrigadir/twitter-advanced-search
- **Proxy Provider**: https://www.webshare.io/?referral_code=4e0q1n00a504

---

**Document Version**: 1.0  
**Last Updated**: November 18, 2025  
**Source**: Based on https://docs.twitterapi.io (accessed November 18, 2025)

**Note**: This documentation is comprehensive and self-contained for AI agents to build applications using the twitterapi.io API. All endpoints, parameters, and examples are based on the official documentation. For the most current information, always refer to the official documentation at https://docs.twitterapi.io.
