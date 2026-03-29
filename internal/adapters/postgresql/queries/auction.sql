-- name: CreateAuction :one
INSERT INTO
    auctions (title, basePrice, highestBid, state, ownerId)
VALUES
    ($1, $2, $3, $4, $5)
RETURNING
    id;

-- name: UpdateAuctionWinner :exec
UPDATE auctions
SET
    winnerId = $1,
    highestBid = $2,
    last_updated = NOW()
WHERE
    id = $3
    AND highestBid < $2
    AND state = 'ongoing';

-- name: PlaceBid :exec
INSERT INTO
    bids (auctionId, bidderId, amount)
VALUES
    ($1, $2, $3);

-- name: GetAuctionByID :one
SELECT
    *
FROM
    auctions
WHERE
    id = $1;

-- name: GetOngoingAuctions :many
SELECT
    id
FROM
    auctions
WHERE
    state = 'ongoing';

-- name: EndAuction :exec
UPDATE auctions
SET
    state = 'ended'
WHERE
    id = $1
    AND state = 'ongoing';

-- name: GetCompletedAuctions :many
SELECT
    *
FROM
    auctions
WHERE
    state = 'ongoing'
    AND last_updated < NOW() - INTERVAL '5 minutes';