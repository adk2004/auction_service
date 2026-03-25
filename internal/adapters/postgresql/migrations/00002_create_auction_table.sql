-- +goose Up
-- +goose StatementBegin

CREATE TYPE auction_status AS ENUM ('ongoing', 'ended');

CREATE TABLE IF NOT EXISTS auctions(
    id BIGSERIAL PRIMARY KEY,
    title TEXT  NOT NULL,
    ownerId BIGINT NOT NULL,
    winnerId BIGINT,
    basePrice BIGINT NOT NULL,
    highestBid BIGINT NOT NULL,
    state auction_status NOT NULL,
    CONSTRAINT fk_owner FOREIGN KEY (ownerId) REFERENCES users(id),
    CONSTRAINT fk_winner FOREIGN KEY (winnerId) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS bids(
    id BIGSERIAL PRIMARY KEY,
    auctionId BIGINT NOT NULL,
    bidderId BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    CONSTRAINT fk_auction FOREIGN KEY (auctionId) REFERENCES auctions(id),
    CONSTRAINT fk_bidder FOREIGN KEY (bidderId) REFERENCES users(id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auctions;
DROP TYPE IF EXISTS auction_status;

DROP TABLE IF EXISTS bids;
-- +goose StatementEnd
