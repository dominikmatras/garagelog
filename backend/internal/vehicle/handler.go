package vehicle

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

// GET ALL

func (h *Handler) GetVehicles(c *gin.Context) {
	query := `SELECT id, brand, model, year, mileage FROM vehicles ORDER BY id`

	rows, err := h.DB.Query(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get vehicles",
		})
		return
	}

	defer rows.Close()

	vehicles := make([]Vehicle, 0)

	for rows.Next() {
		var vehicle Vehicle

		err := rows.Scan(&vehicle.ID, &vehicle.Brand, &vehicle.Model, &vehicle.Year, &vehicle.Mileage)

		if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read vehicle data",
			})
			return
		}

		vehicles = append(vehicles, vehicle)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read vehicle data",
		})
		return
	}

	c.JSON(http.StatusOK, vehicles)
}

// GET BY ID

func (h *Handler) GetVehicleByID (c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid vehicle id",
		})
		return
	}

	var vehicle Vehicle

	query := `SELECT id, brand, model, year, mileage FROM vehicles WHERE id = $1`

	err = h.DB.QueryRow(c.Request.Context(), query, id).Scan(&vehicle.ID, &vehicle.Brand, &vehicle.Model, &vehicle.Year, &vehicle.Mileage)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
			"error": "vehicle not found",
			})
		return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get vehicle",
		})
		return
	}

	c.JSON(http.StatusOK, vehicle)
}

// ADD

func (h *Handler) AddVehicle(c *gin.Context) {
	var newVehicle Vehicle

	if err := c.ShouldBindJSON(&newVehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	query := `INSERT INTO vehicles (brand, model, year, mileage) VALUES ($1, $2, $3, $4) RETURNING id`

	err := h.DB.QueryRow(c, query, newVehicle.Brand, newVehicle.Model, newVehicle.Year, newVehicle.Mileage).Scan(&newVehicle.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to insert vehicle into database",
		})
		return
	}

	c.JSON(http.StatusCreated, newVehicle)
}

// UPDATE

func (h *Handler) UpdateVehicle(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H {
			"error": "invalid vehicle id",
		})
		return
	}

	var updatedVehicle Vehicle

	if err := c.ShouldBindJSON(&updatedVehicle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	query := `UPDATE vehicles SET brand = $1, model = $2, year = $3, mileage = $4, updated_at = NOW() WHERE id = $5 RETURNING id, brand, model, year, mileage`

	err = h.DB.QueryRow(c.Request.Context(), query, updatedVehicle.Brand, updatedVehicle.Model, updatedVehicle.Year, updatedVehicle.Mileage, id).Scan(&updatedVehicle.ID, &updatedVehicle.Brand, &updatedVehicle.Model, &updatedVehicle.Year, &updatedVehicle.Mileage)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "vehicle not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update vehicle",
		})
		return
	}

	c.JSON(http.StatusOK, updatedVehicle)
}

// DELETE

func (h *Handler) DeleteVehicle(c *gin.Context) {
	idParam := c.Param("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid vehicle id",
		})
		return
	}

	commandTag, err := h.DB.Exec(c.Request.Context(), `DELETE FROM vehicles WHERE id = $1`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete vehicle",
		})
		return
	}

	if commandTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "vehicle not found",
		})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}