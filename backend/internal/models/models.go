package models

// User represents a user in the leaderboard
type User struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

// LeaderboardEntry represents an entry in the leaderboard with rank
type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

// LeaderboardResponse represents the paginated leaderboard response
type LeaderboardResponse struct {
	Entries    []LeaderboardEntry `json:"entries"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalUsers int64              `json:"total_users"`
	HasMore    bool               `json:"has_more"`
}

// UserRankResponse represents a user's rank information
type UserRankResponse struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int64  `json:"rank"`
}

// UpdateScoreRequest represents a request to update user score
type UpdateScoreRequest struct {
	Rating int `json:"rating" binding:"required,min=100,max=5000"`
}

// SeedRequest represents a request to seed data
type SeedRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

// StatsResponse represents system statistics
type StatsResponse struct {
	TotalUsers    int64   `json:"total_users"`
	MinRating     float64 `json:"min_rating"`
	MaxRating     float64 `json:"max_rating"`
	AverageRating float64 `json:"average_rating"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
