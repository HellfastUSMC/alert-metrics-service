-- +goose Up
CREATE TABLE IF NOT EXISTS Metrics (
    		NAME text NOT NULL UNIQUE PRIMARY KEY,
    		TYPE text NOT NULL,
    		VALUE double precision,
    		DELTA bigint
        );

-- +goose Down
DROP TABLE Metrics;