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

INSERT INTO tasks (url) 
SELECT mock_url FROM (
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
    ('https://golang.org')
) AS mock_data(mock_url)
WHERE NOT EXISTS (SELECT 1 FROM tasks);

COMMIT;