package integration

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/handlers"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/models"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/repository"
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestApp initializes the entire application stack for integration testing.
func setupTestApp() (*gin.Engine, *gorm.DB, *handlers.Handler) {
	gin.SetMode(gin.TestMode)

	// 1. Setup In-Memory SQLite database
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// 2. Migrate models
	db.AutoMigrate(&models.User{}, &models.CafeOrder{}, &models.UserOrder{}, &models.UserProgress{}, &models.Group{}, &models.StudySession{}, &models.StudyMaterial{})

	// 3. Initialize Repositories
	userRepo := repository.NewUserRepository(db)
	userOrderRepo := repository.NewUserOrdersRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	studyRepo := repository.NewStudyRepository(db)

	// 4. Initialize Services
	userService := services.NewUserService(userRepo)
	userOrderService := services.NewUserOrdersService(userOrderRepo)
	groupService := services.NewGroupService(groupRepo)
	studyService := services.NewStudyService(studyRepo)
	// AIService can be nil or mocked if needed for specific tests

	// 5. Initialize Handler
	h := &handlers.Handler{
		UserService:       userService,
		UserOrdersService: userOrderService,
		GroupService:      groupService,
		StudyService:      studyService,
	}

	// 6. Setup Router
	r := gin.New()
	r.Use(gin.Recovery())

	return r, db, h
}

// mockAuthMiddleware injects a test user into the Gin context.
func mockAuthMiddleware(userID uuid.UUID) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := &auth.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		c.Set("user", claims)
		c.Next()
	}
}

// setupSubtestContext creates a Gin context for testing a specific handler call.
func setupSubtestContext(w *httptest.ResponseRecorder, req *http.Request, userID uuid.UUID) (*gin.Context, *gin.Engine) {
	c, r := gin.CreateTestContext(w)
	c.Request = req
	
	if userID != uuid.Nil {
		claims := &auth.UserClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Subject: userID.String(),
			},
		}
		c.Set("user", claims)
	}
	
	return c, r
}
