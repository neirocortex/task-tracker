CREATE TABLE IF NOT EXISTS tags (
    name VARCHAR(50) PRIMARY KEY,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS task_tags (
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL REFERENCES tags(name) ON DELETE CASCADE, -- Авто-очистка связей
    PRIMARY KEY (task_id, tag_name)
);

CREATE INDEX IF NOT EXISTS idx_task_tags_tag ON task_tags(tag_name);

INSERT INTO tags (name, is_system) VALUES
('отчетность', TRUE),
('операции', TRUE),
('звонок', TRUE)
ON CONFLICT (name) DO UPDATE SET is_system = TRUE;
