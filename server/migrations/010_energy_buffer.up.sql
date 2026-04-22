-- Migration 010: Add energy buffer to planets table
ALTER TABLE planets ADD COLUMN IF NOT EXISTS energy_buffer REAL NOT NULL DEFAULT 100;
