package handler

import (
	"log/slog"
	"marketplace/internal/service"
	"marketplace/pkg/auth"
	"net/http"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	service      *service.Service
	TokenManager *auth.TokenManager
	log          *slog.Logger
}

func NewHandler(services *service.Service, tm *auth.TokenManager, log *slog.Logger) *Handler {
	return &Handler{
		service:      services,
		TokenManager: tm,
		log:          log,
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	apiV1 := router.Group("/api/v1")
	{
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/register", h.signUp)
			authGroup.POST("/login", h.signIn)
		}

		adsGroup := apiV1.Group("/ads")
		{
			adsGroup.GET("", h.GetAllAds)
			adsGroup.GET("/:id", h.GetAdByID)

			adsSecure := adsGroup.Group("")
			adsSecure.Use(h.AuthMiddleware())
			{
				adsSecure.POST("", h.CreateAd)
				adsSecure.PATCH("/:id", h.UpdateAd)
				adsSecure.DELETE("/:id", h.DeleteAd)
			}
		}
	}

	return router
}
