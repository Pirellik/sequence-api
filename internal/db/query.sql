-- name: CreateSequence :one
INSERT INTO sequences (
  name, open_tracking_enabled, click_tracking_enabled
) VALUES ($1, $2, $3) RETURNING id;

-- name: UpdateSequence :exec
UPDATE sequences SET open_tracking_enabled = $1, click_tracking_enabled = $2, updated_at = NOW() WHERE id = $3;

-- name: GetSequenceByID :one
SELECT * FROM sequences WHERE id = $1 LIMIT 1;

-- name: CreateSequenceStep :one
INSERT INTO sequence_steps (
  sequence_id, days_after_previous_step, email_subject, email_content, ordering
) VALUES ($1, $2, $3, $4, $5) RETURNING id;

-- name: UpdateSequenceStep :exec
UPDATE sequence_steps SET email_subject = $1, email_content = $2, updated_at = NOW() WHERE id = $3;

-- name: GetSequenceStepByID :one
SELECT * FROM sequence_steps WHERE id = $1 LIMIT 1;

-- name: GetSequenceStepsBySequenceID :many
SELECT * FROM sequence_steps WHERE sequence_id = $1 ORDER BY ordering ASC;

-- name: DeleteSequenceStep :exec
DELETE FROM sequence_steps WHERE id = $1;
