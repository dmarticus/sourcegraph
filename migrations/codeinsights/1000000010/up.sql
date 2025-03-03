BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

-- Prior to 3.31 this table stored points in two formats. Historical points were stored in a compressed
-- format where samples were only recorded if the underlying repository changed. After 3.31 we are changing
-- the semantic to require full vectors for each data point. To avoid any incompatibilities and to prepare for beta
-- we are going to reset the stored data and all of the underlying Timescale chunks back to zero.
-- Note: This data is by design reproducible, so there is no risk of permanent data loss here. Any and all data
-- will be queued and regenerated as soon as code insights starts up.

-- Clean up the remaining records if any exist.
TRUNCATE series_points CASCADE;

-- There is the possibility that the commit index has fallen out of sync with the primary postgres database in 3.30 due
-- to a data corruption issue. We will regenerate it to be sure it is healthy for beta.
TRUNCATE commit_index;
TRUNCATE commit_index_metadata;

-- Update all of the underlying insights that may have been synced to reset metadata and rebuild their data.
update insight_series set created_at = current_timestamp, backfill_queued_at = null, next_recording_after = date_trunc('month', current_date) + interval '1 month';
COMMIT;

-- This file used to contain TimescaleDB specific migrations that have seen been deprecated.
