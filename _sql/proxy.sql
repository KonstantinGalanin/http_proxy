DROP TABLE IF EXISTS requests;
CREATE TABLE requests (
    id SERIAL PRIMARY KEY,
    scheme TEXT NOT NULL,
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

DROP TABLE IF EXISTS responses;
CREATE TABLE response (
    id SERIAL PRIMARY KEY,
    status_code INT NOT NULL,
    status TEXT NOT NULL,
    headers JSONB NOT NULL,
    body BYTEA,
    content_length INT,
    compressed BOOLEAN DEFAULT FALSE
);
