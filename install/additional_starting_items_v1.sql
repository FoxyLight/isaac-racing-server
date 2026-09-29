USE isaac;

ALTER TABLE races
    ADD COLUMN additional_starting_items VARCHAR(64) NOT NULL DEFAULT "" AFTER starting_build;
