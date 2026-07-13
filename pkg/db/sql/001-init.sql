BEGIN;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
        CREATE TYPE task_status AS ENUM ('pending', 'processing', 'completed', 'failed');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    status task_status DEFAULT 'pending' NOT NULL,
    attempts INTEGER DEFAULT 0 NOT NULL,
    max_attempts INTEGER DEFAULT 3 NOT NULL,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_status_polling ON tasks (id) WHERE status = 'pending';

-- 1. Seed the individual distinct mock targets safely
INSERT INTO tasks (url) 
SELECT mock_url FROM (
    VALUES
    ('https://github.com'),
    ('https://google.com'),
    ('https://this-is-a-completely-fake-domain-12345.com'),
    ('https://github.com/this-path-does-not-exist'),
    ('https://httpbin.org/status/500'),
    ('https://aws.amazon.com'),
    ('https://api.github.com'),
    ('https://golang.org')
) AS base_mock(mock_url)
WHERE NOT EXISTS (
    SELECT 1 FROM tasks WHERE url = base_mock.mock_url
);

-- 2. Seed the 500 heavy load testing tasks safely
INSERT INTO tasks (url)
SELECT 'https://httpbin.org/delay/1'
FROM generate_series(1, 500)
WHERE NOT EXISTS (
    SELECT 1 FROM tasks WHERE url = 'https://httpbin.org/delay/1'
);

COMMIT;
