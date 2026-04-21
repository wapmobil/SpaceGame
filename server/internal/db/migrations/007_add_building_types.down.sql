-- Migration 007 down: Remove new building types
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS buildings_type_check;
ALTER TABLE buildings ADD CONSTRAINT buildings_type_check
    CHECK (type IN ('farm','solar','storage','base','factory','energy_storage','shipyard','comcenter'));
