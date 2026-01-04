-- +goose Up
-- +goose StatementBegin
ALTER TABLE chunks
ADD COLUMN start_line INT;

ALTER TABLE chunks
ADD COLUMN end_line INT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
