BEGIN;

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
        CREATE TYPE task_status AS ENUM ('pending', 'processing', 'completed', 'failed');
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    target_url TEXT NOT NULL,
    status task_status DEFAULT 'pending' NOT NULL,
    http_status_code INTEGER,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks (status);

INSERT INTO tasks (target_url) 
SELECT url FROM (
    VALUES
    ('https://github.com'),
    ('https://google.com'),
    ('https://this-is-a-completely-fake-domain-12345.com'),
    ('https://httpbin.org/delay/1'),
    ('https://github.com/this-path-does-not-exist'),
    ('https://httpbin.org/status/500'),
    ('https://aws.amazon.com'),
    ('https://api.github.com'),
    ('https://httpbin.org/delay/5'),
    ('https://golang.org');
) AS mock_data(url)
WHERE NOT EXISTS (SELECT 1 FROM tasks);

COMMIT;