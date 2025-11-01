-- migration: create secure schema for ask your feed mvp
-- timestamp: 2025-10-23 21:48:26 utc
-- purpose: creates tables for authors, user_following, posts, qa_messages, qa_sources, and ingest_runs.
-- includes: table definitions, constraints, indexes, and row-level security policies for user-scoped tables.
-- notes: all sql is in lowercase and includes ample comments for destructive operations where applicable.

-- create global table: authors (no row level security)
create table if not exists authors (
    x_author_id bigint primary key check (x_author_id > 0),
    handle text not null,
    display_name text,
    last_seen_at timestamptz
);

-- create user-scoped table: user_following
create table if not exists user_following (
    user_id uuid not null,
    x_author_id bigint not null,
    last_checked_at timestamptz,
    constraint pk_user_following primary key (user_id, x_author_id)
);
-- create index for user_following on (user_id, last_checked_at)
create index if not exists idx_user_following_user_last_checked on user_following (user_id, last_checked_at);

-- create user-scoped table: posts
create table if not exists posts (
    user_id uuid not null,
    x_post_id bigint not null check (x_post_id > 0),
    author_id bigint not null,
    published_at timestamptz not null,
    url text not null check (url ~ '^https?://(x|twitter)\\.com/.+/status/\\d+'),
    text text not null,
    conversation_id bigint,
    ingested_at timestamptz not null,
    first_visible_at timestamptz not null,
    edited_seen boolean default false,
    ts tsvector,
    constraint pk_posts primary key (user_id, x_post_id),
    constraint fk_posts_author foreign key (author_id) references authors (x_author_id) on delete cascade
);
-- create indexes for posts
create index if not exists idx_posts_user_published on posts (user_id, published_at desc);
create index if not exists idx_posts_user_conversation on posts (user_id, conversation_id);
create index if not exists idx_posts_ts on posts using gin (ts);

-- create user-scoped table: qa_messages
create table if not exists qa_messages (
    id char(26) primary key,
    user_id uuid not null,
    question text not null,
    answer text not null,
    date_from timestamptz not null,
    date_to timestamptz not null,
    created_at timestamptz not null default now()
);
-- create index for qa_messages on (user_id, created_at desc)
create index if not exists idx_qa_messages_user_created on qa_messages (user_id, created_at desc);

-- create user-scoped junction table: qa_sources
create table if not exists qa_sources (
    qa_id char(26) not null,
    user_id uuid not null,
    x_post_id bigint not null,
    constraint pk_qa_sources primary key (qa_id, x_post_id),
    constraint fk_qa_sources_posts foreign key (user_id, x_post_id) references posts (user_id, x_post_id) on delete cascade
);
-- create indexes for qa_sources
create index if not exists idx_qa_sources_qa on qa_sources (qa_id);
create index if not exists idx_qa_sources_user_post on qa_sources (user_id, x_post_id);

-- create user-scoped table: ingest_runs
create table if not exists ingest_runs (
    id char(26) primary key,
    user_id uuid not null,
    started_at timestamptz not null,
    completed_at timestamptz,
    status text not null check (status in ('ok','rate_limited','error')),
    since_id bigint not null check (since_id > 0),
    fetched_count int not null,
    retried int not null,
    rate_limit_hits int not null,
    err_text text
);
-- create index for ingest_runs on (user_id, started_at desc)
create index if not exists idx_ingest_runs_user_started on ingest_runs (user_id, started_at desc);

---------------------------------------------------------------
-- enable row-level security and create policies for user-scoped tables
---------------------------------------------------------------

-- user_following rls
alter table user_following enable row level security;
create policy user_isolation_user_following on user_following
    using (user_id = current_setting('app.user_id', true)::uuid);
    
-- posts rls
alter table posts enable row level security;
create policy user_isolation_posts on posts
    using (user_id = current_setting('app.user_id', true)::uuid);
    
-- qa_messages rls
alter table qa_messages enable row level security;
create policy user_isolation_qa_messages on qa_messages
    using (user_id = current_setting('app.user_id', true)::uuid);
    
-- qa_sources rls
alter table qa_sources enable row level security;
create policy user_isolation_qa_sources on qa_sources
    using (user_id = current_setting('app.user_id', true)::uuid);
    
-- ingest_runs rls
alter table ingest_runs enable row level security;
create policy user_isolation_ingest_runs on ingest_runs
    using (user_id = current_setting('app.user_id', true)::uuid);

-- end of migration
