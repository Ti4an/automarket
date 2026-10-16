package main

import (
	"errors"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"automarket/handlers"
	"automarket/middleware"
	"automarket/models"
	"automarket/services"
	"automarket/storage"
)

func main() {
	store, err := storage.New("data.json")
	if err != nil {
		log.Fatalf("no se pudo cargar el almacenamiento: %v", err)
	}
	audit, err := middleware.NewAuditLogger("audit.log")
	if err != nil {
		log.Fatalf("no se pudo abrir el registro de auditoría: %v", err)
	}

	// Carpeta donde se guardan las imágenes subidas
	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatalf("no se pudo crear la carpeta uploads: %v", err)
	}

	authService := services.NewAuthService(store)
	vehicleService := services.NewVehicleService(store, "uploads")
	h := handlers.New(authService, vehicleService)

	if err := ensureAdmin(authService); err != nil {
		log.Fatalf("no se pudo crear el administrador: %v", err)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}
	router.Use(gin.Recovery(), audit.Middleware())

	// Interfaz web
	router.StaticFile("/", "./web/index.html")
	router.StaticFile("/panel", "./web/panel.html")
	router.StaticFile("/styles.css", "./web/styles.css")
	router.StaticFile("/app.js", "./web/app.js")
	router.StaticFile("/panel.js", "./web/panel.js")
	router.Static("/uploads", "./uploads")

	// Visitante (público)
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
	router.GET("/vehicles", h.Catalog)
	router.GET("/vehicles/:id", h.GetVehicle)

	authenticated := router.Group("/", middleware.Authenticate(authService))

	// Vendedor
	seller := authenticated.Group("/", middleware.RequireRole(models.RoleSeller))
	seller.POST("/vehicles", h.Publish)
	seller.PATCH("/vehicles/:id/sold", h.MarkSold)
	seller.GET("/my/vehicles", h.MyVehicles)

	// Administrador
	admin := authenticated.Group("/", middleware.RequireRole(models.RoleAdmin))
	admin.PATCH("/vehicles/:id/approve", h.Approve)
	admin.DELETE("/vehicles/:id", h.DeleteVehicle)
	admin.DELETE("/users/:id", h.DeleteUser)
	admin.GET("/admin/vehicles", h.AllVehicles)
	admin.GET("/admin/users", h.Sellers)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("el servidor se detuvo: %v", err)
	}
}

func ensureAdmin(auth *services.AuthService) error {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		return errors.New("defina ADMIN_EMAIL y ADMIN_PASSWORD")
	}
	_, err := auth.CreateUser("Administrador", email, "N/A", password, models.RoleAdmin)
	if errors.Is(err, services.ErrEmailTaken) {
		return nil
	}
	return err
}