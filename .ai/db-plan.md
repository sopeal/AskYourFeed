# Database Schema for Ask Your Feed (MVP)

This document outlines the comprehensive PostgreSQL database schema based on the PRD, planning session notes, and the specified tech stack.

---

## 1. List of Tables and Columns

### 1.1. authors (Global Table - No RLS)
- **x_author_id** BIGINT PRIMARY KEY CHECK (x_author_id > 0)
- **handle** TEXT NOT NULL
- **display_name** TEXT
- **last_seen_at** TIMESTAMPTZ

---

### 1.2. user_following (User-Scoped)
- **user_id** UUID NOT NULL
- **x_author_id** BIGINT NOT NULL  
  _-- References authors(x_author_id)_
- **last_checked_at** TIMESTAMPTZ  
- **Primary Key**: (user_id, x_author_id)

**Indexes:**
- BTREE on (user_id, last_checked_at)

---

### 1.3. posts (User-Scoped)
- **user_id** UUID NOT NULL
- **x_post_id** BIGINT NOT NULL CHECK (x_post_id > 0)
- **author_id** BIGINT NOT NULL  
  _-- Foreign Key referencing authors(x_author_id) ON DELETE CASCADE_
- **published_at** TIMESTAMPTZ NOT NULL
- **url** TEXT NOT NULL  
  _-- CHECK: url ~ '^https?://(x|twitter)\\.com/.+/status/\\d+'_
- **text** TEXT NOT NULL  
  _-- Contains full post content including media transcripts/summaries_
- **conversation_id** BIGINT NULL  
  _-- Threads are stored as separate rows; may be used for timeline view_
- **ingested_at** TIMESTAMPTZ NOT NULL
- **first_visible_at** TIMESTAMPTZ NOT NULL
- **edited_seen** BOOLEAN DEFAULT false
- **ts** TSVECTOR  
  _-- For full-text search (maintenance via trigger or generated column using combined English + Polish dictionaries and unaccent)_

**Primary Key:** (user_id, x_post_id)

**Indexes:**
- BTREE on (user_id, published_at DESC)
- BTREE on (user_id, conversation_id)
- GIN on (ts)

---

### 1.4. qa_messages (User-Scoped)
- **id** CHAR(26) PRIMARY KEY  
  _-- ULID in string format (alternatively, use appropriate ULID datatype)_
- **user_id** UUID NOT NULL
- **question** TEXT NOT NULL
- **answer** TEXT NOT NULL
- **date_from** TIMESTAMPTZ NOT NULL
- **date_to** TIMESTAMPTZ NOT NULL
- **created_at** TIMESTAMPTZ NOT NULL DEFAULT now()

**Indexes:**
- BTREE on (user_id, created_at DESC)

---

### 1.5. qa_sources (User-Scoped - Junction Table)
- **qa_id** CHAR(26) NOT NULL  
  _-- References qa_messages(id)_
- **user_id** UUID NOT NULL
- **x_post_id** BIGINT NOT NULL

**Primary Key:** (qa_id, x_post_id)

**Foreign Keys:**
- FOREIGN KEY (user_id, x_post_id) REFERENCES posts(user_id, x_post_id) ON DELETE CASCADE
- *(Optionally, for enhanced integrity, a FK on (qa_id, user_id) referencing qa_messages can be added if qa_messages includes user_id)*

**Indexes:**
- BTREE on (qa_id)
- BTREE on (user_id, x_post_id)

---

### 1.6. ingest_runs (User-Scoped)
- **id** CHAR(26) PRIMARY KEY  
  _-- ULID in string format (or use an alternative ULID type)_
- **user_id** UUID NOT NULL
- **started_at** TIMESTAMPTZ NOT NULL
- **completed_at** TIMESTAMPTZ
- **status** TEXT NOT NULL CHECK (status IN ('ok','rate_limited','error'))
- **since_id** BIGINT NOT NULL CHECK (since_id > 0)
- **fetched_count** INT NOT NULL
- **retried** INT NOT NULL
- **rate_limit_hits** INT NOT NULL
- **err_text** TEXT

**Indexes:**
- BTREE on (user_id, started_at DESC)

---

## 2. Relationships Between Tables

- **authors**: Global table containing author details; referenced by posts.
- **user_following**: Maps a user (user_id) to the authors (x_author_id) they follow.
- **posts**: Stores individual posts per user with a composite primary key (user_id, x_post_id); each post references an author.
- **qa_messages**: Records user Q&A interactions.
- **qa_sources**: Junction table linking QA messages to posts; represents a many-to-many relationship between qa_messages and posts.
- **ingest_runs**: Tracks feed ingestion runs per user.

---

## 3. Indexes

- **posts**:
  - BTREE on (user_id, published_at DESC)
  - BTREE on (user_id, conversation_id)
  - GIN on (ts)
- **user_following**:
  - BTREE on (user_id, last_checked_at)
- **qa_messages**:
  - BTREE on (user_id, created_at DESC)
- **qa_sources**:
  - BTREE on (qa_id)
  - BTREE on (user_id, x_post_id)
- **ingest_runs**:
  - BTREE on (user_id, started_at DESC)

---

## 4. PostgreSQL Row-Level Security (RLS) Policies

For all user-scoped tables (posts, qa_messages, qa_sources, ingest_runs, user_following), enforce strict isolation based on the current user. An example policy:

```sql
ALTER TABLE <table_name> ENABLE ROW LEVEL SECURITY;

CREATE POLICY user_isolation ON <table_name>
  USING (user_id = current_setting('app.user_id', true)::uuid);
```

Apply the above to:
- posts
- qa_messages
- qa_sources
- ingest_runs
- user_following

---

## 5. Additional Notes and Design Decisions

- **Identifier Types & Constraints:**  
  All X-specific IDs (e.g., x_post_id, since_id) are stored as BIGINT with a check ensuring positive values. ULIDs are used as primary keys for qa_messages and ingest_runs to benefit from time-sortability and simplified pagination.

- **Full-Text Search:**  
  The `ts` column in posts is utilized for full-text search. It is recommended to maintain this column via a generated column or trigger that uses a combination of English and Polish dictionaries along with the unaccent function.

- **Cascading Deletes:**  
  Hard deletes are enforced. Dependent records (e.g., entries in qa_sources) cascade on deletion, ensuring data consistency.

- **RLS Implementation:**  
  Row-level security policies ensure that user-scoped tables only reveal data matching the active user's ID as set via `SET LOCAL app.user_id`.

- **Media Handling:**  
  Media content (images and video transcripts/summaries) is merged into the `posts.text` field. There is no separate media table.

- **URL Constraints:**  
  The canonical URL in posts is validated with a CHECK constraint to match a pattern ensuring proper X/Twitter URL formats.

- **Token and Secret Management:**  
  Sensitive data such as tokens is expected to be stored externally or in an encrypted form if stored within the database. RLS policies further protect any such stored information.

- **Design Philosophy:**  
  The schema adheres to normalization (approximately 3NF) while supporting efficient querying, full-text search, and user isolation via RLS. This design provides a robust basis for handling user feeds, Q&A history, and ingestion tracking as demanded by the application requirements.

---

This schema is ready to serve as the basis for database migrations in the Ask Your Feed MVP.
