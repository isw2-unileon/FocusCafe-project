# Design Decisions and Technical Justifications

This document provides architectural records and justifications for the critical design patterns, library selections, and operational mechanisms implemented in FocusCafe. Each decision is evaluated against its contribution to maintainability, testability, and software resilience.

---

## 1. Monolithic Architecture with Layered Separation

* **Context:** We needed to choose between a microservices-based approach or a monolithic structure for the backend API.
* **Decision:** We adopted a **Modular Monolith** combined with strict **Layered Architecture (Separation of Concerns)**.
* **Justification:** Given the project's scope, a monolith significantly reduces network overhead and deployment complexity. By enforcing rigid boundaries between layers (`handlers` -> `services` -> `repositories`), we ensure that:
  * The HTTP infrastructure (Gin) remains entirely independent of our core business rules.
  * Modifying the underlying database schema (GORM) does not cause cascading code breaks into the routing logic.
* **Consequences:** Highly organized codebase where components are decoupled and easily interchangeable.

---

## 2. Choice of Frameworks and Core Libraries

### 1. HTTP Router: Gin Gonic (`github.com/gin-gonic/gin`)
* **Alternative Considered:** Native Go standard library (`net/http`) or Gorilla Mux.
* **Justification:** Gin provides a highly optimized, custom HTTP router built on a Radix tree structure, which yields exceptional routing performance and minimal memory allocations. Additionally, its built-in context management (`gin.Context`) simplifies JSON payload binding, structural validation, and structured error handling, allowing the team to deliver features faster without sacrificing speed.

### 2. Object-Relational Mapper: GORM (`gorm.io/gorm`)
* **Alternative Considered:** Raw SQL Queries via native `database/sql` or SQLX.
* **Justification:** GORM accelerates development by automating standard CRUD lifecycles, safely handling data migrations, and mitigating SQL injection vulnerabilities out of the box. For advanced relational features like study group joins or user profile links, it enables transparent relational preloading while retaining the flexibility to drop down into native SQL if deep optimization is required.

---

## 3. Asynchronous Task Handling and Concurrency Design

* **Context:** Ingesting PDF materials and requesting external generative AI evaluation (Google Gemini) are time-consuming network operations. Blocking the client's HTTP thread during these routines leads to a poor user experience and connection timeouts.
* **Decision:** We leverage **Go Goroutines** and non-blocking asynchronous orchestrations.
* **Justification:** Go’s native, lightweight concurrency model (goroutines) allows the backend to accept an incoming PDF, immediately store it, and trigger the AI processing pipeline in a background thread. The HTTP handler can then quickly return a `200 OK` or `202 Accepted` status status back to the frontend client, preventing thread starvation.
* **Consequences:** The system maintains high responsiveness. The frontend can query or poll for the quiz status asynchronously without freezing the user interface.

---

## 4. Testing Isolation and Dependency Mocking

* **Context:** The application communicates heavily with external cloud platforms (Supabase Auth, Google Auth and Google Gemini API). Relying on live network services during test execution introduces non-deterministic behaviors, structural costs, and failures during offline CI/CD pipeline runs.
* **Decision:** Total isolation through the Go standard library’s `net/http/httptest` recorder and mock interfaces.
* **Justification:** By creating mock implementations for our service boundaries, we ensure that our automated test suite (`go test -v -race ./...`) runs deterministically in milliseconds. 
  * For database verification, we substitute the production environment with mock records or in-memory targets.
  * For third-party APIs, we capture and replicate network responses using static text fixtures.
* **Consequences:** Flawless execution within the automated GitHub Actions environment, zero operational costs during compilation, and strict alignment with professional test-driven verification standards.