# REST API Plan - Ask Your Feed (MVP)

## 1. Resources

### Core Resources
- **Session** - User authentication and session management (email/password + X username)
- **QA** - Question and Answer messages (maps to `qa_messages` and `qa_sources` tables)
- **Posts** - User's feed posts (maps to `posts` table)
- **Following** - Authors the user follows (maps to `user_following` table)
- **Ingest** - Feed ingestion runs and status (maps to `ingest_runs` table)

---

## 2. Endpoints

### 2.1. Authentication & Session Management

#### POST /api/v1/auth/register
Register a new user account.

**Description:** Creates a new user account with email, password, and X username. Validates X username via twitterapi.io before completing registration.

**Request Payload:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "password_confirmation": "SecurePassword123!",
  "x_username": "elonmusk"
}
```

**Validation:**
- `email`: Required, valid email format, unique
- `password`: Required, min 8 characters, must contain uppercase, lowercase, number, special character
- `password_confirmation`: Required, must match password
- `x_username`: Required, validated via `/twitter/user/info` endpoint from twitterapi.io

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "x_username": "elonmusk",
  "x_display_name": "Elon Musk",
  "created_at": "2025-10-31T18:00:00Z",
  "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Success:** 201 Created  
**Error Codes:**
- 400 Bad Request - Invalid input data or validation error
- 409 Conflict - Email already registered
- 422 Unprocessable Entity - X username does not exist (verified via twitterapi.io)
- 500 Internal Server Error - Registration system error
- 503 Service Unavailable - twitterapi.io temporarily unavailable

---

#### POST /api/v1/auth/login
Authenticate user with email and password.

**Description:** Authenticates user and creates a new session.

**Request Payload:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "x_username": "elonmusk",
  "x_display_name": "Elon Musk",
  "session_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "session_expires_at": "2025-11-07T18:00:00Z"
}
```

**Success:** 200 OK  
**Error Codes:**
- 400 Bad Request - Invalid input data
- 401 Unauthorized - Invalid email or password
- 500 Internal Server Error - Authentication system error

---

#### POST /api/v1/auth/logout
Terminate user session.

**Description:** Invalidates session token and clears authentication state.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Request Payload:** None

**Response:**
```json
{
  "message": "Wylogowano pomyślnie"
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session

---

#### GET /api/v1/session/current
Get current user session information.

**Description:** Returns current authenticated user details and session metadata.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Response:**
```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "x_username": "elonmusk",
  "x_display_name": "Elon Musk",
  "authenticated_at": "2025-10-31T18:00:00Z",
  "session_expires_at": "2025-11-07T18:00:00Z",
  "following_count": 150,
  "following_limit": 150
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session

---

### 2.2. Feed Ingestion

#### POST /api/v1/ingest/trigger
Manually trigger feed ingestion.

**Description:** Initiates an immediate feed ingestion run. Useful for initial backfill (24h) or manual refresh. Uses twitterapi.io to fetch followings and tweets.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Request Payload:**
```json
{
  "backfill_hours": 24
}
```

**Response:**
```json
{
  "ingest_run_id": "01HQKD8YJXM5R3QW9VKZT2BNCP",
  "status": "started",
  "started_at": "2025-10-31T18:00:00Z"
}
```

**Success:** 202 Accepted  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 403 Forbidden - Budget exhausted
- 409 Conflict - Ingest already in progress
- 429 Too Many Requests - Rate limit exceeded
- 500 Internal Server Error - Ingestion system error

---

#### GET /api/v1/ingest/status
Get ingestion status and history.

**Description:** Returns current and recent ingestion run information, including last sync time and following count.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Query Parameters:**
- `limit` (optional, integer, default: 10, max: 50) - Number of recent runs to return

**Response:**
```json
{
  "last_sync_at": "2025-10-31T17:45:00Z",
  "next_sync_at": "2025-10-31T21:45:00Z",
  "following_count": 150,
  "following_limit": 150,
  "current_run": {
    "id": "01HQKD8YJXM5R3QW9VKZT2BNCP",
    "status": "running",
    "started_at": "2025-10-31T18:00:00Z",
    "fetched_count": 42,
    "rate_limit_hits": 0
  },
  "recent_runs": [
    {
      "id": "01HQKD7ZMKR4P2NV8TJYS1AMBO",
      "status": "ok",
      "started_at": "2025-10-31T17:45:00Z",
      "completed_at": "2025-10-31T17:46:30Z",
      "fetched_count": 15,
      "retried": 0,
      "rate_limit_hits": 0
    },
    {
      "id": "01HQKD6XLJQ3O1MT7SIHWR0ZAN",
      "status": "rate_limited",
      "started_at": "2025-10-31T17:30:00Z",
      "completed_at": "2025-10-31T17:32:15Z",
      "fetched_count": 8,
      "retried": 2,
      "rate_limit_hits": 3,
      "error": "Przekroczono limit żądań API twitterapi.io"
    }
  ]
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 500 Internal Server Error - Database error

---

### 2.3. Question & Answer (Q&A)

#### POST /api/v1/qa
Create a new Q&A interaction.

**Description:** Submits a question to LLM and returns an answer based on posts from the specified date range.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Request Payload:**
```json
{
  "question": "Jakie były główne tematy dyskusji w tym tygodniu?",
  "date_from": "2025-10-24T00:00:00Z",
  "date_to": "2025-10-31T23:59:59Z"
}
```

**Validation:**
- `question`: Required, non-empty string, max 2000 characters
- `date_from`: Optional, defaults to 24 hours ago, must be valid ISO 8601 timestamp
- `date_to`: Optional, defaults to now, must be valid ISO 8601 timestamp
- `date_from` must be <= `date_to`

**Response:**
```json
{
  "id": "01HQKD9PKNL6S4RX0WLZU3CODR",
  "question": "Jakie były główne tematy dyskusji w tym tygodniu?",
  "answer": "• Główne tematy obejmowały rozwój sztucznej inteligencji...\n• Dyskutowano również o nowych regulacjach...\n• Pojawiły się pytania dotyczące...",
  "date_from": "2025-10-24T00:00:00Z",
  "date_to": "2025-10-31T23:59:59Z",
  "created_at": "2025-10-31T18:00:00Z",
  "sources": [
    {
      "x_post_id": 1234567890123456,
      "author_handle": "@author1",
      "author_display_name": "Author One",
      "published_at": "2025-10-30T14:30:00Z",
      "url": "https://twitter.com/author1/status/1234567890123456",
      "text_preview": "To jest fragment tekstu postu..."
    },
    {
      "x_post_id": 1234567890123457,
      "author_handle": "@author2",
      "author_display_name": "Author Two",
      "published_at": "2025-10-29T10:15:00Z",
      "url": "https://twitter.com/author2/status/1234567890123457",
      "text_preview": "Inny fragment dyskusji..."
    },
    {
      "x_post_id": 1234567890123458,
      "author_handle": "@author3",
      "author_display_name": "Author Three",
      "published_at": "2025-10-28T16:45:00Z",
      "url": "https://twitter.com/author3/status/1234567890123458",
      "text_preview": "Kolejny interesujący punkt..."
    }
  ]
}
```

**Response (No Content Found):**
```json
{
  "id": "01HQKD9PKNL6S4RX0WLZU3CODR",
  "question": "Jakie były główne tematy dyskusji w tym tygodniu?",
  "answer": "Brak treści w wybranym zakresie dat. Spróbuj rozszerzyć zakres dat lub sprawdź, czy posty zostały poprawnie zaimportowane.",
  "date_from": "2025-10-24T00:00:00Z",
  "date_to": "2025-10-31T23:59:59Z",
  "created_at": "2025-10-31T18:00:00Z",
  "sources": []
}
```

**Success:** 201 Created  
**Error Codes:**
- 400 Bad Request - Invalid parameters or validation error
- 401 Unauthorized - Invalid or expired session
- 403 Forbidden - Budget exhausted
- 422 Unprocessable Entity - date_from > date_to
- 429 Too Many Requests - Rate limit exceeded
- 500 Internal Server Error - LLM service error
- 503 Service Unavailable - LLM service temporarily unavailable

---

#### GET /api/v1/qa
List Q&A history with pagination.

**Description:** Returns paginated list of user's Q&A interactions, ordered by creation time (newest first).

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Query Parameters:**
- `limit` (optional, integer, default: 20, max: 100) - Number of items per page
- `cursor` (optional, string) - Pagination cursor (ULID of last item from previous page)

**Response:**
```json
{
  "items": [
    {
      "id": "01HQKD9PKNL6S4RX0WLZU3CODR",
      "question": "Jakie były główne tematy dyskusji w tym tygodniu?",
      "answer_preview": "• Główne tematy obejmowały rozwój sztucznej inteligencji...",
      "date_from": "2025-10-24T00:00:00Z",
      "date_to": "2025-10-31T23:59:59Z",
      "created_at": "2025-10-31T18:00:00Z",
      "sources_count": 3
    },
    {
      "id": "01HQKD8RMJK5R3PV9UKYS2BMBN",
      "question": "Co mówiono o AI?",
      "answer_preview": "• Sztuczna inteligencja była szeroko dyskutowana...",
      "date_from": "2025-10-30T00:00:00Z",
      "date_to": "2025-10-31T23:59:59Z",
      "created_at": "2025-10-31T12:30:00Z",
      "sources_count": 5
    }
  ],
  "next_cursor": "01HQKD8RMJK5R3PV9UKYS2BMBN",
  "has_more": true
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 500 Internal Server Error - Database error

---

#### GET /api/v1/qa/{id}
Get Q&A details by ID.

**Description:** Returns full details of a specific Q&A interaction including complete answer and all sources.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**URL Parameters:**
- `id` (required, string) - Q&A message ULID

**Response:**
```json
{
  "id": "01HQKD9PKNL6S4RX0WLZU3CODR",
  "question": "Jakie były główne tematy dyskusji w tym tygodniu?",
  "answer": "• Główne tematy obejmowały rozwój sztucznej inteligencji, w szczególności nowe modele językowe i ich zastosowania w różnych branżach.\n• Dyskutowano również o nowych regulacjach dotyczących ochrony danych osobowych i ich wpływie na branżę technologiczną.\n• Pojawiły się pytania dotyczące przyszłości pracy zdalnej i hybrydowej w kontekście rosnącej automatyzacji.\n• Wiele uwagi poświęcono również kwestiom związanym z bezpieczeństwem cybernetycznym.",
  "date_from": "2025-10-24T00:00:00Z",
  "date_to": "2025-10-31T23:59:59Z",
  "created_at": "2025-10-31T18:00:00Z",
  "sources": [
    {
      "x_post_id": 1234567890123456,
      "author_handle": "@author1",
      "author_display_name": "Author One",
      "published_at": "2025-10-30T14:30:00Z",
      "url": "https://twitter.com/author1/status/1234567890123456",
      "text": "Pełna treść pierwszego postu źródłowego..."
    },
    {
      "x_post_id": 1234567890123457,
      "author_handle": "@author2",
      "author_display_name": "Author Two",
      "published_at": "2025-10-29T10:15:00Z",
      "url": "https://twitter.com/author2/status/1234567890123457",
      "text": "Pełna treść drugiego postu źródłowego..."
    },
    {
      "x_post_id": 1234567890123458,
      "author_handle": "@author3",
      "author_display_name": "Author Three",
      "published_at": "2025-10-28T16:45:00Z",
      "url": "https://twitter.com/author3/status/1234567890123458",
      "text": "Pełna treść trzeciego postu źródłowego..."
    }
  ]
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 404 Not Found - Q&A with specified ID not found or doesn't belong to user
- 500 Internal Server Error - Database error

---

#### DELETE /api/v1/qa/{id}
Delete a specific Q&A interaction.

**Description:** Permanently removes a Q&A message and its associated sources from history.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**URL Parameters:**
- `id` (required, string) - Q&A message ULID

**Response:**
```json
{
  "message": "Q&A usunięte pomyślnie"
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 404 Not Found - Q&A with specified ID not found or doesn't belong to user
- 500 Internal Server Error - Database error

---

#### DELETE /api/v1/qa
Delete all Q&A history.

**Description:** Permanently removes all Q&A messages and sources for the authenticated user.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Response:**
```json
{
  "message": "Cała historia Q&A została usunięta",
  "deleted_count": 42
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 500 Internal Server Error - Database error

---

### 2.4. Following

#### GET /api/v1/following
Get list of authors the user follows.

**Description:** Returns paginated list of X authors that the user is following and whose posts are being ingested.

**Request Headers:**
- `Authorization: Bearer <session_token>` (required)

**Query Parameters:**
- `limit` (optional, integer, default: 50, max: 200) - Number of items per page
- `cursor` (optional, integer) - Pagination cursor (x_author_id of last item)

**Response:**
```json
{
  "items": [
    {
      "x_author_id": 123456789,
      "handle": "@author1",
      "display_name": "Author One",
      "last_seen_at": "2025-10-31T15:30:00Z",
      "last_checked_at": "2025-10-31T17:45:00Z"
    },
    {
      "x_author_id": 987654321,
      "handle": "@author2",
      "display_name": "Author Two",
      "last_seen_at": "2025-10-31T14:20:00Z",
      "last_checked_at": "2025-10-31T17:45:00Z"
    }
  ],
  "next_cursor": "987654321",
  "has_more": true,
  "total_count": 150,
  "limit": 150
}
```

**Success:** 200 OK  
**Error Codes:**
- 401 Unauthorized - Invalid or expired session
- 500 Internal Server Error - Database error

---

### 2.5. System

#### GET /api/v1/system/health
Get system health status.

**Description:** Returns health status of the application and its dependencies. Does not require authentication.

**Request Headers:** None required

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-31T18:00:00Z",
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5
    },
    "llm_service": {
      "status": "healthy",
      "response_time_ms": 120
    },
    "twitterapi_io": {
      "status": "healthy",
      "response_time_ms": 700
    }
  }
}
```

**Response (Degraded):**
```json
{
  "status": "degraded",
  "timestamp": "2025-10-31T18:00:00Z",
  "version": "1.0.0",
  "components": {
    "database": {
      "status": "healthy",
      "response_time_ms": 5
    },
    "llm_service": {
      "status": "unhealthy",
      "error": "Connection timeout"
    },
    "twitterapi_io": {
      "status": "rate_limited",
      "error": "429 Too Many Requests"
    }
  }
}
```

**Success:** 200 OK (healthy), 503 Service Unavailable (degraded/unhealthy)

---

## 3. Authentication and Authorization

### 3.1. Email/Password Authentication with X Username

**Flow:** Standard email/password registration and login

**Implementation:**
1. User registers via `POST /api/v1/auth/register` with email, password, and X username
2. Server validates X username via twitterapi.io `/twitter/user/info?userName={username}`
3. If username exists, server creates user account with hashed password (bcrypt/argon2)
4. Server stores X username and display name from twitterapi.io response
5. Server creates session and returns JWT session token
6. For login, user provides email and password via `POST /api/v1/auth/login`
7. Server validates credentials and returns JWT session token

**Session Management:**
- Session tokens implemented as JWT (JSON Web Tokens)
- Session duration: 7 days
- Tokens contain: user_id, email, x_username, issued_at, expires_at
- User ID set via PostgreSQL `SET LOCAL app.user_id` for RLS enforcement

**Password Security:**
- Passwords hashed using bcrypt or argon2
- Minimum password requirements enforced (8 chars, uppercase, lowercase, number, special char)
- Password reset functionality (future enhancement, not in MVP)

**X Username Validation:**
- Validated during registration via twitterapi.io endpoint: `GET /twitter/user/info?userName={username}`
- If user not found (400 error), registration fails with 422 error
- Display name stored from API response for UI purposes

---

### 3.2. Request Authentication

All API endpoints (except `/api/v1/auth/register`, `/api/v1/auth/login`, and `/api/v1/system/health`) require authentication via session token.

**Authentication Header:**
```
Authorization: Bearer <session_token>
```

Or via HTTP-only cookie (preferred for browser requests):
```
Cookie: session_token=<session_token>
```

**Authentication Flow:**
1. Extract session token from Authorization header or cookie
2. Validate JWT signature and expiration
3. Extract user_id from token claims
4. Set PostgreSQL session variable: `SET LOCAL app.user_id = '<user_id>'`
5. Execute request with RLS policies enforced

**Authorization:**
- Row-Level Security (RLS) enforced at database level
- All user-scoped tables filtered by `user_id = current_setting('app.user_id')`
- Global tables (`authors`) accessible to all authenticated users
- Budget checks performed before expensive operations (Q&A, ingestion)

---

## 4. Validation and Business Logic

### 4.1. Input Validation

#### General Rules
- All string inputs sanitized to prevent XSS and SQL injection
- Maximum request payload size: 10 MB
- Request timeout: 30 seconds (120 seconds for Q&A endpoint)
- Content-Type validation: `application/json` required for POST/PUT requests

#### Field-Specific Validation

**Registration Fields:**
- `email`: Required, valid email format, unique in database
- `password`: Required, min 8 characters, must contain uppercase, lowercase, number, special character
- `password_confirmation`: Required, must match password
- `x_username`: Required, alphanumeric + underscore, max 15 characters, validated via twitterapi.io

**Q&A Question:**
- Required: Yes
- Type: String
- Min length: 1 character
- Max length: 2000 characters
- Trimmed of leading/trailing whitespace
- Must not be only whitespace

**Date Range (date_from/date_to):**
- Type: ISO 8601 timestamp with timezone
- date_from must be <= date_to
- Maximum range: 90 days
- Future dates rejected

**Pagination:**
- limit: Integer, 1 to resource-specific maximum
- cursor: ULID or resource-specific ID format
- Invalid cursor returns 400 Bad Request

**Post URL Validation:**
- Must match regex: `^https?://(x|twitter)\.com/.+/status/\d+$`
- Enforced at database level via CHECK constraint

**Author IDs and Post IDs:**
- Must be positive BIGINT (> 0)
- Enforced at database level via CHECK constraints

**Ingest Status:**
- Must be one of: 'ok', 'rate_limited', 'error'
- Enforced at database level via CHECK constraint

---

### 4.2. Business Logic Implementation

#### Feed Ingestion Logic
1. **Scheduled Execution:** Runs every 4 hours ± 15 minutes (jitter)
2. **Following Fetch:** Uses twitterapi.io `/twitter/user/followings?userName={username}` to get list of followed users (max 150)
3. **Tweet Fetch:** For each followed user, calls `/twitter/user/last_tweets?userName={followed_username}`
4. **Filtering:**
   - Only original posts: `isReply === false` OR (`isReply === true` AND `inReplyToUserId === author.id`)
   - No `retweeted_tweet` field (exclude retweets)
   - No `quoted_tweet` field (exclude quote tweets)
   - Self-reply threads are allowed
5. **Pagination:**
   - Regular ingest: Only first page (~20 tweets per followed user)
   - Backfill: Paginate using `cursor` and `has_next_page` until 24h reached or no more pages
6. **Temporal Filtering:** Filter by `createdAt` field (no `since_id` parameter in twitterapi.io)
7. **Media Processing:**
   - Images: Max 4 per post, converted to text descriptions via OpenRouter
   - Videos: Max 90 seconds or 25 MB, transcribed to text
   - Media content merged into `posts.text` field
8. **Rate Limiting:**
   - Exponential backoff on 429 responses from twitterapi.io
   - Maximum 3 retries per request
   - Tracking: `rate_limit_hits` and `retried` fields
9. **Author Updates:**
   - Update `authors.last_seen_at` on post ingestion
   - Update `user_following.last_checked_at` on author check
10. **Full-Text Index:**
    - `posts.ts` updated via trigger using Polish + English dictionaries
    - Unaccent applied for diacritic-insensitive search

#### Q&A Logic
1. **Date Range:** Default to last 24 hours if not specified
2. **LLM Prompting:**
   - System prompt: Feed-only context, no web browsing
   - Include post content in chronological order
   - Request structured response with bullet points
3. **Source Selection:**
   - Minimum 3 sources if available
   - All sources if < 3 available
   - Sources linked via `qa_sources` junction table
4. **No Content Handling:**
   - Return specific message suggesting date range expansion
   - Empty sources array
   - Still create Q&A record for history
5. **Source URL Format:** `https://twitter.com/{userName}/status/{tweetId}` (from twitterapi.io Tweet object)

#### History Management
1. **Pagination:** Cursor-based using ULID ordering
2. **Soft Limits:** No automatic deletion/retention in MVP
3. **Bulk Delete:** Single transaction for "Delete All" operation

#### Security Logic
1. **Password Hashing:**
   - bcrypt or argon2 for password hashing
   - Salt rounds: 12 (bcrypt) or appropriate work factor (argon2)
   - Passwords never stored in plain text
2. **Session Management:**
   - JWT tokens with HS256 signing
   - Secret key stored in environment variables
   - Token expiry: 7 days
3. **Session Expiry:**
   - Sessions expire after 7 days
   - No automatic refresh in MVP
   - User must re-authenticate after expiry
4. **RLS Enforcement:**
   - `SET LOCAL app.user_id` on every request
   - Automatic filtering at database level
   - No application-level user filtering needed

---

### 4.3. Error Handling

#### Standard Error Response Format
```json
{
  "error": {
    "code": "NO_CONTENT_FOUND",
    "message": "Brak treści w wybranym zakresie dat. Spróbuj rozszerzyć zakres dat.",
    "details": {
        "date_from": "2025-10-24T00:00:00Z",
        "date_to": "2025-10-31T23:59:59Z"
    }
  }
}
```

#### Error Codes and Messages (Polish)

**Authentication Errors:**
- `UNAUTHORIZED` - "Brak autoryzacji. Zaloguj się ponownie."
- `SESSION_EXPIRED` - "Sesja wygasła. Zaloguj się ponownie."
- `INVALID_TOKEN` - "Nieprawidłowy token autoryzacji."
- `INVALID_CREDENTIALS` - "Nieprawidłowy email lub hasło."
- `EMAIL_ALREADY_EXISTS` - "Konto z tym adresem email już istnieje."
- `X_USERNAME_NOT_FOUND` - "Konto X o nazwie '{username}' nie istnieje. Sprawdź poprawność nazwy użytkownika."

**Validation Errors:**
- `INVALID_INPUT` - "Nieprawidłowe dane wejściowe."
- `INVALID_DATE_RANGE` - "Nieprawidłowy zakres dat. Data początkowa nie może być późniejsza niż końcowa."
- `QUESTION_TOO_LONG` - "Pytanie przekracza maksymalną długość 2000 znaków."
- `QUESTION_REQUIRED` - "Pytanie jest wymagane."
- `PASSWORD_TOO_WEAK` - "Hasło musi zawierać minimum 8 znaków, w tym wielką literę, małą literę, cyfrę i znak specjalny."
- `EMAIL_INVALID` - "Nieprawidłowy format adresu email."

**Business Logic Errors:**
- `NO_CONTENT_FOUND` - "Brak treści w wybranym zakresie dat. Spróbuj rozszerzyć zakres dat."
- `RATE_LIMIT_EXCEEDED` - "Przekroczono limit żądań. Spróbuj ponownie za {retry_after} sekund."
- `INGEST_IN_PROGRESS` - "Ingest jest już w toku. Poczekaj na zakończenie obecnego procesu."

**System Errors:**
- `DATABASE_ERROR` - "Błąd bazy danych. Spróbuj ponownie później."
- `LLM_SERVICE_ERROR` - "Błąd usługi LLM. Spróbuj ponownie później."
- `TWITTERAPI_IO_ERROR` - "Błąd komunikacji z API twitterapi.io. Spróbuj ponownie później."
- `TWITTERAPI_IO_RATE_LIMITED` - "Przekroczono limit żądań API twitterapi.io. Spróbuj ponownie później."
- `INTERNAL_ERROR` - "Wystąpił błąd wewnętrzny. Spróbuj ponownie później."

**Resource Errors:**
- `NOT_FOUND` - "Zasób nie został znaleziony."
- `ALREADY_EXISTS` - "Zasób już istnieje."

---

### 4.4. Observability

**Logging:**
- Structured JSON logs
- No PII in logs (user IDs hashed if logged)
- Log levels: DEBUG, INFO, WARN, ERROR
- Request ID correlation across all logs

**Metrics (OpenTelemetry):**
- Request latency (p50, p95, p99)
- Request rate by endpoint
- Error rate by endpoint and error type
- Ingest run statistics (fetched count, retries, rate limit hits)
- LLM call latency and token usage
- Database query performance
- twitterapi.io API call statistics (request count, response time, error rate)

**Tracing (OpenTelemetry):**
- Distributed traces across all service boundaries
- Span annotations for key operations
- DB queries traced with sanitized query text
- External API calls traced (twitterapi.io, LLM, OpenRouter)

**Key SLIs:**
- Availability: 99.5% uptime
- Request latency: p95 < 500ms (excluding Q&A)
- Q&A latency: p95 < 10s
- Ingest freshness: Average post-to-visibility < 4.5 hours

---

## 5. API Versioning

**Versioning Strategy:** URL path versioning

**Current Version:** v1

**Version in URL:** `/api/v1/...`

---

## 6. CORS and Security Headers

**CORS Configuration:**
```
Access-Control-Allow-Origin: https://app.domain.com
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

**Security Headers:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Content-Security-Policy: default-src 'self'
Referrer-Policy: strict-origin-when-cross-origin
```

---

## 7. twitterapi.io Integration

### 7.1. API Configuration

**Base URL:** `https://api.twitterapi.io`

**Authentication:**
- Header: `x-api-key: {API_KEY}`
- API key stored in environment variables or secrets manager
- Single API key shared across all users (application-level authentication)

**Pricing:**
- $0.15 per 1,000 tweets
- $0.18 per 1,000 user profiles
- $0.15 per 1,000 followers
- Minimum charge: $0.00015 per request

**Rate Limits:**
- Maximum 200 QPS (queries per second) per client
- Average response time: 700ms

### 7.2. Endpoints Used

| Endpoint | Purpose | Frequency | Cost Impact |
|----------|---------|-----------|-------------|
| `/twitter/user/info` | Validate X username during registration | Once per registration | $0.00018 per user |
| `/twitter/user/followings` | Fetch list of followed users | Every 4h per user | $0.00015 per ingest |
| `/twitter/user/last_tweets` | Fetch tweets for each followed user | Every 4h × 150 users | $0.15 per 1k tweets |

### 7.3. Cost Estimation

**Per User Per Day:**
- 6 ingest runs (every 4h)
- 150 followed users
- ~20 tweets per user per ingest (first page only)
- Total: 6 × 150 × 20 = 18,000 tweets/day
- Cost: 18,000 / 1,000 × $0.15 = $2.70/day/user
- Monthly: ~$81/user (tweets only)

**Additional Costs:**
- User profile fetches: 150 profiles × $0.18/1k = $0.027 per backfill
- OpenRouter for media processing (separate service)
- LLM API for Q&A (separate service)

### 7.4. Error Handling

**429 Rate Limit:**
- Implement exponential backoff (2^attempt × base_delay)
- Maximum 3 retries per request
- Track rate_limit_hits in ingest_runs table
- Log warning and continue with next user

**400 Bad Request:**
- User not found during registration: Return 422 to user
- Invalid parameters: Log error and skip
- Malformed response: Log error and retry

**5xx Server Errors:**
- Retry with exponential backoff
- Maximum 3 retries
- If all retries fail, mark ingest run as 'error'

**Network Timeouts:**
- Request timeout: 30 seconds
- Retry with exponential backoff
- Maximum 3 retries

### 7.5. Response Parsing

**User Object Mapping:**
```
twitterapi.io → Application
- userName → handle (with @ prefix)
- name → display_name
- id → x_author_id
- followers → (not stored, display only)
- following → (not stored, display only)
```

**Tweet Object Mapping:**
```
twitterapi.io → Application
- id → x_post_id
- text → text (merged with media descriptions)
- createdAt → published_at
- url → url
- author.userName → author handle
- author.name → author display_name
- likeCount, retweetCount, etc. → metadata (optional storage)
```

**Filtering Logic:**
```go
func isOriginalPost(tweet Tweet) bool {
    // Exclude replies (except self-replies)
    if tweet.IsReply && tweet.InReplyToUserId != tweet.Author.Id {
        return false
    }
    
    // Exclude retweets
    if tweet.RetweetedTweet != nil {
        return false
    }
    
    // Exclude quote tweets
    if tweet.QuotedTweet != nil {
        return false
    }
    
    return true
}
```

---

## 8. Assumptions and Design Decisions

1. **User Identity:** 1:1 mapping between application user and X account; no support for multiple X accounts per user in MVP
2. **Authentication Model:** Email/password authentication with X username validation (no OAuth with X)
3. **API Access:** Application uses single twitterapi.io API key for all users (not per-user authentication)
4. **Following Limit:** Maximum 150 followed users per user (hard limit in MVP)
5. **Ingest Frequency:** Every 4 hours ± 15 minutes (jitter for load distribution)
6. **Pagination Strategy:** 
   - Regular ingest: First page only (~20 tweets per user)
   - Backfill: Full pagination until 24h reached
7. **Media Limits:** Videos > 90s or > 25MB are skipped during ingestion without error
8. **Pagination:** Cursor-based pagination used for all list endpoints to ensure consistency during concurrent writes
9. **Default Date Range:** 24 hours used as default for Q&A when not specified
10. **Source Minimum:** Minimum 3 sources in Q&A response only applies when ≥3 posts available in date range
11. **Timeline View:** Alternative response formats (e.g., "timeline" command) implemented via prompt engineering, not separate endpoints
12. **Language:** All error messages, UI text in responses in Polish (pl-PL)
13. **Telemetry:** No user behavior analytics (DAU/WAU) in MVP; only operational telemetry
14. **No Soft Delete:** All delete operations are hard deletes
15. **Session Storage:** Session tokens as JWT (no Redis required for MVP)
16. **Ingestion Scheduler:** Runs as separate background service, not triggered via API except for manual `POST /api/v1/ingest/trigger`
17. **LLM Context:** Full post text (including media descriptions) sent to LLM; no summarization or truncation
18. **URL Format:** Source URLs use format `https://twitter.com/{userName}/status/{tweetId}` (from twitterapi.io)
19. **Temporal Filtering:** Uses `createdAt` field for filtering (no `since_id` parameter in twitterapi.io)
