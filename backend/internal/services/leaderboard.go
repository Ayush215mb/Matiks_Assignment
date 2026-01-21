// internal/services/leaderboard.go
package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"backend/internal/models"
	"backend/pkg/store"
)

type LeaderboardService struct {
	store *store.MemoryStore
}

func NewLeaderboardService(store *store.MemoryStore) *LeaderboardService {
	return &LeaderboardService{store: store}
}

// SeedData seeds the leaderboard with random users
func (s *LeaderboardService) SeedData(ctx context.Context, count int) error {
	log.Printf("Seeding %d users...", count)

	for i := 0; i < count; i++ {
		username := fmt.Sprintf("user_%d", i+1)
		rating := rand.Intn(4901) + 100 // Random rating between 100 and 5000

		if err := s.store.AddUser(username, rating); err != nil {
			return fmt.Errorf("failed to add user: %w", err)
		}

		if (i+1)%1000 == 0 {
			log.Printf("Seeded %d users...", i+1)
		}
	}

	log.Printf("✓ Successfully seeded %d users", count)
	return nil
}

// GetLeaderboard retrieves paginated leaderboard with correct ranks
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, page, limit int) (*models.LeaderboardResponse, error) {
	// Get all users sorted
	allUsers := s.store.GetAllUsers()
	total := len(allUsers)

	// Calculate pagination
	offset := (page - 1) * limit
	if offset >= total {
		return &models.LeaderboardResponse{
			Entries:    []models.LeaderboardEntry{},
			Page:       page,
			Limit:      limit,
			TotalUsers: int64(total),
			HasMore:    false,
		}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	// Calculate ranks considering ties
	entries := make([]models.LeaderboardEntry, 0, end-offset)
	currentRank := 1

	for i := 0; i < end; i++ {
		// Update rank when score changes
		if i > 0 && allUsers[i].Rating != allUsers[i-1].Rating {
			currentRank = i + 1
		}

		// Only add entries within the requested page
		if i >= offset {
			entries = append(entries, models.LeaderboardEntry{
				Rank:     currentRank,
				Username: allUsers[i].Username,
				Rating:   allUsers[i].Rating,
			})
		}
	}

	hasMore := end < total

	return &models.LeaderboardResponse{
		Entries:    entries,
		Page:       page,
		Limit:      limit,
		TotalUsers: int64(total),
		HasMore:    hasMore,
	}, nil
}

// GetUserRank retrieves a specific user's rank
func (s *LeaderboardService) GetUserRank(ctx context.Context, username string) (*models.UserRankResponse, error) {
	user, err := s.store.GetUser(username)
	if err != nil {
		return nil, err
	}

	rank, err := s.store.GetUserRank(username)
	if err != nil {
		return nil, err
	}

	return &models.UserRankResponse{
		Username: username,
		Rating:   user.Rating,
		Rank:     int64(rank),
	}, nil
}

// UpdateScore updates a user's score
func (s *LeaderboardService) UpdateScore(ctx context.Context, username string, newRating int) error {
	// Check if user exists
	user, err := s.store.GetUser(username)
	if err != nil {
		return err
	}

	oldRating := user.Rating

	// Update score
	if err := s.store.AddUser(username, newRating); err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	log.Printf("Updated %s: %d -> %d", username, oldRating, newRating)
	return nil
}

// SearchUser searches for users by username prefix
func (s *LeaderboardService) SearchUser(ctx context.Context, query string) ([]models.UserRankResponse, error) {
	users := s.store.SearchUsers(query, 50)

	results := make([]models.UserRankResponse, 0, len(users))
	for _, user := range users {
		rank, _ := s.store.GetUserRank(user.Username)
		results = append(results, models.UserRankResponse{
			Username: user.Username,
			Rating:   user.Rating,
			Rank:     int64(rank),
		})
	}

	return results, nil
}

// GetStats returns leaderboard statistics
func (s *LeaderboardService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	total, minRating, maxRating, avgRating := s.store.GetStats()

	return &models.StatsResponse{
		TotalUsers:    int64(total),
		MinRating:     float64(minRating),
		MaxRating:     float64(maxRating),
		AverageRating: avgRating,
	}, nil
}

// StartRandomUpdates simulates random score updates
func (s *LeaderboardService) StartRandomUpdates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("🎲 Started random score updates (every 5 seconds)")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count := s.store.GetUserCount()
			if count == 0 {
				continue
			}

			// Get random user
			randomIndex := rand.Intn(count)
			allUsers := s.store.GetAllUsers()
			if randomIndex >= len(allUsers) {
				continue
			}

			username := allUsers[randomIndex].Username
			newRating := rand.Intn(4901) + 100

			if err := s.UpdateScore(ctx, username, newRating); err != nil {
				log.Printf("Failed to update random score: %v", err)
			}
		}
	}
}
