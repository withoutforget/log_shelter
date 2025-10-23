CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    raw_log TEXT NOT NULL,
    log_level VARCHAR(16) NOT NULL,
    source VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    request_id VARCHAR(64),
    logger_name VARCHAR(128),
    is_deleted BOOLEAN DEFAULT FALSE
);