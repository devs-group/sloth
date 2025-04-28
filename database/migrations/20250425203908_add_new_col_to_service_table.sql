-- +goose Up
-- +goose StatementBegin
ALTER TABLE services
ADD COLUMN usn VARCHAR(255);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
ALTER TABLE services
DROP COLUMN usn;

-- +goose StatementEnd
