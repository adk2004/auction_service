package pubsub

import "time"

type Event struct {
	AuctionID int64
	BidderID  int64
	BidAmount float64
	Completed bool
	Time   time.Time
}