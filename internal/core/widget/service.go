package widget

import (
	"context"
	"fmt"
	"go-platform-template/internal/core/apperror"
	"log/slog"
)

type Service struct {
	repo   Repository
	logger *slog.Logger
}

func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

func (s *Service) Create(ctx context.Context, w Widget) (Widget, error) {
	if err := w.Validate(); err != nil {
		s.logger.Debug("create widget error", "error", err)
		return Widget{}, err
	}

	newWidget, err := s.repo.Create(ctx, w)
	if err != nil {
		return Widget{}, err
	}

	return newWidget, nil
}

func (s *Service) Get(ctx context.Context, id int64) (Widget, error) {
	if id <= 0 {
		return Widget{}, fmt.Errorf("%w: id must be positive", apperror.ErrInvalid)
	}

	widget, err := s.repo.Get(ctx, id)
	if err != nil {
		s.logger.Debug("error while fetching widget", "err", err)
		return Widget{}, err
	}

	return widget, nil
}

func (s *Service) Update(ctx context.Context, w Widget) (Widget, error) {
	if err := w.Validate(); err != nil {
		s.logger.Debug("update widget error", "error", err)
		return Widget{}, err
	}

	if w.ID <= 0 {
		return Widget{}, fmt.Errorf("%w: id must be positive", apperror.ErrInvalid)
	}

	updatedWidget, err := s.repo.Update(ctx, w)
	if err != nil {
		return Widget{}, err
	}

	return updatedWidget, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", apperror.ErrInvalid)
	}

	err := s.repo.Delete(ctx, id)
	if err != nil {
		s.logger.Debug("error while deleting widget", "err", err)
		return err
	}

	return nil
}

func (s *Service) List(ctx context.Context) ([]Widget, error) {
	widgets, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	return widgets, nil
}
