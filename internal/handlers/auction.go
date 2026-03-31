package handlers

import (
	"fmt"
	"net/http"

	"github.com/adk2004/auction_service/internal/json"
	"github.com/adk2004/auction_service/internal/services"
	"github.com/adk2004/auction_service/internal/types"
)

type AuctionHandler struct {
	aucSvc services.AuctionService
}

func NewAuctionHandler(aucSvc services.AuctionService) *AuctionHandler {
	return &AuctionHandler{
		aucSvc: aucSvc,
	}
}

func (h *AuctionHandler) PostAuction(w http.ResponseWriter, r *http.Request) {
	var aucRq types.PostAuctionRequest
	if err := json.Read(r, &aucRq); err != nil {
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
	var bid types.PostBidRequest
	if err := json.Read(r, &bid); err != nil {
		json.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if bid.Amount <= 0 {
		json.WriteError(w, http.StatusBadRequest, "Invalid bid amount")
		return
	}
	bidUpdates, err := h.aucSvc.CreateBid(r.Context(), bid)

	if err != nil {
		json.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		json.WriteError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", "queued")
	flusher.Flush()
	for update := range bidUpdates {
		fmt.Fprintf(w, "data: %s\n\n", update)
		flusher.Flush()
	}
}
