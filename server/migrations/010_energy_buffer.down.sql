-- Migration 010 down
ALTER TABLE planets DROP COLUMN IF EXISTS energy_buffer;
