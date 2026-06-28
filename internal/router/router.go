package router

import (
	"github.com/gin-gonic/gin"

	"pharmacy-backend/internal/config"
	"pharmacy-backend/internal/handler"
	"pharmacy-backend/internal/middleware"
	"pharmacy-backend/pkg/jwt"
)

// Handlers bundles the HTTP handlers the router needs.
type Handlers struct {
	Auth     *handler.AuthHandler
	User     *handler.UserHandler
	Category  *handler.CategoryHandler
	Medicine  *handler.MedicineHandler
	Stock     *handler.StockHandler
	Dashboard *handler.DashboardHandler
}

// New builds the Gin engine with all middleware and routes registered.
func New(cfg *config.Config, jwtMgr *jwt.Manager, h Handlers) *gin.Engine {
	if cfg.IsProd {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS(cfg.CORSAllowedOrigins))

	// Health check (public).
	r.GET("/api/v1/health", handler.Health)

	v1 := r.Group("/api/v1")

	// --- Auth ---
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
		auth.POST("/logout", middleware.Auth(jwtMgr), h.Auth.Logout)
		auth.GET("/me", middleware.Auth(jwtMgr), h.Auth.Me)
	}

	// --- Users (Admin only) ---
	users := v1.Group("/users", middleware.Auth(jwtMgr), middleware.AdminOnly())
	{
		users.GET("", h.User.List)
		users.POST("", h.User.Create)
		users.GET("/:id", h.User.Get)
		users.PUT("/:id", h.User.Update)
		users.PATCH("/:id/status", h.User.SetStatus)
		users.PATCH("/:id/password", h.User.ResetPassword)
	}

	// --- Categories (read: any authenticated; write: Admin) ---
	categories := v1.Group("/categories", middleware.Auth(jwtMgr))
	{
		categories.GET("", h.Category.List)
		categories.GET("/:id", h.Category.Get)
		categories.POST("", middleware.AdminOnly(), h.Category.Create)
		categories.PUT("/:id", middleware.AdminOnly(), h.Category.Update)
		categories.DELETE("/:id", middleware.AdminOnly(), h.Category.Delete)
	}

	// --- Medicines (read: any authenticated; write: Admin) ---
	medicines := v1.Group("/medicines", middleware.Auth(jwtMgr))
	{
		medicines.GET("", h.Medicine.List)
		medicines.GET("/:id", h.Medicine.Get)
		medicines.GET("/:id/lots", h.Stock.LotsByMedicine)
		medicines.POST("", middleware.AdminOnly(), h.Medicine.Create)
		medicines.PUT("/:id", middleware.AdminOnly(), h.Medicine.Update)
		medicines.DELETE("/:id", middleware.AdminOnly(), h.Medicine.Delete)
	}

	// --- Stock movements (Admin + Staff) ---
	stock := v1.Group("/stock", middleware.Auth(jwtMgr))
	{
		stock.POST("/in", h.Stock.StockIn)
		stock.POST("/out", h.Stock.StockOut)
		stock.POST("/return", h.Stock.Return)
		stock.GET("/transactions", h.Stock.Transactions)
	}

	// --- Dashboard & alerts (Admin + Staff) ---
	dash := v1.Group("/dashboard", middleware.Auth(jwtMgr))
	{
		dash.GET("/summary", h.Dashboard.Summary)
		dash.GET("/near-expiry", h.Dashboard.NearExpiry)
		dash.GET("/low-stock", h.Dashboard.LowStock)
	}

	return r
}
