package handlers

import (
	"github.com/isw2-unileon/FocusCafe-project/backend/internal/auth"
	service "github.com/isw2-unileon/FocusCafe-project/backend/internal/services"
)

// Handler defines the required dependencies for managing auth petitions
type Handler struct {
	SupabaseURL            string
	SupabaseKey            string
	SupabaseServiceRoleKey string
	ClientURL              string
	Auth                   auth.TokenValidator

	UserService       service.UserServiceInterface
	UserOrdersService service.UserOrdersServiceInterface
	StudyService      service.StudyServiceInterface
	AIService         service.AIServiceInterface
}

// NewHandler creates a new instance of Handler with the provided dependencies
func NewHandler(url string, key string, serviceRoleKey string, clientURL string, auth auth.TokenValidator, userService *service.UserService, userOrdersService *service.UserOrdersService, studyService *service.StudyService, aiService service.AIServiceInterface) *Handler {
	return &Handler{
		SupabaseURL:            url,
		SupabaseKey:            key,
		SupabaseServiceRoleKey: serviceRoleKey,
		ClientURL:              clientURL,
		Auth:                   auth,
		UserService:            userService,
		UserOrdersService:      userOrdersService,
		StudyService:           studyService,
		AIService:              aiService,
	}
}
