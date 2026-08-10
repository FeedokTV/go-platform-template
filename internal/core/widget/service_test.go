package widget_test

import (
	"context"
	"errors"
	"go-platform-template/internal/core/apperror"
	"go-platform-template/internal/core/widget"
	"log/slog"
	"testing"
	"time"
)

type FakeRepository struct {
	FakeDB map[int64]widget.Widget
	LastID int64
}

func newTestService() *widget.Service {
	fakeRepo := &FakeRepository{
		FakeDB: make(map[int64]widget.Widget),
		LastID: 0,
	}

	testLogger := slog.New(slog.DiscardHandler)

	return widget.NewService(fakeRepo, testLogger)
}

func (fr *FakeRepository) Create(ctx context.Context, w widget.Widget) (widget.Widget, error) {
	now := time.Now()

	id := fr.LastID + 1

	w.ID = id
	w.CreatedAt = now
	w.UpdatedAt = now

	fr.FakeDB[id] = w

	fr.LastID = id

	return w, nil
}

func (fr *FakeRepository) Get(ctx context.Context, id int64) (widget.Widget, error) {
	w, ok := fr.FakeDB[id]
	if !ok {
		return widget.Widget{}, apperror.ErrNotFound
	}

	return w, nil
}

func (fr *FakeRepository) Update(ctx context.Context, w widget.Widget) (widget.Widget, error) {
	stored, ok := fr.FakeDB[w.ID]
	if !ok {
		return widget.Widget{}, apperror.ErrNotFound
	}

	now := time.Now()

	stored.Name = w.Name
	stored.Weight = w.Weight
	stored.UpdatedAt = now

	fr.FakeDB[stored.ID] = stored

	return stored, nil
}

func (fr *FakeRepository) List(ctx context.Context) ([]widget.Widget, error) {
	widgets := []widget.Widget{}

	for _, widget := range fr.FakeDB {
		widgets = append(widgets, widget)
	}

	return widgets, nil
}

func (fr *FakeRepository) Delete(ctx context.Context, id int64) error {
	_, ok := fr.FakeDB[id]
	if !ok {
		return apperror.ErrNotFound
	}

	delete(fr.FakeDB, id)

	return nil
}

func TestCreate(t *testing.T) {
	tests := []struct {
		Name    string
		Input   widget.Widget
		Want    widget.Widget
		WantErr error
	}{
		{
			Name:    "create success",
			Input:   widget.Widget{Name: "widget1", Weight: 42},
			Want:    widget.Widget{ID: 1, Name: "widget1", Weight: 42},
			WantErr: nil,
		},
		{
			Name:    "create invalid weight",
			Input:   widget.Widget{Name: "widget2", Weight: -1},
			Want:    widget.Widget{},
			WantErr: apperror.ErrInvalid,
		},
		{
			Name:    "create invalid name",
			Input:   widget.Widget{Name: "w1", Weight: 42},
			Want:    widget.Widget{},
			WantErr: apperror.ErrInvalid,
		},
	}

	service := newTestService()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := service.Create(t.Context(), test.Input)

			if test.WantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, test.WantErr) {
					t.Fatalf("expected error: %v, got: %v", test.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check fields
			if got.ID != test.Want.ID {
				t.Errorf("got ID %d, want %d", got.ID, test.Want.ID)
			}

			if got.Name != test.Want.Name {
				t.Errorf("got Name %q, want %q", got.Name, test.Want.Name)
			}

			if got.Weight != test.Want.Weight {
				t.Errorf("got Weight %q, want %q", got.Weight, test.Want.Weight)
			}

		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		Name    string
		Input   int64
		Want    widget.Widget
		WantErr error
	}{
		{
			Name:    "get success",
			Input:   1,
			Want:    widget.Widget{ID: 1, Name: "widget1", Weight: 42},
			WantErr: nil,
		},
		{
			Name:    "get not found",
			Input:   3,
			Want:    widget.Widget{},
			WantErr: apperror.ErrNotFound,
		},
		{
			Name:    "get invalid ID",
			Input:   -1,
			Want:    widget.Widget{},
			WantErr: apperror.ErrInvalid,
		},
	}

	service := newTestService()
	createdWidget, err := service.Create(t.Context(), widget.Widget{Name: "widget1", Weight: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := service.Get(t.Context(), test.Input)

			if test.WantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, test.WantErr) {
					t.Fatalf("expected error: %v, got: %v", test.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check fields
			if got.ID != test.Want.ID {
				t.Errorf("got ID %d, want %d", got.ID, test.Want.ID)
			}

			if got.Name != test.Want.Name {
				t.Errorf("got Name %q, want %q", got.Name, test.Want.Name)
			}

			if got.Weight != test.Want.Weight {
				t.Errorf("got Weight %q, want %q", got.Weight, test.Want.Weight)
			}

			// Check creation time
			if !createdWidget.CreatedAt.Equal(got.CreatedAt) {
				t.Errorf("creation time are not equal: after creation: %v, after fetching: %v", createdWidget.CreatedAt, got.CreatedAt)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		Name    string
		Input   widget.Widget
		Want    widget.Widget
		WantErr error
	}{
		{
			Name:    "update success",
			Input:   widget.Widget{ID: 1, Name: "widget_updated", Weight: 42},
			Want:    widget.Widget{ID: 1, Name: "widget_updated", Weight: 42},
			WantErr: nil,
		},
		{
			Name:    "update validation error",
			Input:   widget.Widget{ID: 1, Name: "up1", Weight: -1},
			Want:    widget.Widget{},
			WantErr: apperror.ErrInvalid,
		},
		{
			Name:    "update not found",
			Input:   widget.Widget{ID: 10, Name: "widget_updated", Weight: 42},
			Want:    widget.Widget{},
			WantErr: apperror.ErrNotFound,
		},
	}

	service := newTestService()
	createdWidget, err := service.Create(t.Context(), widget.Widget{Name: "widget1", Weight: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Sleep for gaining difference between creation and update
	time.Sleep(time.Millisecond)

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := service.Update(t.Context(), test.Input)

			if test.WantErr != nil {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !errors.Is(err, test.WantErr) {
					t.Fatalf("expected error: %v, got: %v", test.WantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check fields
			if got.ID != test.Want.ID {
				t.Errorf("got ID %d, want %d", got.ID, test.Want.ID)
			}

			if got.Name != test.Want.Name {
				t.Errorf("got Name %q, want %q", got.Name, test.Want.Name)
			}

			if got.Weight != test.Want.Weight {
				t.Errorf("got Weight %q, want %q", got.Weight, test.Want.Weight)
			}

			// Check updated time changed
			if createdWidget.UpdatedAt.Equal(got.UpdatedAt) {
				t.Errorf("update time are equal: after creation: %v, after update: %v", createdWidget.UpdatedAt, got.UpdatedAt)
			}

			// Check creation time not changed
			if !createdWidget.CreatedAt.Equal(got.CreatedAt) {
				t.Errorf("create time are changed: before update: %v, after update: %v", createdWidget.CreatedAt, got.CreatedAt)
			}
		})
	}
}

func TestDelete(t *testing.T) {

	service := newTestService()

	createdWidget, err := service.Create(t.Context(), widget.Widget{Name: "widget1", Weight: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert deletion
	err = service.Delete(t.Context(), createdWidget.ID)
	if err != nil {
		t.Fatalf("error while deleting widget: %v", err)
	}

	_, err = service.Get(t.Context(), createdWidget.ID)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Errorf("want err: %v, got: %v", apperror.ErrInvalid, err)
	}

	// Delete with wrong ID
	err = service.Delete(t.Context(), 200)
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	if !errors.Is(err, apperror.ErrNotFound) {
		t.Errorf("want err: %v, got: %v", apperror.ErrInvalid, err)
	}
}

func TestList(t *testing.T) {
	service := newTestService()

	_, err := service.Create(t.Context(), widget.Widget{Name: "widget1", Weight: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = service.Create(t.Context(), widget.Widget{Name: "widget2", Weight: 110})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotWidgets, err := service.List(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(gotWidgets) != 2 {
		t.Errorf("expected list len: 2, got: %d", len(gotWidgets))
	}

}
