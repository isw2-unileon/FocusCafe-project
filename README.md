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
- Real-time collaborative café rooms

Multiple users can join the same café "salon" and manage it together, contributing energy earned from their individual study sessions.

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
| **Backend** | Go 1.24+, Gin, GORM, JWT/JWKS (Supabase) |
| **Frontend** | React 18+, TypeScript, Vite, Tailwind CSS, shadcn/ui, Lucide React, react-hot-toast |
| **Auth** | Supabase Auth (email/password + Google OAuth) |
| **AI** | Gemini API (Gemini-flash 2.5) for quiz generation |
| **Database** | PostgreSQL via Supabase (local or remote) |
| **Testing** | Go table-driven tests, Playwright E2E (structure ready) |

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
│       ├── models/             # GORM models
│       ├── repository/         # Data access layer
│       ├── services/           # Business logic
│       └── supabase/           # Supabase client & JWT adapter
├── frontend/
│   └── src/
│       ├── components/         # Reusable UI components
│       ├── context/            # React contexts (Auth)
│       ├── lib/                # Utilities (JWT, notifications)
│       ├── pages/              # Route pages (Dashboard, Admin, EditProfile, ...)
│       ├── services/           # API client functions
│       └── types/              # TypeScript interfaces
├── supabase/
│   ├── config.toml             # Supabase local config
│   └── migrations/             # SQL migrations
├── Makefile
└── .env                        # Single env file (no .env.local)
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
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
| `SUPABASE_ANON_KEY` | Supabase anon/public key | Yes |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key (admin ops) | Yes |
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `OPENAI_API_KEY` | OpenAI API key (AI quizzes) | Yes |
| `CLIENT_URL` | Frontend origin (CORS) | Yes |
| `PORT` | Backend server port | Optional (default 8080) |

## Commands

| Command | Description |
|---------|-------------|
| `make install` | Install Go tools, download modules, install npm deps |
| `make run-backend` | Backend with hot reload (Air) |
| `make run-frontend` | Frontend dev server (Vite) |
| `make test` | Run Go + frontend unit tests |
| `make lint` | Run linters (golangci-lint + ESLint) |
| `make db-up` | Start Supabase local + apply migrations + seed |
| `make db-reset` | Reset database + re-apply migrations + seed |
| `make db-down` | Stop Supabase local |

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

### Admin (requires admin role)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/admin/users` | List all users |
| `GET` | `/api/admin/users?email=` | Search user by email |
| `POST` | `/api/admin/users` | Create new user |
| `DELETE` | `/api/admin/users/:id` | Delete user (DB + Supabase Auth) |

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

- **Go:** Table-driven unit tests for handlers, services, and repositories (`go test ./...`)
- **Frontend:** Unit test structure prepared (`npm run test`)
- **E2E:** Playwright tests (requires backend + frontend running)
