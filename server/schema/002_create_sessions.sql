CREATE TABLE sessions (
    token VARCHAR(255) PRIMARY KEY,
    data TEXT NOT NULL,
    expiry TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_expiry ON sessions(expiry);
