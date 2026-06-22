DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'task_status') THEN
        CREATE TYPE task_status AS ENUM ('NEW', 'IN_PROGRESS', 'DONE', 'CANCELED');
    END IF;
END$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'recurrence_type_enum') THEN
        CREATE TYPE recurrence_type_enum AS ENUM ('DAILY', 'MONTHLY', 'DATES', 'EVEN', 'ODD');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS tasks (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status task_status NOT NULL DEFAULT 'NEW',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    recurrence_type recurrence_type_enum DEFAULT NULL,
    recurrence_interval INT DEFAULT NULL,
    recurrence_day_of_month INT DEFAULT NULL,
    recurrence_specific_dates DATE[] DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_due_date ON tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_recurring_templates ON tasks(due_date) WHERE recurrence_type IS NOT NULL;