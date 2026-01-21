package store

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// User represents a user in the leaderboard
type User struct {
	Username string
	Rating   int
}

// MemoryStore is an in-memory leaderboard store
type MemoryStore struct {
	mu    sync.RWMutex
	users map[string]*User // username -> User
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		users: make(map[string]*User),
	}
}

// AddUser adds or updates a user
func (s *MemoryStore) AddUser(username string, rating int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[username] = &User{
		Username: username,
		Rating:   rating,
	}
	return nil
}

// GetUser retrieves a user by username
func (s *MemoryStore) GetUser(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// GetAllUsers returns all users sorted by rating (descending)
func (s *MemoryStore) GetAllUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	// Sort by rating descending, then by username ascending (for stable sort)
	sort.Slice(users, func(i, j int) bool {
		if users[i].Rating != users[j].Rating {
			return users[i].Rating > users[j].Rating
		}
		return users[i].Username < users[j].Username
	})

	return users
}

// GetUserCount returns total number of users
func (s *MemoryStore) GetUserCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users)
}

// GetUserRank calculates a user's rank (handles ties correctly)
func (s *MemoryStore) GetUserRank(username string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return 0, fmt.Errorf("user not found")
	}

	rank := 1
	for _, u := range s.users {
		if u.Rating > user.Rating {
			rank++
		}
	}

	return rank, nil
}

// SearchUsers searches for users by username prefix
func (s *MemoryStore) SearchUsers(query string, limit int) []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	results := make([]*User, 0)

	for _, user := range s.users {
		if strings.Contains(strings.ToLower(user.Username), query) {
			results = append(results, user)
			if len(results) >= limit {
				break
			}
		}
	}

	// Sort results by rating
	sort.Slice(results, func(i, j int) bool {
		if results[i].Rating != results[j].Rating {
			return results[i].Rating > results[j].Rating
		}
		return results[i].Username < results[j].Username
	})

	return results
}

// GetStats calculates leaderboard statistics
func (s *MemoryStore) GetStats() (total int, minRating, maxRating int, avgRating float64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.users) == 0 {
		return 0, 0, 0, 0
	}

	total = len(s.users)
	minRating = 5000
	maxRating = 100
	sum := 0

	for _, user := range s.users {
		if user.Rating < minRating {
			minRating = user.Rating
		}
		if user.Rating > maxRating {
			maxRating = user.Rating
		}
		sum += user.Rating
	}

	avgRating = float64(sum) / float64(total)
	return
}

// Clear removes all users
func (s *MemoryStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users = make(map[string]*User)
}
