# Job Portal - Backend

A robust Go REST API for AI-powered job matching with blockchain payment verification and JWT authentication.

## LIVE DEMO - https://job-portal-assg.netlify.app/
## Features

- **User Management**: Secure registration, login, and profile management
- **Job Posting**: Create and list jobs with blockchain payment verification
- **AI Skill Matching**: Extract skills from text and compute job match scores
- **Authentication**: JWT-based token security
- **Blockchain Integration**: Verify Sepolia ETH transactions for job posting fees
- **Database**: PostgreSQL with connection pooling
- **API**: RESTful endpoints with proper HTTP status codes
- **CORS**: Configured for frontend development and production

## Tech Stack

- **Language**: Go
- **Framework**: Fiber v2 (lightweight, fast HTTP framework)
- **Database**: PostgreSQL (Neon)
- **ORM**: pgx (native Go PostgreSQL driver with type-safe queries)
- **Authentication**: JWT (github.com/golang-jwt/jwt/v5)
- **AI/LLM**: OpenAI API integration for skill extraction
- **Blockchain**: Ethereum/Sepolia transaction validation
- **Utilities**: uuid, bcrypt for password hashing

## Getting Started

### Prerequisites

- Go
- Neon PostgreSQL DB
- Environment variables configured
- Gemini API key (for AI skill extraction)

### Installation

```bash
cd Job-portal-backend
go mod download
```

### Environment Setup

Create `.env` file with the following variables (copy from `.env.example`):

```env
# Server Configuration
PORT=8080
FRONTEND_URL=http://localhost:5173

# Database Configuration
DATABASE_URL=postgres://user:password@localhost:5432/job_portal

# JWT Security
JWT_SECRET=your_secure_key_at_least_32_chars_change_in_production

# Google Gemini API
GEMINI_API_KEY=your_api_key_here
```

**Development**: Use values in `.env` for local testing
**Production**: Update all values to production/staging endpoints and secrets

## Environment Variables Reference

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `PORT` | No | 8080 | HTTP server port |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret key for JWT signing |
| `FRONTEND_URL` | No | http://localhost:5173 | Frontend URL for CORS |
| `GEMINI_API_KEY` | Yes | - | Google Gemini AI API key |

## Deployment

### Deployment Checklist

1. **Database Setup**:
   - Create PostgreSQL instance (managed service or self-hosted)
   - Set `DATABASE_URL` to production database
   - Run migrations: `psql $DATABASE_URL < migrations/database.sql`
   - Enable automated backups

2. **Secrets Management**:
   - Generate strong `JWT_SECRET` with: `openssl rand -base64 32`
   - Store in secure vault (AWS Secrets Manager, HashiCorp Vault, etc.)
   - Never commit secrets to git

3. **Environment Variables**:
   - Set `FRONTEND_URL` to production frontend domain
   - Configure `GEMINI_API_KEY` for AI features
   - Update `PORT` if needed (default 8080)

4. **Build & Deploy**:
   - Build binary: `go build -o job-portal-server main.go`
   - Or use Docker image (see below)

5. **CORS Configuration**:
   - Update backend `FRONTEND_URL` to match frontend domain
   - Backend automatically uses it in CORS headers

6. **Monitoring**:
   - Enable application logs
   - Monitor database connection pool
   - Set up alerts for API errors

## Project Structure

```
.
├── main.go                    # Server entry point, route setup
├── go.mod, go.sum            # Dependency management
├── internal/
│   ├── config/               # Configuration loading
│   │   └── config.go        # Env var parsing
│   ├── db/                  # Database connectivity
│   │   └── db.go            # PostgreSQL connection pool
│   ├── handlers/            # HTTP request handlers
│   │   ├── auth_handler.go  # Login/signup endpoints
│   │   ├── job_handler.go   # Job CRUD operations
│   │   ├── profile_handler.go# User profile management
│   │   └── ai_handler.go    # AI skill extraction
│   ├── middleware/          # HTTP middleware
│   │   └── auth_middleware.go# JWT validation
│   ├── models/              # Data structures
│   │   ├── user.go          # User model
│   │   └── job.go           # Job model
│   └── services/            # Business logic
│       ├── auth_service.go  # User registration/login logic
│       ├── job_service.go   # Job creation & retrieval
│       └── ai_service.go    # AI integration for skill extraction
├── pkg/                     # Public utilities
│   └── utils/
│       ├── jwt.go           # JWT token generation/validation
│       └── hash.go          # Password hashing with bcrypt
├── migrations/              # Database schemas
│   └── database.sql         # Initial schema & tables
└── README.md               # This file
```

## Authentication

### JWT Flow

1. User registers with email/password
2. Password hashed with bcrypt
3. Login returns JWT token
4. Token contains user ID + issue/expiry timestamps
5. Token included in `Authorization: Bearer <token>` header
6. Middleware validates token on protected routes

### Protected Routes

Require valid JWT token in Authorization header:

- `GET /me` - Current user info
- `PUT /profile` - Update user profile
- `POST /jobs` - Create job
- `GET /jobs/:id` - Get job details with match score
- `POST /ai/extract-skills` - Extract skills with AI

## Blockchain Payment Verification

### Payment Flow

1. Frontend sends job data + blockchain transaction hash
2. Backend validates transaction hash format (66 chars, 0x prefix)
3. Hash stored with job for audit trail
4. **Note**: Full on-chain validation not implemented (future enhancement)

### Transaction Hash Validation

```go
// Validates format: 0x + 64 hex characters
if len(paymentTx) != 66 || paymentTx[:2] != "0x" {
    return "", errors.New("invalid transaction hash format")
}
```

## AI Features

### Skill Extraction

**Endpoint**: `POST /ai/extract-skills`

Takes resume/bio text and returns extracted skills using OpenAI's GPT models.

**Request**:
```json
{
  "bio": "I have 5 years experience in React, Node.js, and PostgreSQL..."
}
```

**Response**:
```json
{
  "skills": ["React", "Node.js", "PostgreSQL", "REST APIs"]
}
```

### Job Match Scoring

**Endpoint**: `GET /jobs/:id`

Computes match score between user's skills and job description using AI.

**Response includes**:
```json
{
  "job": { ... },
  "match_score": 85
}
```

## Database Schema

### users

```sql
CREATE TABLE users (
  id UUID PRIMARY KEY,
  name VARCHAR NOT NULL,
  email VARCHAR UNIQUE NOT NULL,
  password_hash VARCHAR NOT NULL,
  bio VARCHAR,
  linkedin_url VARCHAR,
  skills JSONB DEFAULT 'null',
  wallet_address VARCHAR,
  created_at TIMESTAMP
);
```

### jobs

```sql
CREATE TABLE jobs (
  id UUID PRIMARY KEY,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  skills JSONB DEFAULT 'null',
  salary VARCHAR,
  location VARCHAR,
  user_id UUID REFERENCES users(id),
  payment_tx_hash VARCHAR,
  created_at TIMESTAMP
);
```

## API Endpoints

### Authentication
- `POST /auth/register` - Create account
- `POST /auth/login` - Login

### Profile
- `GET /profile/:id` - Get user profile (public)
- `GET /me` - Current user profile (protected)
- `PUT /profile` - Update profile (protected)

### Jobs
- `GET /jobs` - List jobs (public)
- `POST /jobs` - Create job (protected)
- `GET /jobs/:id` - Get job with match score (protected)

### AI
- `POST /ai/extract-skills` - Extract skills from text (protected)

## Service Layer Architecture

### auth_service.go
- `RegisterUser(email, password)` - Create new user
- `LoginUser(email, password)` - Authenticate user
- `GetUserByID(id)` - Fetch user profile
- `UpdateUser(id, updates)` - Modify user data

### job_service.go
- `CreateJob(...)` - Create job posting
- `ListJobs(limit)` - Fetch recent jobs
- `GetJobByID(id)` - Fetch single job
- `ComputeMatchScore(skills, description)` - AI match calculation

### ai_service.go
- `ExtractSkills(text)` - Parse skills from resume using OpenAI
- `ComputeMatchScore(...)` - Score job/user alignment

## Middleware

### AuthRequired()

Validates JWT token and extracts user ID for protected routes.

```go
protected := app.Group("", middleware.AuthRequired())
protected.Post("/jobs", handlers.CreateJob)
```

## Code Patterns

### Error Handling

```go
if err != nil {
  return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
    "error": err.Error(),
  })
}
```

### Database Queries

```go
err := db.Pool.QueryRow(context.Background(), 
  "SELECT * FROM users WHERE id=$1", userID).
  Scan(&id, &name, ...)
```

### Request Parsing

```go
var req createJobRequest
if err := c.BodyParser(&req); err != nil {
  return c.Status(fiber.StatusBadRequest).JSON(...)
}
```

## Performance Optimization

- Connection pooling with pgx
- JWT caching (no database lookup per request)
- Indexed queries on user.email, jobs.created_at
- Request logging with Fiber middleware
- CORS preflight caching

## Security Notes

- Passwords hashed with bcrypt (cost 10)
- JWT secrets must be 32+ characters
- SQL injection prevented with parameterized queries
- CORS restricted to frontend domain
- Blockchain transactions verified before database insert
- User IDs from JWT token (not user input)