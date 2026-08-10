package postgresadapter

import (
	"context"
	"errors"
	"fmt"
	"go-platform-template/internal/core/apperror"
	"go-platform-template/internal/core/widget"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ widget.Repository = (*WidgetRepository)(nil)

type WidgetRepository struct {
	pool *pgxpool.Pool
}

func NewWidgetRepository(pool *pgxpool.Pool) *WidgetRepository {
	return &WidgetRepository{pool: pool}
}

func (wr *WidgetRepository) Create(ctx context.Context, w widget.Widget) (widget.Widget, error) {
	// Return in one query, its about concurrency. We want data exactly after creation
	// That's why we dont use two queries: create + select then
	const q = `INSERT INTO widgets(name, weight)
				VALUES ($1, $2)
				RETURNING id, created_at, updated_at`

	var id int64
	var createdAt, updatedAt time.Time

	row := wr.pool.QueryRow(ctx, q, w.Name, w.Weight)

	err := row.Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return widget.Widget{}, fmt.Errorf("create widget: %w", err)
	}

	w.ID = id
	w.CreatedAt = createdAt
	w.UpdatedAt = updatedAt

	return w, nil
}

func (wr *WidgetRepository) Get(ctx context.Context, id int64) (widget.Widget, error) {
	const q = `SELECT id, name, weight, created_at, updated_at FROM widgets WHERE id=$1`

	var fetchedWidget widget.Widget

	row := wr.pool.QueryRow(ctx, q, id)

	err := row.Scan(&fetchedWidget.ID, &fetchedWidget.Name, &fetchedWidget.Weight, &fetchedWidget.CreatedAt, &fetchedWidget.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return widget.Widget{}, apperror.ErrNotFound
		} else {
			return widget.Widget{}, fmt.Errorf("get widget: %w", err)
		}
	}

	return fetchedWidget, nil
}

func (wr *WidgetRepository) Update(ctx context.Context, w widget.Widget) (widget.Widget, error) {
	const q = `UPDATE widgets
				SET name = $2, weight = $3, updated_at = NOW()
				WHERE id=$1
				RETURNING created_at, updated_at`

	var createdAt, updatedAt time.Time

	err := wr.pool.QueryRow(ctx, q, w.ID, w.Name, w.Weight).Scan(&createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return widget.Widget{}, apperror.ErrNotFound
		} else {
			return widget.Widget{}, fmt.Errorf("update widget: %w", err)
		}
	}

	w.CreatedAt = createdAt
	w.UpdatedAt = updatedAt

	return w, nil
}

func (wr *WidgetRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM widgets WHERE id=$1`

	commandTag, err := wr.pool.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete widget: %w", err)
	}

	if commandTag.RowsAffected() != 1 {
		return apperror.ErrNotFound
	}

	return nil
}

func (wr *WidgetRepository) List(ctx context.Context) ([]widget.Widget, error) {
	const q = `SELECT id, name, weight, created_at, updated_at FROM widgets`

	widgetList := []widget.Widget{}

	rows, err := wr.pool.Query(ctx, q)
	if err != nil {
		return widgetList, fmt.Errorf("list widget: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w widget.Widget
		if err := rows.Scan(&w.ID, &w.Name, &w.Weight, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return widgetList, fmt.Errorf("list widget: %w", err)
		}

		widgetList = append(widgetList, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list widget: %w", err)
	}

	return widgetList, nil
}
