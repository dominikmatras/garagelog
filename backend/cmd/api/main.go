package main

import (
	"net/http"

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

	vehicleHandler := &vehicle.Handler{
		DB: db,
	}

	userHandler := &user.Handler{
		DB: db,
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app": "GarageLog",
		})
	})

	router.POST("/auth/register", userHandler.Register)

	router.GET("/vehicles", vehicleHandler.GetVehicles)
	router.GET("/vehicles/:id", vehicleHandler.GetVehicleByID)
	router.POST("/vehicles", vehicleHandler.AddVehicle)
	router.PUT("/vehicles/:id", vehicleHandler.UpdateVehicle)
	router.DELETE("/vehicles/:id", vehicleHandler.DeleteVehicle)

	router.Run(":8080")
}