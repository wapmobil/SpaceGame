-- Migration 013: Add dynamo building type
ALTER TABLE buildings DROP CONSTRAINT IF EXISTS buildings_type_check;
ALTER TABLE buildings ADD CONSTRAINT buildings_type_check
    CHECK (type IN ('farm','solar','storage','base','factory','energy_storage','shipyard','command_center','composite_drone','mechanism_factory','reagent_lab','dynamo'));
