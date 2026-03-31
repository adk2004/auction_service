package types

type PostAuctionRequest struct {
	OwnerId int64 `json:"ownerId"`
	BasePrice int64 `json:"basePrice"`
	Title string `json:"title"`
}

type PostBidRequest struct {
	UserId    int64 `json:"userId"`
	AuctionId int64 `json:"auctionId"`
	Amount    int64 `json:"amount"`
}