CREATE TABLE IF NOT EXISTS task_executions (
    task_id BIGINT NOT NULL,
    execution_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL, 
    override_title TEXT DEFAULT NULL,
    override_description TEXT DEFAULT NULL,
    
    PRIMARY KEY (task_id, execution_date),
    CONSTRAINT fk_task_executions_task FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_task_executions_date ON task_executions (execution_date);