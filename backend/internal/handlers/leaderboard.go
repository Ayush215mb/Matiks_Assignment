package handlers

import (
	"net/http"
	"strconv"

	"backend/internal/models"
	"backend/internal/services"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	service *services.LeaderboardService
}

func NewLeaderboardHandler(service *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{service: service}
}

// SeedData seeds the leaderboard with users
// POST /api/seed
func (h *LeaderboardHandler) SeedData(c *gin.Context) {
	var req models.SeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.SeedData(c.Request.Context(), req.Count); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "seed_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data seeded successfully",
		"count":   req.Count,
	})
}

// GetLeaderboard retrieves paginated leaderboard
// GET /api/leaderboard?page=1&limit=50
func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	leaderboard, err := h.service.GetLeaderboard(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, leaderboard)
}

// GetUserRank retrieves a specific user's rank
// GET /api/users/:username
func (h *LeaderboardHandler) GetUserRank(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_username",
			Message: "Username is required",
		})
		return
	}

	userRank, err := h.service.GetUserRank(c.Request.Context(), username)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "user_not_found",
				Message: "User does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "fetch_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, userRank)
}

// UpdateScore updates a user's score
// POST /api/users/:username/score
func (h *LeaderboardHandler) UpdateScore(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_username",
			Message: "Username is required",
		})
		return
	}

	var req models.UpdateScoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdateScore(c.Request.Context(), username, req.Rating); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "user_not_found",
				Message: "User does not exist",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "update_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Score updated successfully",
	})
}

// SearchUser searches for users
// GET /api/search?q=user_123
func (h *LeaderboardHandler) SearchUser(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_query",
			Message: "Search query is required",
		})
		return
	}

	results, err := h.service.SearchUser(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "search_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// GetStats retrieves leaderboard statistics
// GET /api/stats
func (h *LeaderboardHandler) GetStats(c *gin.Context) {
	stats, err := h.service.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "stats_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
