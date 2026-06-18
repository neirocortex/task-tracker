package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"taskTracker/internal/domain"
)

var ErrTaskNotFound = errors.New("task not found")

// implements db contracts
type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// TaskSaver
func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error {
	query := `
		INSERT INTO tasks (title, description, due_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query, t.Title, t.Description, t.DueDate, t.Status).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

// TaskViewer
func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	query := `SELECT id, title, description, due_date, status, created_at, updated_at FROM tasks WHERE id = $1`

	var t domain.Task
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status, &t.CreatedAt, &t.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	return &t, err
}

// TaskViewer
func (r *TaskRepository) GetList(ctx context.Context, filter domain.TaskFilter) ([]domain.Task, error) {
	query := `SELECT id, title, description, due_date, status, created_at, updated_at FROM tasks WHERE 1=1`
	var args []interface{}
	argCount := 1

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *filter.Status)
		argCount++
	}
	if filter.DueDateFrom != nil {
		query += fmt.Sprintf(" AND due_date >= $%d", argCount)
		args = append(args, *filter.DueDateFrom)
		argCount++
	}
	if filter.DueDateTo != nil {
		query += fmt.Sprintf(" AND due_date <= $%d", argCount)
		args = append(args, *filter.DueDateTo)
		argCount++
	}

	query += " ORDER BY due_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// TaskModifier
func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) error {
	query := `
		UPDATE tasks 
		SET title = $1, description = $2, due_date = $3, status = $4, updated_at = NOW()
		WHERE id = $5`

	res, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.DueDate, t.Status, t.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// TaskRemover
func (r *TaskRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM tasks WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrTaskNotFound
	}
	return nil
}
