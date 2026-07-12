package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"taskTracker/internal/domain"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

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
		INSERT INTO tasks (
			title, description, due_date, status, 
			recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates, 
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at`

	var recType *string
	var recInterval *int
	var recDayOfMonth *int
	var recSpecifics []string

	if t.Recurrence != nil {
		strType := string(t.Recurrence.Type)
		recType = &strType

		if t.Recurrence.Interval > 0 {
			recInterval = &t.Recurrence.Interval
		}
		if t.Recurrence.DayOfMonth > 0 {
			recDayOfMonth = &t.Recurrence.DayOfMonth
		}
		if len(t.Recurrence.Specifics) > 0 {
			recSpecifics = make([]string, len(t.Recurrence.Specifics))
			for i, specDate := range t.Recurrence.Specifics {
				recSpecifics[i] = specDate.Format("2006-01-02")
			}
		}
	}

	var specificsArg any = nil
	if len(recSpecifics) > 0 {
		specificsArg = recSpecifics
	}

	return r.db.QueryRowContext(ctx, query,
		t.Title,
		t.Description,
		t.DueDate,
		t.Status,
		recType,
		recInterval,
		recDayOfMonth,
		specificsArg,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

// TaskViewer
func (r *TaskRepository) GetByID(ctx context.Context, id int64) (*domain.Task, error) {
	query := `SELECT id, title, description, due_date, status, created_at, updated_at, recurrence_type 
	          FROM tasks WHERE id = $1`

	var t domain.Task
	var recType *string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status, &t.CreatedAt, &t.UpdatedAt, &recType,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTaskNotFound
	}
	if err != nil {
		return nil, err
	}

	if recType != nil {
		t.Recurrence = &domain.TaskRecurrence{
			Type: domain.RecurrenceType(*recType),
		}
	}

	return &t, nil
}

// TaskViewer
func (r *TaskRepository) GetList(ctx context.Context, filter *domain.TaskFilter) ([]domain.Task, error) {
	query := `
		SELECT 
			id, title, description, due_date, status, created_at, updated_at,
			recurrence_type, recurrence_interval, recurrence_day_of_month, recurrence_specific_dates
		FROM tasks`
	var args []interface{}
	argCount := 1

	if filter.DueDateFrom != nil && filter.DueDateTo != nil {
		queryNotReq := query + fmt.Sprintf(` WHERE recurrence_type IS NULL AND due_date >= $%d AND due_date <= $%d`, argCount, argCount+1)
		args = append(args, *filter.DueDateFrom, *filter.DueDateTo)
		argCount += 2

		if filter.Status != nil {
			queryNotReq += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, *filter.Status)
			argCount++
		}

		queryReq := query + fmt.Sprintf(` WHERE recurrence_type IS NOT NULL AND due_date <= $%d`, argCount)
		args = append(args, *filter.DueDateTo)
		argCount++

		if filter.Status != nil {
			queryReq += fmt.Sprintf(" AND status = $%d", argCount)
			args = append(args, *filter.Status)
			argCount++
		}

		query = queryNotReq + ` UNION ALL ` + queryReq
	} else {
		if filter.Status != nil {
			query += fmt.Sprintf(" WHERE status = $%d", argCount)
			args = append(args, *filter.Status)
			argCount++
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("Database GetList query execution failed", "error", err, "query", query)
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task

	pgMap := pgtype.NewMap()
	rowsCount := 0

	for rows.Next() {
		rowsCount++
		var t domain.Task
		var recType *string
		var recInterval *int
		var recDayOfMonth *int

		var recSpecificsRaw []time.Time

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status, &t.CreatedAt, &t.UpdatedAt,
			&recType, &recInterval, &recDayOfMonth,
			pgMap.SQLScanner(&recSpecificsRaw),
		)
		if err != nil {
			slog.Error("Database GetList row scan failed", "error", err, "row_index", rowsCount)
			return nil, err
		}

		if recType != nil {
			t.Recurrence = &domain.TaskRecurrence{
				Type: domain.RecurrenceType(*recType),
			}
			if recInterval != nil {
				t.Recurrence.Interval = *recInterval
			}
			if recDayOfMonth != nil {
				t.Recurrence.DayOfMonth = *recDayOfMonth
			}
			t.Recurrence.Specifics = recSpecificsRaw
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Database rows iteration error", "error", err)
		return nil, err
	}

	return tasks, nil
}

// TaskModifier
func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) error {
	query := `
		UPDATE tasks 
		SET title = $1, description = $2, due_date = $3, status = $4, 
		    recurrence_type = $5, recurrence_interval = $6, recurrence_day_of_month = $7, recurrence_specific_dates = $8,
		    updated_at = NOW()
		WHERE id = $9`

	var recType *string
	var recInterval *int
	var recDayOfMonth *int
	var recSpecifics []string

	if t.Recurrence != nil {
		strType := string(t.Recurrence.Type)
		recType = &strType

		if t.Recurrence.Interval > 0 {
			recInterval = &t.Recurrence.Interval
		}
		if t.Recurrence.DayOfMonth > 0 {
			recDayOfMonth = &t.Recurrence.DayOfMonth
		}
		if len(t.Recurrence.Specifics) > 0 {
			recSpecifics = make([]string, len(t.Recurrence.Specifics))
			for i, specDate := range t.Recurrence.Specifics {
				recSpecifics[i] = specDate.Format("2006-01-02")
			}
		}
	}

	var specificsArg interface{} = nil
	if len(recSpecifics) > 0 {
		specificsArg = recSpecifics
	}

	res, err := r.db.ExecContext(ctx, query,
		t.Title,
		t.Description,
		t.DueDate,
		t.Status,
		recType,
		recInterval,
		recDayOfMonth,
		specificsArg,
		t.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTaskNotFound
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
		return domain.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepository) FetchExecutionsForPeriod(ctx context.Context, taskIDs []int64, from, to time.Time) (map[int64]map[string]domain.TaskStatus, error) {
	result := make(map[int64]map[string]domain.TaskStatus)
	if len(taskIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT task_id, execution_date, status 
		FROM task_executions 
		WHERE task_id = ANY($1) AND execution_date >= $2 AND execution_date <= $3`

	rows, err := r.db.QueryContext(ctx, query, taskIDs, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID int64
		var execDate time.Time
		var statusStr string
		if err := rows.Scan(&taskID, &execDate, &statusStr); err != nil {
			return nil, err
		}

		dateKey := execDate.Format("2006-01-02")
		if _, exists := result[taskID]; !exists {
			result[taskID] = make(map[string]domain.TaskStatus)
		}
		result[taskID][dateKey] = domain.TaskStatus(statusStr)
	}
	return result, rows.Err()
}

func (r *TaskRepository) SaveExecutionStatus(ctx context.Context, taskID int64, date time.Time, status domain.TaskStatus) error {
	query := `
		INSERT INTO task_executions (task_id, execution_date, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (task_id, execution_date) 
		DO UPDATE SET status = EXCLUDED.status`

	_, err := r.db.ExecContext(ctx, query, taskID, date.Format("2006-01-02"), string(status))
	return err
}
