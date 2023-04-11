CREATE INDEX IF NOT EXISTS links_name_idx
	ON links USING GIN (to_tsvector('simple', name));

-- Not needed because unique constraint on token automatically generates btree index
-- CREATE INDEX IF NOT EXISTS links_token_idx
-- 	ON links(token);
