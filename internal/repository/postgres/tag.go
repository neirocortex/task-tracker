package postgres

import (
	"context"
	"database/sql"
	"errors"
	"taskTracker/internal/domain"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) SyncTaskTags(ctx context.Context, taskID int64, tagNames []string) ([]string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "DELETE FROM task_tags WHERE task_id = $1", taskID)
	if err != nil {
		return nil, err
	}

	if len(tagNames) == 0 {
		return []string{}, tx.Commit()
	}

	insertQuery := `
		INSERT INTO task_tags (task_id, tag_name)
		SELECT $1, name 
		FROM tags 
		WHERE name = $2
		ON CONFLICT DO NOTHING
		RETURNING tag_name`

	stmt, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var syncTags []string
	for _, name := range tagNames {
		var insertedName string
		err := stmt.QueryRowContext(ctx, taskID, name).Scan(&insertedName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// silent fail: tag is removed
				continue
			}
			return nil, err
		}
		syncTags = append(syncTags, insertedName)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return syncTags, nil
}

func (r *TagRepository) FetchTagsForTasks(ctx context.Context, taskIDs []int64) (map[int64][]domain.Tag, error) {
	result := make(map[int64][]domain.Tag)
	if len(taskIDs) == 0 {
		return result, nil
	}

	query := `
		SELECT tt.task_id, tt.tag_name, t.is_system 
		FROM task_tags tt
		JOIN tags t ON tt.tag_name = t.name
		WHERE tt.task_id = ANY($1)`

	rows, err := r.db.QueryContext(ctx, query, taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID int64
		var t domain.Tag
		if err := rows.Scan(&taskID, &t.Name, &t.IsSystem); err != nil {
			return nil, err
		}
		result[taskID] = append(result[taskID], t)
	}
	return result, nil
}

func (r *TagRepository) FindAllTags(ctx context.Context) ([]domain.Tag, error) {
	query := `SELECT name, is_system FROM tags ORDER BY is_system DESC, name ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.Name, &t.IsSystem); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *TagRepository) DeleteTagFromRegistry(ctx context.Context, tagName string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM tags WHERE name = $1 AND is_system = FALSE", tagName)
	return err
}

func (r *TagRepository) CreateTagInRegistry(ctx context.Context, tag domain.Tag) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO tags (name, is_system) VALUES ($1, $2) ON CONFLICT DO NOTHING", tag.Name, tag.IsSystem)
	return err
}
