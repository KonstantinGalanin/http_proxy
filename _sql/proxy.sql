DROP TABLE IF EXISTS requests;
CREATE TABLE requests (
    id SERIAL PRIMARY KEY,
    method VARCHAR(10) NOT NULL,
    path TEXT NOT NULL,
    host TEXT NOT NULL,
    get_params JSONB,
    headers JSONB NOT NULL,
    cookies JSONB,
    post_params JSONB,
    body BYTEA,
    created_at TIMESTAMP DEFAULT NOW()
);