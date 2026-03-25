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
    highestBid = $2
WHERE
    id = $3 AND highestBid < $1;

-- name: PlaceBid :exec
INSERT INTO
    bids (auctionId, bidderId, amount)
VALUES
    ($1, $2, $3);

-- name: GetAuctionByID :one
SELECT
    id,
    title,
    ownerId,
    winnerId,
    basePrice,
    highestBid,
    state
FROM
    auctions
WHERE
    id = $1;