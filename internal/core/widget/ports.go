package widget

import "context"

type Repository interface {
	Create(ctx context.Context, w Widget) (Widget, error)
	Get(ctx context.Context, id int64) (Widget, error)
	Update(ctx context.Context, w Widget) (Widget, error)
	List(ctx context.Context) ([]Widget, error)
	Delete(ctx context.Context, id int64) error
}
