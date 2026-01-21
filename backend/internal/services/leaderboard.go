package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"backend/internal/models"
	redisClient "backend/pkg/redis"

	"github.com/redis/go-redis/v9"
)

type LeaderboardService struct {
	redis *redis.Client
}

func NewLeaderboardService(redis *redis.Client) *LeaderboardService {
	return &LeaderboardService{redis: redis}
}

// SeedData seeds the leaderboard with random users
func (s *LeaderboardService) SeedData(ctx context.Context, count int) error {
	log.Printf("Seeding %d users...", count)

	// Use pipeline for batch operations
	pipe := s.redis.Pipeline()

	for i := 0; i < count; i++ {
		username := fmt.Sprintf("user_%d", i+1)
		rating := rand.Intn(4901) + 100 // Random rating between 100 and 5000

		// Add to sorted set (leaderboard)
		pipe.ZAdd(ctx, redisClient.LeaderboardKey, redis.Z{
			Score:  float64(rating),
			Member: username,
		})

		// Execute in batches of 1000 for better performance
		if (i+1)%1000 == 0 {
			if _, err := pipe.Exec(ctx); err != nil {
				return fmt.Errorf("failed to seed batch: %w", err)
			}
			pipe = s.redis.Pipeline()
			log.Printf("Seeded %d users...", i+1)
		}
	}

	// Execute remaining operations
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to seed final batch: %w", err)
	}

	log.Printf("✓ Successfully seeded %d users", count)
	return nil
}

// GetLeaderboard retrieves paginated leaderboard with correct ranks
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, page, limit int) (*models.LeaderboardResponse, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Get total count
	total, err := s.redis.ZCard(ctx, redisClient.LeaderboardKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get users with scores in descending order
	users, err := s.redis.ZRevRangeWithScores(ctx, redisClient.LeaderboardKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	// Calculate ranks considering ties
	entries := make([]models.LeaderboardEntry, 0, len(users))

	for i, user := range users {
		var rank int

		if i == 0 {
			// For first user, calculate rank based on users with higher scores
			count, err := s.redis.ZCount(ctx, redisClient.LeaderboardKey,
				fmt.Sprintf("(%f", user.Score), "+inf").Result()
			if err != nil {
				return nil, err
			}
			rank = int(count) + 1
		} else {
			// If same score as previous, same rank
			if users[i].Score == users[i-1].Score {
				rank = entries[i-1].Rank
			} else {
				// Calculate rank for new score
				count, err := s.redis.ZCount(ctx, redisClient.LeaderboardKey,
					fmt.Sprintf("(%f", user.Score), "+inf").Result()
				if err != nil {
					return nil, err
				}
				rank = int(count) + 1
			}
		}

		entries = append(entries, models.LeaderboardEntry{
			Rank:     rank,
			Username: user.Member.(string),
			Rating:   int(user.Score),
		})
	}

	hasMore := int64(offset+limit) < total

	return &models.LeaderboardResponse{
		Entries:    entries,
		Page:       page,
		Limit:      limit,
		TotalUsers: total,
		HasMore:    hasMore,
	}, nil
}

// GetUserRank retrieves a specific user's rank
func (s *LeaderboardService) GetUserRank(ctx context.Context, username string) (*models.UserRankResponse, error) {
	// Get user's score
	score, err := s.redis.ZScore(ctx, redisClient.LeaderboardKey, username).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user score: %w", err)
	}

	// Count users with strictly higher scores
	count, err := s.redis.ZCount(ctx, redisClient.LeaderboardKey,
		fmt.Sprintf("(%f", score), "+inf").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate rank: %w", err)
	}

	rank := count + 1

	return &models.UserRankResponse{
		Username: username,
		Rating:   int(score),
		Rank:     rank,
	}, nil
}

// UpdateScore updates a user's score
func (s *LeaderboardService) UpdateScore(ctx context.Context, username string, newRating int) error {
	// Check if user exists
	exists, err := s.redis.ZScore(ctx, redisClient.LeaderboardKey, username).Result()
	if err == redis.Nil {
		return fmt.Errorf("user not found")
	}
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	// Update score
	_, err = s.redis.ZAdd(ctx, redisClient.LeaderboardKey, redis.Z{
		Score:  float64(newRating),
		Member: username,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to update score: %w", err)
	}

	log.Printf("Updated %s: %.0f -> %d", username, exists, newRating)
	return nil
}

// SearchUser searches for users by username prefix
func (s *LeaderboardService) SearchUser(ctx context.Context, query string) ([]models.UserRankResponse, error) {
	query = strings.ToLower(query)

	// Get all users (for now, in production you'd use a separate index)
	users, err := s.redis.ZRevRangeWithScores(ctx, redisClient.LeaderboardKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	results := make([]models.UserRankResponse, 0)
	currentRank := 1

	for i, user := range users {
		username := user.Member.(string)

		// Calculate rank
		if i > 0 && users[i].Score != users[i-1].Score {
			currentRank = i + 1
		}

		// Check if username matches query
		if strings.Contains(strings.ToLower(username), query) {
			results = append(results, models.UserRankResponse{
				Username: username,
				Rating:   int(user.Score),
				Rank:     int64(currentRank),
			})

			// Limit results
			if len(results) >= 50 {
				break
			}
		}
	}

	return results, nil
}

// GetStats returns leaderboard statistics
func (s *LeaderboardService) GetStats(ctx context.Context) (*models.StatsResponse, error) {
	// Total users
	total, err := s.redis.ZCard(ctx, redisClient.LeaderboardKey).Result()
	if err != nil {
		return nil, err
	}

	// Min rating (lowest score)
	minUsers, err := s.redis.ZRangeWithScores(ctx, redisClient.LeaderboardKey, 0, 0).Result()
	if err != nil || len(minUsers) == 0 {
		return nil, err
	}

	// Max rating (highest score)
	maxUsers, err := s.redis.ZRevRangeWithScores(ctx, redisClient.LeaderboardKey, 0, 0).Result()
	if err != nil || len(maxUsers) == 0 {
		return nil, err
	}

	// Calculate average (sum all scores / count)
	allUsers, err := s.redis.ZRangeWithScores(ctx, redisClient.LeaderboardKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var sum float64
	for _, user := range allUsers {
		sum += user.Score
	}
	avg := sum / float64(total)

	return &models.StatsResponse{
		TotalUsers:    total,
		MinRating:     minUsers[0].Score,
		MaxRating:     maxUsers[0].Score,
		AverageRating: avg,
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
			// Get random user
			count, err := s.redis.ZCard(ctx, redisClient.LeaderboardKey).Result()
			if err != nil || count == 0 {
				continue
			}

			randomIndex := rand.Int63n(count)
			users, err := s.redis.ZRangeWithScores(ctx, redisClient.LeaderboardKey, randomIndex, randomIndex).Result()
			if err != nil || len(users) == 0 {
				continue
			}

			username := users[0].Member.(string)
			newRating := rand.Intn(4901) + 100

			if err := s.UpdateScore(ctx, username, newRating); err != nil {
				log.Printf("Failed to update random score: %v", err)
			}
		}
	}
}
