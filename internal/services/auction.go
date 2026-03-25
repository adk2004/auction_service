package services

import (
	"context"

	repo "github.com/adk2004/auction_service/internal/adapters/postgresql/sqlc"
)

type AuctionService interface {
	CreateAuction(ctx context.Context, title string, basePrice int64, ownerId int64) (int64, error)
}

type aucService struct {
	repo repo.Querier
}

func NewAuctionService(repo repo.Querier) AuctionService {
	return &aucService{
		repo: repo,
	}
}

func (s *aucService) CreateAuction(ctx context.Context, title string, basePrice int64, ownerId int64) (int64, error) {
	return  s.repo.CreateAuction(ctx, repo.CreateAuctionParams{
		Title:      title,
		Baseprice:  basePrice,
		Highestbid: basePrice,
		State:      repo.AuctionStatusOngoing,
		Ownerid:    ownerId,
	})
}