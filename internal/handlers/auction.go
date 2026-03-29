package handlers

import (
	"net/http"

	"github.com/adk2004/auction_service/internal/json"
	"github.com/adk2004/auction_service/internal/services"
)

type PostAuctionRequest struct {
	OwnerId int64 `json:"ownerId"`
	BasePrice int64 `json:"basePrice"`
	Title string `json:"title"`
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
	if err:= json.Read(r, &aucRq); err!= nil {
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if aucRq.BasePrice <= 0 || len(aucRq.Title) == 0 {
		json.WriteError(w, http.StatusBadRequest, "Invalid json payload")
		return
	}
	auctionId, err := h.aucSvc.CreateAuction(r.Context(), aucRq.Title, aucRq.BasePrice, aucRq.OwnerId)
	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.WriteJSON(w, http.StatusOK, map[string]int64{"auctionId": auctionId})
}

func (h *AuctionHandler) PostBid(w http.ResponseWriter, r *http.Request) {
	var bid services.Bid
	if err:= json.Read(r, &bid); err!= nil {
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if bid.Amount <= 0 {
		json.WriteError(w, http.StatusBadRequest, "Invalid bid amount")
		return
	}
	err := h.aucSvc.CreateBid(r.Context(), bid)
	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	json.WriteJSON(w, http.StatusOK, map[string]string{"message": "Bid placed successfully"})
}