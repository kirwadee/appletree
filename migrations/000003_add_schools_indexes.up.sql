--Filename:migrations/000003/add_schools_indexes.up.sql

CREATE INDEX IF NOT EXISTS schools_name_idx ON schools USING GIN(to_tsvector('simple', name))