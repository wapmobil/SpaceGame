-- Migration 014 down: Add pending column back
ALTER TABLE buildings ADD COLUMN pending BOOLEAN NOT NULL DEFAULT false;
