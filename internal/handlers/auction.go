package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	js "github.com/adk2004/auction_service/internal/json"
	"github.com/adk2004/auction_service/internal/services"
	"github.com/adk2004/auction_service/internal/types"
	"github.com/go-chi/chi/v5"
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
	if err := js.Read(r, &aucRq); err != nil {
		js.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if aucRq.BasePrice <= 0 || len(aucRq.Title) == 0 {
		js.WriteError(w, http.StatusBadRequest, "Invalid json payload")
		return
	}
	auctionId, err := h.aucSvc.CreateAuction(r.Context(), aucRq.Title, aucRq.BasePrice, aucRq.OwnerId)
	if err != nil {
		js.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	js.WriteJSON(w, http.StatusOK, map[string]int64{"auctionId": auctionId})
}

func (h *AuctionHandler) PostBid(w http.ResponseWriter, r *http.Request) {
	var bid types.PostBidRequest
	if err := js.Read(r, &bid); err != nil {
		js.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if bid.Amount <= 0 {
		js.WriteError(w, http.StatusBadRequest, "Invalid bid amount")
		return
	}
	bidUpdates, err := h.aucSvc.CreateBid(r.Context(), bid)

	if err != nil {
		js.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		js.WriteError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	fmt.Fprintf(w, "data: %s\n\n", "queued")
	flusher.Flush()
	for update := range bidUpdates {
		fmt.Fprintf(w, "data: %s\n\n", update)
		flusher.Flush()
	}
}

func (h *AuctionHandler) GetAuction(w http.ResponseWriter, r *http.Request) {

	id := chi.URLParam(r, "id")
	aucId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		js.WriteError(w, http.StatusBadRequest, "Invalid auction ID")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		js.WriteError(w, http.StatusInternalServerError, "Streaming not supported")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	ch := h.aucSvc.SubAuction(r.Context(), int64(aucId))
	fmt.Fprintf(w, "data: %s\n\n", "subscribed")
	flusher.Flush()

	for {
		select {
		case event, ok := <-ch:
			if !ok {
				fmt.Fprintf(w, "data: %s\n\n", "subscription closed")
				flusher.Flush()
				return
			}
			jsonEvent, err := json.Marshal(event)
			if err != nil {
				js.WriteError(w, http.StatusInternalServerError, "Failed to marshal event")
				return
			}
			if event.Completed {
				fmt.Fprintf(w, "data: %s\n\n", "auction closed")
				flusher.Flush()
				h.aucSvc.UnsubAuction(int64(aucId), ch)
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", string(jsonEvent))
			flusher.Flush()
		case <-r.Context().Done():
			h.aucSvc.UnsubAuction(int64(aucId), ch)
			fmt.Fprintf(w, "data: %s\n\n", "unsubscribed")
			flusher.Flush()
			return
		}
	}

}
