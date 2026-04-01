package pubsub

import "time"

type Event struct {
    AuctionID int64     `json:"auctionId"`
    BidderID  int64     `json:"bidderId"`
    BidAmount float64   `json:"bidAmount"`
    Completed bool      `json:"completed"`
    Time      time.Time `json:"time"`
}
