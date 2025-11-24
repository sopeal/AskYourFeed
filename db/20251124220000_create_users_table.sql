-- migration: create users table for authentication
-- timestamp: 2025-11-24 22:00:00 utc
-- purpose: creates users table for email/password authentication with x username
-- includes: table definition, constraints, indexes, and unique constraints

-- create users table
create table if not exists users (
    id uuid primary key default gen_random_uuid(),
    email text not null unique,
    password_hash text not null,
    x_username text not null,
    x_display_name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- create index on email for fast lookups during login
create index if not exists idx_users_email on users (email);

-- create index on x_username for validation checks
create index if not exists idx_users_x_username on users (x_username);

-- create sessions table for JWT token management
create table if not exists sessions (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references users(id) on delete cascade,
    token_hash text not null unique,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,
    revoked_at timestamptz
);

-- create index on user_id for session lookups
create index if not exists idx_sessions_user_id on sessions (user_id);

-- create index on token_hash for fast token validation
create index if not exists idx_sessions_token_hash on sessions (token_hash);

-- create index on expires_at for cleanup of expired sessions
create index if not exists idx_sessions_expires_at on sessions (expires_at);

-- end of migration
