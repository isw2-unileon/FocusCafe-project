# Architectural and Design Documentation

This document describes the architectural patterns, system design, and software engineering principles applied in FocusCafe. The system is built following a decoupled Client-Server architecture to maximize scalability, testability, and resilience.

---

## 1. High-Level System Architecture

FocusCafe is structured around a **Client-Server model**, dividing the responsibilities into an interactive frontend client, a robust monolithic backend API, and managed external microservices.

```test
+-------------------------------------------------------------------+
|                           CLIENT TIER                             |
|    Frontend Application (React / Next.js / Vue / Angular)         |
+-------------------------------------------------------------------+
|                        HTTP REST / WebSockets                     |
|                                 v                                 | 
+-------------------------------------------------------------------+
|                           SERVER TIER                             |
|          Go (Golang) REST API - Modular Monolith                  |
+-------------------------------------------------------------------+
|                              |                  |                 |
| GORM (SQL)                   | HTTP REST        | HTTP REST       |
|     v                              v                  v           |
+--------------------+          +----------------+    +-------------+
|   DATA STORAGE     |          | AUTH & PROFILE |    |  AI ENGINE  |
| PostgreSQL/SQLite  |          |  Supabase Auth |    |Google Gemini|
+--------------------+          +----------------+    +-------------+
```

### Architectural Components

1. **Client Tier (Frontend):** Manages the User Interface (UI) and local state, capturing user actions (e.g., uploading a PDF, completing an order or loging in) and communicating asynchronously with the backend.
2. **Server Tier (Backend REST API):** Built natively in Go. It acts as the core controller of business rules, securely validating requests, coordinating persistence, and managing concurrent internal flows.
3. **External Services & Adapters:**
   * **Supabase & Google OAuth2:** Delegated cloud infrastructure utilized for secure identity management. It acts as an OAuth2 broker, enabling users to authenticate seamlessly using their **Google Accounts**. Upon successful third-party login, Supabase issues a cryptographically signed JSON Web Token (JWT) containing the user's secure metadata, which our Go backend subsequently validates.
   * **Google Gemini API:** External generative AI service used asynchronously to ingest parsed document strings and emit formatted educational assets (JSON quizzes).
   * **SQL Database:** Relational database managed through an ORM (GORM) supporting robust transactional consistency.

---

## 2. Backend Internal Architecture (Separation of Concerns)

To avoid tight coupling and fulfill the assignment's requirement for *Separation of Concerns* and *Clean Code*, the Go backend is structured into **three distinct layers**:

```test
[ HTTP Request ] ---> ( 1. HANDLER LAYER )  <-- Handlers, Routers & GIN Context
|
v
( 2. SERVICE LAYER )  <-- Business Logic & AI Orchestration
|
v
( 3. REPOSITORY LAYER ) <-- GORM Database Access (SQL)
|
v
[ Database / Disk ]
```

### 1. Handler Layer (`/internal/handlers`)

* **Responsibility:** Handles network protocols, routing via the Gin Gonic framework, HTTP request binding, input validation, and HTTP status code responses (e.g., `200 OK`, `400 Bad Request`, `404 Not Found`).
* **Isolation:** It does not know *how* data is stored or *how* an AI quiz is mathematically processed; it only knows how to interpret a network request and return a JSON payload.

### 2. Service Layer (`/internal/services`)

* **Responsibility:** Encapsulates the core business logic of FocusCafe. This is the brain of the application.
* **Key Tasks:** Orchestrates energy point attribution algorithms, parses incoming multipart files (PDF text extraction), and configures the detailed prompt contracts sent to the Google Gemini AI interface.

### 3. Repository Layer (`/internal/repository`)

* **Responsibility:** Direct communication with the SQL database infrastructure using GORM.
* **Isolation:** It contains explicit SQL queries or ORM functions (e.g., `db.Create()`, `db.Where()`). No business rules or HTTP abstractions cross into this layer.

---

## 3. Design Decisions & Justifications

### 1. Choice of Go (Golang) for the Backend

* **Justification:** Go provides exceptional performance, high concurrency safety through native goroutines, and a very low memory footprint compared to runtimes like Node.js or Java. It compiling to a single binary greatly optimizes containerization (`Dockerfile`) and execution speed within the CI/CD deployment pipelines.

### 2. Relational SQL Database Layer via GORM

* **Justification:** Complex sub-systems like study groups, invite codes, and users require strict referential integrity constraints, relational cross-joins, and transaction handling. Using a structured relational SQL schema prevents data corruption in shared group counters or gamification progress metrics.

### 3. Comprehensive Mocking Architecture for Resilient Testing

* **Justification:** To achieve a robust, isolated testing environment without triggering real network costs or database pollution during CI/CD runs:
  * Database persistence is completely mocked or substituted with local fast in-memory instances.
  * Third-party networks (Supabase Auth and Gemini API) are isolated via HTTP request interceptors and test recorders (`httptest`), ensuring that tests check the application logic deterministically without real external dependencies.
  
## 4. Database Schema

The application utilizes a relational SQL database schema managed via GORM. Since identity management is delegated, our schema connects directly with the internal **Supabase Auth Engine**.

Below is the Entity-Relationship (ER) representation highlighting how the internal authentication records link directly to our custom application tables (`users`, `groups`, `study_sessions`)

### Table `users`

| Name | Type | Constraints |
|------|------|-------------|
| `first_name` | `text` |  |
| `last_name` | `text` |  |
| `username` | `text` |  |
| `email` | `text` |  Unique |
| `role` | `text` |  Nullable |
| `id` | `uuid` | Primary |
| `created_at` | `timestamptz` |  Nullable |
| `group_id` | `int8` |  Nullable |

### Table `user_progress`

| Name | Type | Constraints |
|------|------|-------------|
| `user_id` | `uuid` | Primary |
| `energy` | `int4` |  Nullable |
| `level` | `int4` |  Nullable |
| `xp` | `int4` |  Nullable |

### Table `study_materials`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `user_id` | `uuid` |  |
| `title` | `text` |  |
| `subject_name` | `text` |  |
| `file_path` | `text` |  |
| `upload_date` | `timestamp` |  Nullable |
| `content` | `text` |  Nullable |

### Table `study_sessions`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `user_id` | `uuid` |  |
| `material_id` | `int8` |  |
| `duration_minutes` | `int8` |  |
| `start_time` | `timestamp` |  |
| `end_time` | `timestamp` |  Nullable |
| `status` | `text` |  |

### Table `cafe_orders`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `name` | `text` |  |
| `category` | `text` |  Nullable |
| `energy_cost` | `int8` |  |
| `reward_xp` | `int8` |  |
| `description` | `text` |  Nullable |
| `required_level` | `int8` |  Nullable |

### Table `quizzes`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `session_id` | `int8` |  |
| `generated_at` | `timestamp` |  Nullable |

### Table `questions`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `quiz_id` | `int8` |  |
| `question_text` | `text` |  |
| `option_a` | `text` |  |
| `option_b` | `text` |  |
| `option_c` | `text` |  |
| `option_d` | `text` |  |
| `correct_answer` | `bpchar` |  |
| `explanation` | `text` |  Nullable |

### Table `user_orders`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `cafe_order_id` | `int8` |  |
| `status` | `text` |  Nullable |
| `created_at` | `timestamp` |  Nullable |
| `user_id` | `uuid` |  Nullable |
| `group_id` | `int8` |  Nullable |

### Table `groups`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `created_at` | `timestamptz` |  |
| `name` | `text` |  |
| `invite_code` | `text` |  Unique |
| `leader_id` | `uuid` |  |
