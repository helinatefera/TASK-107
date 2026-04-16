package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/jmoiron/sqlx"

	"github.com/chargeops/api/internal/apperror"
	"github.com/chargeops/api/internal/config"
	"github.com/chargeops/api/internal/db"
	"github.com/chargeops/api/internal/dto"
	"github.com/chargeops/api/internal/handler"
	mw "github.com/chargeops/api/internal/middleware"
	"github.com/chargeops/api/internal/service"
	"github.com/chargeops/api/internal/worker"
)

type customValidator struct {
	validator *validator.Validate
}

func (cv *customValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			msgs := make([]string, 0, len(ve))
			for _, fe := range ve {
				msgs = append(msgs, fe.Field()+" failed on "+fe.Tag())
			}
			return apperror.New(http.StatusBadRequest, "validation failed: "+joinStrings(msgs, "; "))
		}
		return apperror.New(http.StatusBadRequest, "invalid request")
	}
	return nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// seedAdmin creates an administrator user from env vars if configured and no
// admin exists yet. This removes the need for manual DB commands during setup.
func seedAdmin(cfg *config.Config, database *sqlx.DB) {
	if cfg.SeedAdminEmail == "" || cfg.SeedAdminPass == "" {
		return
	}
	ctx := context.Background()
	// Check if this email already exists (idempotent)
	user, err := service.Register(ctx, database, &dto.RegisterRequest{
		Email:       cfg.SeedAdminEmail,
		Password:    cfg.SeedAdminPass,
		DisplayName: "Admin",
	}, cfg)
	if err != nil {
		// Already exists or invalid — skip silently
		log.Info().Str("email", cfg.SeedAdminEmail).Msg("seed admin: user already exists or registration skipped")
		return
	}
	if err := service.UpdateUserRole(ctx, database, user.ID, "administrator"); err != nil {
		log.Error().Err(err).Msg("seed admin: failed to promote user")
		return
	}
	log.Info().Str("email", cfg.SeedAdminEmail).Msg("seed admin: administrator user created")
}

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer database.Close()

	if err := db.RunMigrations(database, cfg.MigrationsDir); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations applied successfully")

	seedAdmin(cfg, database)

	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = apperror.HTTPErrorHandler
	e.Validator = &customValidator{validator: validator.New()}

	// Global middleware stack
	e.Use(mw.RequestID())
	e.Use(mw.Logger())
	e.Use(mw.MetricsRecorder(database))
	e.Use(echomw.Recover())
	e.Use(mw.Auth(database))

	// Health endpoint (public, skipped by auth)
	e.GET("/health", handler.HealthHandler(database, cfg))

	// API v1 routes
	v1 := e.Group("/api/v1")

	// Auth routes (public, skipped by auth middleware)
	auth := v1.Group("/auth")
	auth.POST("/register", handler.RegisterHandler(database, cfg))
	auth.POST("/login", handler.LoginHandler(database, cfg), mw.RateLimit(database, cfg))
	auth.POST("/logout", handler.LogoutHandler(database, cfg))
	auth.POST("/refresh", handler.RefreshHandler(database, cfg))
	auth.POST("/recover", handler.RecoverHandler(database, cfg))
	auth.POST("/recover/reset", handler.ResetPasswordHandler(database, cfg))

	// User routes (admin + self)
	users := v1.Group("/users", mw.RequireRole("administrator", "user", "merchant"))
	users.GET("", handler.ListUsersHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("user.read"))
	users.GET("/me", handler.GetCurrentUserHandler(database, cfg), mw.RequirePermission("user.read"))
	users.GET("/:id", handler.GetUserHandler(database, cfg), mw.RequirePermission("user.read"))
	users.PUT("/:id", handler.UpdateUserHandler(database, cfg), mw.RequirePermission("user.read"))
	users.PUT("/:id/role", handler.UpdateRoleHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("user.manage"), mw.Audit(database))
	users.PUT("/:id/org", handler.UpdateUserOrgHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("user.manage"), mw.Audit(database))
	users.DELETE("/:id", handler.DeleteUserHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("user.manage"), mw.Audit(database))
	users.GET("/:id/permissions", handler.GetPermissionsHandler(database, cfg), mw.RequirePermission("user.read"))
	users.PUT("/:id/permissions", handler.UpdatePermissionsHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("user.manage"), mw.Audit(database))

	// Organization routes
	orgs := v1.Group("/orgs", mw.RequireRole("administrator", "merchant"))
	orgs.POST("", handler.CreateOrgHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("org.manage"), mw.Audit(database))
	orgs.GET("", handler.ListOrgsHandler(database, cfg), mw.RequirePermission("org.read"))
	orgs.GET("/:id", handler.GetOrgHandler(database, cfg), mw.RequirePermission("org.read"))
	orgs.PUT("/:id", handler.UpdateOrgHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("org.manage"), mw.Audit(database))
	orgs.DELETE("/:id", handler.DeleteOrgHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("org.manage"), mw.Audit(database))

	// Warehouse routes
	warehouses := v1.Group("/warehouses", mw.RequireRole("administrator", "merchant"))
	warehouses.POST("", handler.CreateWarehouseHandler(database, cfg), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	warehouses.GET("", handler.ListWarehousesHandler(database, cfg), mw.RequirePermission("warehouse.read"))
	warehouses.GET("/:id", handler.GetWarehouseHandler(database, cfg), mw.RequirePermission("warehouse.read"))
	warehouses.PUT("/:id", handler.UpdateWarehouseHandler(database, cfg), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	warehouses.DELETE("/:id", handler.DeleteWarehouseHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	warehouses.POST("/:id/zones", handler.CreateZoneHandler(database, cfg), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	warehouses.GET("/:id/zones", handler.ListZonesHandler(database, cfg), mw.RequirePermission("warehouse.read"))

	zones := v1.Group("/zones", mw.RequireRole("administrator", "merchant"))
	zones.POST("/:id/bins", handler.CreateBinHandler(database, cfg), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	zones.GET("/:id/bins", handler.ListBinsHandler(database, cfg), mw.RequirePermission("warehouse.read"))

	bins := v1.Group("/bins", mw.RequireRole("administrator", "merchant"))
	bins.PUT("/:id", handler.UpdateBinHandler(database, cfg), mw.RequirePermission("warehouse.manage"), mw.Audit(database))
	bins.DELETE("/:id", handler.DeleteBinHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("warehouse.manage"), mw.Audit(database))

	// Category routes
	categories := v1.Group("/categories")
	categories.POST("", handler.CreateCategoryHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("item.manage"), mw.Audit(database))
	categories.GET("", handler.ListCategoriesHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("item.read"))

	// Item routes
	items := v1.Group("/items")
	items.POST("", handler.CreateItemHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("item.manage"), mw.Audit(database))
	items.GET("", handler.ListItemsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("item.read"))
	items.GET("/:id", handler.GetItemHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("item.read"))
	items.PUT("/:id", handler.UpdateItemHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("item.manage"), mw.Audit(database))
	items.DELETE("/:id", handler.DeleteItemHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("item.manage"), mw.Audit(database))

	// Unit routes
	units := v1.Group("/units")
	units.POST("", handler.CreateUnitHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("item.manage"), mw.Audit(database))
	units.GET("", handler.ListUnitsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("item.read"))
	units.POST("/conversions", handler.CreateConversionHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("item.manage"), mw.Audit(database))
	units.GET("/conversions", handler.ListConversionsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("item.read"))

	// Supplier routes
	suppliers := v1.Group("/suppliers", mw.RequireRole("administrator", "merchant"))
	suppliers.POST("", handler.CreateSupplierHandler(database, cfg), mw.RequirePermission("supplier.manage"), mw.Audit(database))
	suppliers.GET("", handler.ListSuppliersHandler(database, cfg), mw.RequirePermission("supplier.read"))
	suppliers.GET("/:id", handler.GetSupplierHandler(database, cfg), mw.RequirePermission("supplier.read"))
	suppliers.PUT("/:id", handler.UpdateSupplierHandler(database, cfg), mw.RequirePermission("supplier.manage"), mw.Audit(database))

	// Carrier routes
	carriers := v1.Group("/carriers", mw.RequireRole("administrator", "merchant"))
	carriers.POST("", handler.CreateCarrierHandler(database, cfg), mw.RequirePermission("supplier.manage"), mw.Audit(database))
	carriers.GET("", handler.ListCarriersHandler(database, cfg), mw.RequirePermission("supplier.read"))
	carriers.GET("/:id", handler.GetCarrierHandler(database, cfg), mw.RequirePermission("supplier.read"))
	carriers.PUT("/:id", handler.UpdateCarrierHandler(database, cfg), mw.RequirePermission("supplier.manage"), mw.Audit(database))

	// Station routes
	stations := v1.Group("/stations", mw.RequireRole("administrator", "merchant"))
	stations.POST("", handler.CreateStationHandler(database, cfg), mw.RequirePermission("station.manage"), mw.Audit(database))
	stations.GET("", handler.ListStationsHandler(database, cfg), mw.RequirePermission("station.read"))
	stations.GET("/:id", handler.GetStationHandler(database, cfg), mw.RequirePermission("station.read"))
	stations.PUT("/:id", handler.UpdateStationHandler(database, cfg), mw.RequirePermission("station.manage"), mw.Audit(database))
	stations.DELETE("/:id", handler.DeleteStationHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("station.manage"), mw.Audit(database))
	stations.POST("/:id/devices", handler.CreateDeviceHandler(database, cfg), mw.RequirePermission("station.manage"), mw.Audit(database))
	stations.GET("/:id/devices", handler.ListDevicesHandler(database, cfg), mw.RequirePermission("station.read"))

	devices := v1.Group("/devices", mw.RequireRole("administrator", "merchant"))
	devices.PUT("/:id", handler.UpdateDeviceHandler(database, cfg), mw.RequirePermission("station.manage"), mw.Audit(database))
	devices.DELETE("/:id", handler.DeleteDeviceHandler(database, cfg), mw.RequirePermission("station.manage"), mw.Audit(database))

	// Pricing routes
	pricing := v1.Group("/pricing", mw.RequireRole("administrator", "merchant"))
	pricing.POST("/templates", handler.CreateTemplateHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.GET("/templates", handler.ListTemplatesHandler(database, cfg), mw.RequirePermission("pricing.read"))
	pricing.GET("/templates/:id", handler.GetTemplateHandler(database, cfg), mw.RequirePermission("pricing.read"))
	pricing.POST("/templates/:id/versions", handler.CreateVersionHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.GET("/templates/:id/versions", handler.ListVersionsHandler(database, cfg), mw.RequirePermission("pricing.read"))
	pricing.GET("/versions/:id", handler.GetVersionHandler(database, cfg), mw.RequirePermission("pricing.read"))
	pricing.POST("/versions/:id/activate", handler.ActivateVersionHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.POST("/versions/:id/deactivate", handler.DeactivateVersionHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.POST("/versions/:id/rollback", handler.RollbackVersionHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.POST("/versions/:id/tou-rules", handler.CreateTOURuleHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))
	pricing.GET("/versions/:id/tou-rules", handler.ListTOURulesHandler(database, cfg), mw.RequirePermission("pricing.read"))
	pricing.DELETE("/tou-rules/:id", handler.DeleteTOURuleHandler(database, cfg), mw.RequirePermission("pricing.manage"), mw.Audit(database))

	// Order routes
	orders := v1.Group("/orders", mw.RequireRole("administrator", "merchant", "user"))
	orders.POST("", handler.CreateOrderHandler(database, cfg), mw.RequirePermission("order.create"))
	orders.GET("", handler.ListOrdersHandler(database, cfg), mw.RequirePermission("order.read"))
	orders.GET("/:id", handler.GetOrderHandler(database, cfg), mw.RequirePermission("order.read"))
	orders.POST("/:id/recalculate", handler.RecalculateOrderHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("order.create"))

	// Content routes
	content := v1.Group("/content")
	content.POST("/carousel", handler.CreateCarouselHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.GET("/carousel", handler.ListCarouselsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user", "guest"), mw.RequirePermission("content.read"))
	content.PUT("/carousel/:id", handler.UpdateCarouselHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.DELETE("/carousel/:id", handler.DeleteCarouselHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.POST("/campaigns", handler.CreateCampaignHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.GET("/campaigns", handler.ListCampaignsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user", "guest"), mw.RequirePermission("content.read"))
	content.PUT("/campaigns/:id", handler.UpdateCampaignHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.DELETE("/campaigns/:id", handler.DeleteCampaignHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.POST("/rankings", handler.CreateRankingHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.GET("/rankings", handler.ListRankingsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user", "guest"), mw.RequirePermission("content.read"))
	content.PUT("/rankings/:id", handler.UpdateRankingHandler(database, cfg), mw.RequireRole("administrator", "merchant"), mw.RequirePermission("content.manage"), mw.Audit(database))
	content.DELETE("/rankings/:id", handler.DeleteRankingHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("content.manage"), mw.Audit(database))

	// Notification routes
	notif := v1.Group("/notifications")
	notif.GET("/inbox", handler.ListInboxHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("notification.read"))
	notif.POST("/inbox/:id/read", handler.MarkReadHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("notification.read"))
	notif.POST("/inbox/:id/dismiss", handler.MarkDismissHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("notification.read"))
	notif.GET("/subscriptions", handler.ListSubscriptionsHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("notification.read"))
	notif.PUT("/subscriptions/:id", handler.UpdateSubscriptionHandler(database, cfg), mw.RequireRole("administrator", "merchant", "user"), mw.RequirePermission("notification.read"))
	notif.POST("/templates", handler.CreateNotificationTemplateHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("notification.manage"), mw.Audit(database))
	notif.GET("/templates", handler.ListNotificationTemplatesHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("notification.manage"))
	notif.POST("/send", handler.SendNotificationHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("notification.manage"), mw.Audit(database))
	notif.GET("/stats", handler.GetDeliveryStatsHandler(database, cfg), mw.RequireRole("administrator"), mw.RequirePermission("notification.manage"))

	// Admin routes
	admin := v1.Group("/admin", mw.RequireRole("administrator"))
	admin.GET("/audit-logs", handler.ListAuditLogsHandler(database, cfg), mw.RequirePermission("admin.audit"))
	admin.GET("/config", handler.GetConfigHandler(database, cfg), mw.RequirePermission("admin.config"))
	admin.PUT("/config/:key", handler.UpdateConfigHandler(database, cfg), mw.RequirePermission("admin.config"), mw.Audit(database))
	admin.GET("/metrics", handler.ListMetricsHandler(database, cfg), mw.RequirePermission("admin.metrics"))
	admin.GET("/metrics/export", handler.ExportMetricsHandler(database, cfg), mw.RequirePermission("admin.metrics"))

	// Start background workers
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	worker.StartAll(workerCtx, database)

	// Graceful shutdown
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server...")

	workerCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced shutdown")
	}
	log.Info().Msg("server stopped")
}
