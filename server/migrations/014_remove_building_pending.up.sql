-- Migration 014: Remove pending column (replaced by build_progress state machine)
ALTER TABLE buildings DROP COLUMN IF EXISTS pending;
