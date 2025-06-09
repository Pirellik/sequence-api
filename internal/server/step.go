package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pirellik/sequence-api/internal/db/models"
	"github.com/pirellik/sequence-api/internal/openapi"
)

func SequenceStepFromDB(step *models.SequenceStep) openapi.SequenceStep {
	return openapi.SequenceStep{
		Id:                    step.ID,
		EmailSubject:          step.EmailSubject,
		EmailContent:          step.EmailContent,
		DaysAfterPreviousStep: int(step.DaysAfterPreviousStep),
		CreatedAt:             &step.CreatedAt.Time,
		UpdatedAt:             &step.UpdatedAt.Time,
	}
}

func (s *StrictHandler) UpdateSequenceStep(ctx context.Context, request openapi.UpdateSequenceStepRequestObject) (openapi.UpdateSequenceStepResponseObject, error) {
	sequenceID, err := uuid.Parse(request.SequenceId)
	if err != nil {
		return nil, ErrBadRequest("Invalid sequence ID")
	}

	stepID, err := uuid.Parse(request.StepId)
	if err != nil {
		return nil, ErrBadRequest("Invalid step ID")
	}

	step, err := s.svc.UpdateSequenceStep(ctx, sequenceID, stepID, request.Body.EmailSubject, request.Body.EmailContent)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("Sequence step not found")
		}
		return nil, ErrInternal("Failed to update sequence step")
	}

	return openapi.UpdateSequenceStep200JSONResponse(SequenceStepFromDB(step)), nil
}

func (s *StrictHandler) DeleteSequenceStep(ctx context.Context, request openapi.DeleteSequenceStepRequestObject) (openapi.DeleteSequenceStepResponseObject, error) {
	sequenceID, err := uuid.Parse(request.SequenceId)
	if err != nil {
		return nil, ErrBadRequest("Invalid sequence ID")
	}

	stepID, err := uuid.Parse(request.StepId)
	if err != nil {
		return nil, ErrBadRequest("Invalid step ID")
	}

	err = s.svc.DeleteSequenceStep(ctx, sequenceID, stepID)
	if err != nil {
		return nil, ErrInternal("Failed to delete sequence step")
	}

	return openapi.DeleteSequenceStep204Response{}, nil
}
