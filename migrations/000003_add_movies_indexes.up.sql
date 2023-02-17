CREATE INDEX IF NOT EXISTS footballers_name_idx ON footballers USING GIN (to_tsvector('simple', names));
CREATE INDEX IF NOT EXISTS footballers_positions_idx ON footballers USING GIN (positions);
