-- 005-saved-trees.sql
-- Tables for cloud syncing of study trees.
-- Maps to the user session validated from becoming.ibeco.me.

CREATE TABLE IF NOT EXISTS users (
  becoming_user_id  BIGINT PRIMARY KEY,
  email             TEXT NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS saved_trees (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  becoming_user_id  BIGINT NOT NULL REFERENCES users(becoming_user_id) ON DELETE CASCADE,
  title             TEXT NOT NULL,
  tree_data         JSONB NOT NULL,
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_saved_trees_user ON saved_trees(becoming_user_id);
