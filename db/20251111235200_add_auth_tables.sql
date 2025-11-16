-- migration: add authentication tables for OAuth flow
-- timestamp: 2025-11-11 23:52:00 utc
-- purpose: creates tables for OAuth state management, user sessions, and encrypted OAuth tokens
-- includes: table definitions, constraints, indexes, and row-level security policies

-- create user-scoped table: oauth_state
-- stores PKCE state and CSRF tokens during OAuth flow
create table if not exists oauth_state (
    state_token text primary key,
    user_id uuid, -- null for anonymous OAuth initiation
    code_verifier text not null, -- PKCE code verifier
    code_challenge text not null, -- PKCE code challenge
    redirect_uri text not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null
);
-- create index for oauth_state cleanup
create index if not exists idx_oauth_state_expires on oauth_state (expires_at);

-- create user-scoped table: user_sessions
-- stores active user sessions with encrypted tokens
create table if not exists user_sessions (
    session_token text primary key,
    user_id uuid not null,
    x_user_id text not null, -- X (Twitter) user ID
    x_handle text not null,
    x_display_name text,
    encrypted_access_token text not null, -- AES-256-GCM encrypted
    encrypted_refresh_token text, -- AES-256-GCM encrypted (nullable)
    access_token_expires_at timestamptz,
    refresh_token_expires_at timestamptz,
    authenticated_at timestamptz not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,
    constraint chk_sessions_expires check (expires_at > created_at)
);
-- create indexes for user_sessions
create index if not exists idx_user_sessions_user on user_sessions (user_id);
create index if not exists idx_user_sessions_expires on user_sessions (expires_at);

-- create user-scoped table: user_oauth_tokens
-- stores encrypted OAuth tokens separately for security
create table if not exists user_oauth_tokens (
    user_id uuid primary key,
    x_user_id text not null,
    x_handle text not null,
    x_display_name text,
    encrypted_access_token text not null,
    encrypted_refresh_token text,
    access_token_expires_at timestamptz,
    refresh_token_expires_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
-- create index for user_oauth_tokens lookup
create index if not exists idx_user_oauth_tokens_x_user on user_oauth_tokens (x_user_id);

---------------------------------------------------------------
-- enable row-level security and create policies for auth tables
---------------------------------------------------------------

-- oauth_state rls (users can only access their own state)
alter table oauth_state enable row level security;
create policy user_isolation_oauth_state on oauth_state
    using (user_id = current_setting('app.user_id', true)::uuid or user_id is null);

-- user_sessions rls
alter table user_sessions enable row level security;
create policy user_isolation_user_sessions on user_sessions
    using (user_id = current_setting('app.user_id', true)::uuid);

-- user_oauth_tokens rls
alter table user_oauth_tokens enable row level security;
create policy user_isolation_user_oauth_tokens on user_oauth_tokens
    using (user_id = current_setting('app.user_id', true)::uuid);

-- end of migration
