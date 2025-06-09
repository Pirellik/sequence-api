package sequence

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pirellik/sequence-api/internal/db/models"
)

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) CreateSequence(
	ctx context.Context,
	sequence *models.Sequence,
	steps []*models.SequenceStep,
) (*models.Sequence, []*models.SequenceStep, error) {
	q := models.New(s.db)
	id, err := q.CreateSequence(ctx, &models.CreateSequenceParams{
		Name:                 sequence.Name,
		OpenTrackingEnabled:  sequence.OpenTrackingEnabled,
		ClickTrackingEnabled: sequence.ClickTrackingEnabled,
	})
	if err != nil {
		return nil, nil, err
	}

	for i, step := range steps {
		_, err = q.CreateSequenceStep(ctx, &models.CreateSequenceStepParams{
			SequenceID:            id,
			EmailSubject:          step.EmailSubject,
			EmailContent:          step.EmailContent,
			DaysAfterPreviousStep: step.DaysAfterPreviousStep,
			Ordering:              float32(i),
		})
		if err != nil {
			return nil, nil, err
		}
	}

	created, err := q.GetSequenceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	steps, err = q.GetSequenceStepsBySequenceID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return created, steps, nil
}

func (s *Service) UpdateSequence(
	ctx context.Context,
	id uuid.UUID,
	openTrackingEnabled, clickTrackingEnabled *bool,
) (*models.Sequence, []*models.SequenceStep, error) {
	q := models.New(s.db)
	sequence, err := q.GetSequenceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	params := models.UpdateSequenceParams{
		ID:                   id,
		OpenTrackingEnabled:  sequence.OpenTrackingEnabled,
		ClickTrackingEnabled: sequence.ClickTrackingEnabled,
	}

	if openTrackingEnabled != nil {
		params.OpenTrackingEnabled = *openTrackingEnabled
	}
	if clickTrackingEnabled != nil {
		params.ClickTrackingEnabled = *clickTrackingEnabled
	}

	err = q.UpdateSequence(ctx, &params)
	if err != nil {
		return nil, nil, err
	}

	updated, err := q.GetSequenceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	steps, err := q.GetSequenceStepsBySequenceID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return updated, steps, nil
}

func (s *Service) UpdateSequenceStep(
	ctx context.Context,
	sequenceID, stepID uuid.UUID,
	emailSubject, emailContent *string,
) (*models.SequenceStep, error) {
	q := models.New(s.db)
	step, err := q.GetSequenceStepByID(ctx, stepID)
	if err != nil {
		return nil, err
	}

	params := models.UpdateSequenceStepParams{
		ID:           stepID,
		EmailSubject: step.EmailSubject,
		EmailContent: step.EmailContent,
	}
	if emailSubject != nil {
		params.EmailSubject = *emailSubject
	}
	if emailContent != nil {
		params.EmailContent = *emailContent
	}

	err = q.UpdateSequenceStep(ctx, &params)
	if err != nil {
		return nil, err
	}

	updated, err := q.GetSequenceStepByID(ctx, stepID)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) DeleteSequenceStep(ctx context.Context, sequenceID, stepID uuid.UUID) error {
	err := models.New(s.db).DeleteSequenceStep(ctx, stepID)
	if err != nil {
		return err
	}

	return nil
}
