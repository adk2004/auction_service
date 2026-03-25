package handlers

import (
	"net/http"

	"github.com/adk2004/auction_service/internal/json"
	"github.com/adk2004/auction_service/internal/services"
)

type PostAuctionRequest struct {
	ownerId int64
	basePrice int64
	title string
}

type PostBidRequest struct {
	userId int64
	auctionId int64
	amount int64
}

type AuctionHandler struct {
	aucSvc services.AuctionService
}

func NewAuctionHandler(aucSvc services.AuctionService) *AuctionHandler{
	return &AuctionHandler{
		aucSvc: aucSvc,
	}
}

func (h *AuctionHandler) PostAuction(w http.ResponseWriter, r *http.Request) {
	var aucRq PostAuctionRequest
	if err:= json.Read(r, aucRq); err!= nil {
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if aucRq.basePrice <= 0 || len(aucRq.title) == 0 {
		json.WriteError(w, http.StatusBadRequest, "Invalid json payload")
		return
	}
	h.aucSvc.CreateAuction(r.Context(), aucRq.title, aucRq.basePrice, aucRq.ownerId)
}

func (h *AuctionHandler) PostBid(w http.ResponseWriter, r *http.Request) {

}