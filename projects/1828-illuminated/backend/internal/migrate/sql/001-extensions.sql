-- 001-extensions.sql
-- Postgres extensions needed by later migrations. pg_trgm for trigram
-- indexes (search by prefix / fuzzy match on scripture text + 1828
-- headwords). unaccent is a small extra we may want later for archaic
-- text normalization; enabling now avoids a second restart.

CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;
