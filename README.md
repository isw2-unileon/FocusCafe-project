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
---

# Monorepo Template: Go + React/Vite

A monorepo template for full-stack applications with a **Go** backend and a **React + TypeScript + Vite** frontend.

## Project Structure

```text
├── backend/              Go API server (Gin)
│   ├── cmd/server/       Entry point
│   └── internal/config/  Environment config
│
├── frontend/             React + TypeScript + Vite + Tailwind
│   └── src/
│
├── e2e/                  Playwright E2E tests
├── .github/workflows/    CI/CD pipelines
└── Makefile              Dev commands
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 22+

## Getting Started

```bash
make install

# Terminal 1
make run-backend    # port 8080

# Terminal 2
make run-frontend   # port 5173
```

The Vite dev server proxies `/api` requests to the backend.

## Commands

| Command              | Description                     |
|----------------------|---------------------------------|
| `make install`       | Install all dependencies        |
| `make run-backend`   | Backend with hot reload (Air)   |
| `make run-frontend`  | Frontend dev server (Vite)      |
| `make test`          | Run all tests                   |
| `make lint`          | Run all linters                 |
| `make e2e`           | Run Playwright E2E tests        |

## API

| Method | Path         | Description    |
|--------|------------- |----------------|
| `GET`  | `/health`    | Health check   |
| `GET`  | `/api/hello` | Sample endpoint|
