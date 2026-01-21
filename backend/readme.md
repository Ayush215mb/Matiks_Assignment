# Leaderboard Platform - Backend

A high-performance leaderboard system built with Go, designed to handle millions of users with efficient ranking, search, and real-time updates using in-memory storage.

## 🏗️ Project Structure

```
leaderboard-backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/
│   │   └── leaderboard.go       # HTTP request handlers
│   ├── services/
│   │   └── leaderboard.go       # Business logic
│   └── models/
│       └── models.go            # Data models
├── pkg/
│   └── store/
│       └── memory.go            # In-memory storage with sync.RWMutex
├── .env                         # Environment variables
├── go.mod                       # Go dependencies
├── go.sum                       # Dependency checksums
├── Makefile                     # Build commands
└── README.md                    # This file
```

## 🔧 Installation

1. **Clone the repository**
```bash
git clone <your-repo>
cd leaderboard-backend
```

2. **Install dependencies**
```bash
go mod download
```

3. **Run the application**
```bash
go run cmd/server/main.go
# Or use Make
make run
```

The server will start on `http://localhost:8080`

## 📡 API Endpoints

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "store": "in-memory"
}
```

### Seed Data
```http
POST /api/seed
Content-Type: application/json

{
  "count": 10000
}
```

**Response:**
```json
{
  "message": "Data seeded successfully",
  "count": 10000
}
```

### Get Leaderboard
```http
GET /api/leaderboard?page=1&limit=50
```

**Response:**
```json
{
  "entries": [
    {
      "rank": 1,
      "username": "user_123",
      "rating": 4950
    }
  ],
  "page": 1,
  "limit": 50,
  "total_users": 10000,
  "has_more": true
}
```

### Get User Rank
```http
GET /api/users/:username
```

**Response:**
```json
{
  "username": "user_123",
  "rating": 4950,
  "rank": 1
}
```

### Update User Score
```http
POST /api/users/:username/score
Content-Type: application/json

{
  "rating": 4500
}
```

**Response:**
```json
{
  "message": "Score updated successfully"
}
```

### Search Users
```http
GET /api/search?q=user_123
```

**Response:**
```json
{
  "results": [
    {
      "username": "user_123",
      "rating": 4500,
      "rank": 15
    },
    {
      "username": "user_1234",
      "rating": 3200,
      "rank": 850
    }
  ],
  "count": 2
}
```

### Get Statistics
```http
GET /api/stats
```

**Response:**
```json
{
  "total_users": 10000,
  "min_rating": 100,
  "max_rating": 5000,
  "average_rating": 2550.5
}
```
