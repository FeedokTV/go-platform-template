package widget

import (
	"fmt"
	"go-platform-template/internal/core/apperror"
	"time"
)

type Widget struct {
	ID        int64
	Name      string
	Weight    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (w Widget) Validate() error {
	if len(w.Name) <= 4 || len(w.Name) > 250 {
		return fmt.Errorf("%w: name length must be 5-250", apperror.ErrInvalid)
	}

	if w.Weight < 0 {
		return fmt.Errorf("%w: weight must be equal or greater than zero", apperror.ErrInvalid)
	}

	return nil
}
