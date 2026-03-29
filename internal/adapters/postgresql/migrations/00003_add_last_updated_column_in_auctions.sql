-- +goose Up
-- +goose StatementBegin
ALTER TABLE auctions
    ADD COLUMN last_updated TIMESTAMPTZ NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE auctions
    DROP COLUMN last_updated;
-- +goose StatementEnd
