package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	repo "github.com/adk2004/auction_service/internal/adapters/postgresql/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Bid struct {
	UserId    int64 `json:"userId"`
	AuctionId int64 `json:"auctionId"`
	Amount    int64 `json:"amount"`
}

// stores auctionId to mutex mapping for synchronizing bid placements on the same auction
type bidManager struct {
	chans map[int64]chan Bid
	mu    sync.RWMutex
}

type AuctionService interface {
	CreateAuction(ctx context.Context, title string, basePrice int64, ownerId int64) (int64, error)
	CreateBid(ctx context.Context, bid Bid) error
	Cancel()
}

type aucService struct {
	repo     *repo.Queries
	bm       *bidManager
	db       *pgx.Conn
	ctx      context.Context
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
}

func NewAuctionService(repo *repo.Queries, db *pgx.Conn) (AuctionService, error) {
	ctx, cancel := context.WithCancel(context.Background())
	aucService := &aucService{
		repo: repo,
		bm: &bidManager{
			chans: make(map[int64]chan Bid),
		},
		db:       db,
		ctx:      ctx,
		cancelFn: cancel,
	}

	ids, err := aucService.repo.GetOngoingAuctions(aucService.ctx)
	if err != nil {
		log.Printf("Error fetching ongoing auctions: %v", err)
		return aucService, err
	}
	for _, id := range ids {
		aucService.bm.chans[id] = make(chan Bid, 1000)
		aucService.wg.Add(1)
		go aucService.placeBids( aucService.bm.chans[id])
	}
	go aucService.closeAuctions(aucService.ctx)
	return aucService, nil
}

func (s *aucService) Cancel() {
	s.cancelFn()
	s.bm.mu.Lock()
	for auctionId, bids := range s.bm.chans {
		close(bids)
		delete(s.bm.chans, auctionId)
	}
	s.bm.mu.Unlock()
	s.wg.Wait()
}

func (s *aucService) CreateAuction(ctx context.Context, title string, basePrice int64, ownerId int64) (int64, error) {
	auctionId, err := s.repo.CreateAuction(ctx, repo.CreateAuctionParams{
		Title:      title,
		Baseprice:  basePrice,
		Highestbid: basePrice,
		State:      repo.AuctionStatusOngoing,
		Ownerid:    ownerId,
	})
	if err != nil {
		return 0, fmt.Errorf("error creating auction: %w", err)
	}
	s.bm.mu.Lock()
	defer s.bm.mu.Unlock()
	bids := make(chan Bid, 1000)
	s.wg.Add(1)
	go s.placeBids( bids)
	s.bm.chans[auctionId] = bids
	return auctionId, err
}

func (s *aucService) CreateBid(ctx context.Context, bid Bid) error {
	// dont queue if ctx is cancelled
	select {
    case <-ctx.Done():
        return fmt.Errorf("request cancelled: %w", ctx.Err())
    default:
    }
	s.bm.mu.RLock()
	defer s.bm.mu.RUnlock()
	bids, ok := s.bm.chans[bid.AuctionId]
	if !ok {
		return errors.New("auction Id not found")
	}
	select {
	case bids <- bid:
		log.Printf("Bid placed: AuctionId=%d, UserId=%d, Amount=%d", bid.AuctionId, bid.UserId, bid.Amount)
	default:
		return errors.New("bid channel is full, try again later")
	}
	return nil
}

func (s *aucService) placeBids(bids chan Bid) {
	defer s.wg.Done()
	for bid := range bids {
		log.Printf("Processing bid: AuctionId=%d, UserId=%d, Amount=%d", bid.AuctionId, bid.UserId, bid.Amount)
		if err := s.processBid(context.Background(), bid); err != nil {
			log.Printf("Error processing bid: %v", err)
		}
	}
}

func (s *aucService) processBid(ctx context.Context, bid Bid) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.repo.WithTx(tx)

	auction, err := qtx.GetAuctionByID(ctx, bid.AuctionId)
	if err != nil {
		return fmt.Errorf("error fetching auction: %w", err)
	}

	if auction.State != repo.AuctionStatusOngoing || bid.Amount <= auction.Highestbid {
		log.Printf("Bid rejected: AuctionId=%d, UserId=%d, Amount=%d", bid.AuctionId, bid.UserId, bid.Amount)
		return nil
	}

	err = qtx.UpdateAuctionWinner(ctx, repo.UpdateAuctionWinnerParams{
		ID:         bid.AuctionId,
		Highestbid: bid.Amount,
		Winnerid:   pgtype.Int8{Int64: bid.UserId, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("error updating auction winner: %w", err)
	}

	err = qtx.PlaceBid(ctx, repo.PlaceBidParams{
		Auctionid: bid.AuctionId,
		Bidderid:  bid.UserId,
		Amount:    bid.Amount,
	})
	if err != nil {
		return fmt.Errorf("error saving bid: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Printf("Bid committed successfully: AuctionId=%d, UserId=%d, Amount=%d", bid.AuctionId, bid.UserId, bid.Amount)
	return nil
}

func (s *aucService) closeAuctions(ctx context.Context) {
	// crude approach - like cron job
	// poll the db evry minute to find auctions that have not benn updated for more than 5 minutes and close them
	// get all auctions that are ongoing
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("closeAuctions shutting down")
			return
		case <-ticker.C:
			s.processExpiredAuctions(ctx)
		}
	}
}

func (s *aucService) processExpiredAuctions(ctx context.Context) {
	auctions, err := s.repo.GetCompletedAuctions(ctx)
	if err != nil {
		log.Printf("Error fetching completed auctions: %v", err)
		return
	}
	for _, auc := range auctions {
		// close this auction
		err := s.repo.EndAuction(ctx, auc.ID)
		if err != nil {
			log.Printf("Error ending auction %d: %v", auc.ID, err)
			continue
		}
		// close bid channel and delete from map
		s.closeBidChannel(auc.ID)
		log.Printf("Auction closed: AuctionId=%d", auc.ID)
	}

}

func (s *aucService) closeBidChannel(auctionId int64) {
	s.bm.mu.Lock()
	defer s.bm.mu.Unlock()
	if bids, ok := s.bm.chans[auctionId]; ok {
		close(bids)
		delete(s.bm.chans, auctionId)
	}
}
