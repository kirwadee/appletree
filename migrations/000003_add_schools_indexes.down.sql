--Filename:migrations/000003/add_schools_indexes.up.sql

DROP INDEX IF EXISTS schools_name_idx;
DROP INDEX IF EXISTS schools_level_idx;
DROP INDEX IF EXISTS schools_mode_idx;