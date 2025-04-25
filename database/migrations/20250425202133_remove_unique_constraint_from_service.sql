-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS services_name_project_id_unique;
CREATE UNIQUE INDEX IF NOT EXISTS services_name_project_id_idx ON services(name, project_id);
DROP INDEX IF EXISTS services_name_project_id_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS services_name_project_id_unique ON services(name, project_id);
-- +goose StatementEnd
