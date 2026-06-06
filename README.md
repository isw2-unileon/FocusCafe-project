# FocusCafe
**Where studying fuels progress.**

## The idea
FocusCafe is a gamified web platform where every user owns a virtual café. 

**No studying. No energy. No growth.**

Users can earn "energy" by completing focused study sessions. After each session, an AI generates questions based on the selected topic to verify real understanding. Correct answers reward energy, which can then be used to fulfill customer orders and level up.

If you run out of "energy", the only way to get back into the game is by studying.

## How it works
- Users upload their own study materials
- AI generates comprehension questions
- Energy-based progression system
- Leveling and gamified growth
- **Real-time collaborative café rooms** — multiple users join a team, share group orders, and receive instant WebSocket notifications when orders are completed or the group is deleted
- **Global leaderboard** — top 5 users ranked by XP, with personal rank visibility

## Built-in Motivation Engine
A global leaderboard turns studying into a competitive challenge. Ranking is based on AI-validated performance.

Users can track their position based on real study performance, not just time logged.

**The more you truly understand, the higher you climb**.

## Educational Purposes
This platform promotes:
- Active recall
- Time-blocked study sessions
- Immediate feedback
- Motivation through gamification
- Collaborative learning dynamics

It transforms long hours of studying into engaging and interactive progress.


---

# Development Guide

A full-stack application with a **Go** backend and a **React + TypeScript + Vite** frontend.

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.24+, Gin, GORM, Gorilla WebSocket, JWT/JWKS (Supabase) |
| **Frontend** | React 18+, TypeScript, Vite, Tailwind CSS, shadcn/ui, Lucide React, react-hot-toast |
| **Auth** | Supabase Auth (email/password + Google OAuth) |
| **AI** | Gemini API (Gemini 2.5 Flash) for quiz generation |
| **Database** | PostgreSQL via Supabase (production), SQLite in-memory (tests) |
| **Testing** | Go table-driven unit tests, Playwright E2E (26 tests passing) |

## Project Structure

```text
├── backend/
│   ├── cmd/
│   │   ├── server/main.go      # Entry point
│   │   └── seed/main.go        # DB seeding
│   └── internal/
│       ├── auth/               # JWT validation (JWKS)
│       ├── config/             # Environment configuration
│       ├── database/           # DB connection & migrations
│       ├── domain/             # Domain entities
│       ├── handlers/           # HTTP handlers (Gin)
│       ├── integration/        # Integration test suite
│       ├── models/             # GORM models
│       ├── repository/         # Data access layer
│       ├── services/           # Business logic
│       ├── supabase/           # Supabase client & JWT adapter
│       └── ws/                 # WebSocket hub and client management
├── frontend/
│   └── src/
│       ├── components/         # Reusable UI components
│       ├── context/            # React contexts (Auth, WebSocket)
│       ├── lib/                # Utilities (JWT, notifications)
│       ├── pages/              # Route pages (Dashboard, Admin, EditProfile, ...)
│       ├── services/           # API client functions
│       └── types/              # TypeScript interfaces
├── docker/
│   ├── Dockerfile.backend      # Container definition for Go API
│   ├── Dockerfile.frontend     # Container definition for React
│   └── nginx.conf              # Reverse proxy configuration
├── docs/
│   ├── adr/                    # Architecture Decision Records
│   ├── images/                 # Screenshots and diagrams
│   └── *.md                    # Guides and documentation
├── e2e/
│   ├── tests/                  # Playwright E2E test specs
│   └── lib/                    # E2E helpers (auth, setup)
├── .github/
│   └── workflows/              # GitHub Actions (CI/CD)
├── supabase/
│   ├── config.toml             # Supabase local config
│   └── migrations/             # SQL migrations
├── scripts/
│   └── seed-local.ps1          # Test data seeding script
├── docker-compose.yml          # Local container orchestration
├── Makefile
└── .env                        # Single env file (no .env.local)
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.25+
- [Node.js](https://nodejs.org/) 22+
- [Supabase CLI](https://supabase.com/docs/guides/cli/getting-started) (for local database)
- A Supabase project (for remote Auth & Database)

## Getting Started

1. Copy `.env.example` to `.env` and fill in the variables (see Environment Variables below).
2. Read `LOCAL_SETUP_GUIDE.md` for detailed Supabase local configuration.
3. Start the local database and dependencies:
   ```bash
   make db-up
   ```
4. Start the backend and frontend in separate terminals:
   ```bash
   # Terminal 1
   make run-backend    # port 8080

   # Terminal 2
   make run-frontend   # port 5173
   ```

The Vite dev server proxies `/api` requests to the backend.

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `SUPABASE_URL` | Supabase project URL | Yes |
| `SUPABASE_KEY` | Supabase anon/public key | Yes |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key (admin ops) | Yes |
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `GEMINI_API_KEY` | Gemini API key (AI quiz generation) | Yes |
| `CLIENT_URL` | Frontend origin (CORS) | Yes |
| `PORT` | Backend server port | Optional (default 8080) |

## Commands

| Command | Description |
|---------|-------------|
| `make install` | Install Go tools, download modules, install npm deps |
| `make install-e2e` | Install Playwright and its dependencies |
| `make run-backend` | Backend with hot reload (Air) |
| `make run-frontend` | Frontend dev server (Vite) |
| `make build-backend` | Build backend binary |
| `make build-frontend` | Build frontend for production |
| `make test` | Run Go + frontend unit tests |
| `make lint` | Run linters (golangci-lint + ESLint) |
| `make e2e` | Run Playwright E2E tests |
| `make db-up` | Start Supabase local + apply migrations + seed |
| `make db-reset` | Reset database + re-apply migrations + seed |
| `make db-down` | Stop Supabase local |
| `make docker-up` | Build and run the stack in Docker |
| `make docker-down` | Stop and remove Docker containers |
| `make docker-up-remote`| Run in Docker connecting to remote Supabase |

## API Endpoints

### Public

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/login` | Email/password login |
| `POST` | `/api/register` | User registration |
| `GET` | `/api/auth/google` | Google OAuth redirect |
| `POST` | `/api/auth/sync` | Sync Google auth user |

### Protected (requires `Authorization: Bearer <jwt>`)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/users/me` | Get current user profile |
| `PUT` | `/api/users/me` | Update profile (first/last name) |
| `GET` | `/api/users/me/orders` | Get pending cafe orders |
| `POST` | `/api/users/me/orders/:id/complete` | Complete order (spend energy, gain XP) |
| `POST` | `/api/study/start` | Start study session |
| `POST` | `/api/study/generate-quiz/:session_id` | Generate AI quiz from material |
| `POST` | `/api/user/progress` | Update gamified stats (energy, XP, level) |

### Leaderboard

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/leaderboard` | Top 5 users by XP (admins excluded) |
| `GET` | `/api/leaderboard/me` | Current user's global rank and profile |

### Groups (requires auth)

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/groups` | Create a new group (leader) |
| `POST` | `/api/groups/join` | Join a group by invite code |
| `POST` | `/api/groups/leave` | Leave current group |
| `DELETE` | `/api/groups` | Delete group (leader only) |

### Admin (requires admin role)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/users` | List all users |
| `GET` | `/api/admin/users/search?email=` | Search user by email |
| `POST` | `/api/admin/users` | Create new user |
| `DELETE` | `/api/admin/users/:id` | Delete user (DB + Supabase Auth) |
| `GET` | `/api/admin/groups` | List all groups with members |
| `DELETE` | `/api/admin/groups/:id` | Delete any group |

### Real-Time (WebSocket)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/ws` | WebSocket connection for real-time order and group updates |

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Server health check |

## Gamification System

- **Energy:** Max 500. Spent to complete cafe orders. Recovered via study sessions.
- **XP & Levels:** `level * 100` XP is required to reach the next level. Rewarded from quizzes and orders.
- **Ranks:** Every 3 levels unlocks a new rank badge:
  - Coffee Novice (levels 1-3)
  - Focus Apprentice (levels 4-6)
  - Concentration Expert (levels 7-9)
  - Flow Master (levels 10-12)
  - Zen Grandmaster (levels 13+)
- **Global Leaderboard:** Top 5 users by XP (admins excluded). Each user can view their own global rank and profile.

## Authentication

- JWT tokens issued by Supabase Auth
- Backend validates tokens via JWKS endpoint (no local secret)
- Role-based access control: `user` vs `admin`
- Google OAuth via Supabase Auth redirect

## Database

Supabase local is used for development:
- `make db-up` starts PostgreSQL + Auth + Studio
- Migrations live in `supabase/migrations/`
- Seed data is applied automatically
- Detailed setup: see `LOCAL_SETUP_GUIDE.md`

## Testing

- **Go:** Table-driven unit tests for handlers, services, repositories, and integration (`go test ./...`)
- **Frontend:** Production build verified via TypeScript compiler (`npm run build`)
- **E2E:** Playwright end-to-end tests covering authentication, study sessions, collaborative orders, group management, admin dashboard, leaderboard, and WebSocket real-time updates
