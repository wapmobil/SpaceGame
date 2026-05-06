-- Add 'generating' status to expedition_chains for async LLM event generation
ALTER TABLE expedition_chains
    DROP CONSTRAINT IF EXISTS expedition_chains_status_check;

ALTER TABLE expedition_chains
    ADD CONSTRAINT expedition_chains_status_check
    CHECK (status IN ('active', 'generating', 'completed', 'failed'));
