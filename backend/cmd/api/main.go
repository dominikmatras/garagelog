package main

import (
	"net/http"

	"github.com/dominikmatras/garagelog/internal/auth"
	"github.com/dominikmatras/garagelog/internal/database"
	"github.com/dominikmatras/garagelog/internal/user"
	"github.com/dominikmatras/garagelog/internal/vehicle"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		panic(err)
	}
	
	defer db.Close()

	router := gin.Default()

	// Handler 

	userHandler := &user.Handler{
		DB: db,
	}

	vehicleHandler := &vehicle.Handler{
		DB: db,
	}

	// Public

	router.POST("/auth/register", userHandler.Register)
	router.POST("/auth/login", userHandler.Login)

	// User

	users := router.Group("/users")
	users.Use(auth.AuthMiddleware())

	users.GET("/me", userHandler.GetMe)

	// Admin

	admin := router.Group("/admin")
	admin.Use(auth.AuthMiddleware())
	admin.Use(auth.RequireAdmin())

	admin.GET("/users", userHandler.GetUsers)

	// Vehicles

	vehicles := router.Group("/vehicles")
	vehicles.Use(auth.AuthMiddleware())

	vehicles.GET("", vehicleHandler.GetVehicles)
	vehicles.GET("/:id", vehicleHandler.GetVehicleByID)
	vehicles.POST("", vehicleHandler.AddVehicle)
	vehicles.PUT("/:id", vehicleHandler.UpdateVehicle)
	vehicles.DELETE("/:id", vehicleHandler.DeleteVehicle)

	// Health Check

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app": "GarageLog",
		})
	})

	router.Run(":8080")
}