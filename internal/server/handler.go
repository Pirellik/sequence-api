package server

import (
	"context"

	"github.com/google/uuid"
	"github.com/pirellik/sequence-api/internal/db/models"
	"github.com/pirellik/sequence-api/internal/openapi"
)

type StrictHandler struct {
	svc SequenceService
}

var _ openapi.StrictServerInterface = (*StrictHandler)(nil)

//go:generate go tool go.uber.org/mock/mockgen -source=handler.go -package=server -destination=mock_test.go -typed=true
type SequenceService interface {
	CreateSequence(ctx context.Context, sequence *models.Sequence, steps []*models.SequenceStep) (*models.Sequence, []*models.SequenceStep, error)
	UpdateSequence(ctx context.Context, id uuid.UUID, openTrackingEnabled, clickTrackingEnabled *bool) (*models.Sequence, []*models.SequenceStep, error)
	UpdateSequenceStep(ctx context.Context, sequenceID, stepID uuid.UUID, emailSubject, emailContent *string) (*models.SequenceStep, error)
	DeleteSequenceStep(ctx context.Context, sequenceID, stepID uuid.UUID) error
}

func NewHandler(svc SequenceService) *StrictHandler {
	return &StrictHandler{svc: svc}
}
