BEGIN;

ALTER TABLE gitserver_repos ADD COLUMN last_fetched TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now();

COMMIT;
