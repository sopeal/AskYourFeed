-- migration: update ingest_runs table for cursor-based pagination
-- timestamp: 2025-11-20 21:53:56 utc
-- purpose: replace since_id with cursor fields for twitterapi.io compatibility
-- notes: removes since_id column, adds cursor and last_cursor fields

-- add new columns for cursor-based pagination
alter table ingest_runs
add column if not exists cursor text,
add column if not exists last_cursor text;

-- update existing records to have empty cursors (for new pagination)
update ingest_runs
set cursor = '', last_cursor = ''
where cursor is null or last_cursor is null;

-- make cursor columns not null with default empty string
alter table ingest_runs
alter column cursor set not null,
alter column cursor set default '',
alter column last_cursor set not null,
alter column last_cursor set default '';

-- remove the old since_id column
alter table ingest_runs drop column if exists since_id;

-- add comment to document the new cursor fields
comment on column ingest_runs.cursor is 'Current pagination cursor for ongoing ingestion';
comment on column ingest_runs.last_cursor is 'Last cursor used in previous ingestion run';

-- end of migration
