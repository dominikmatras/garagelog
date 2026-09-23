package user

import (
	"errors"
	"net/http"

	"github.com/dominikmatras/garagelog/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB *pgxpool.Pool
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to hash password",
		})
		return
	}

	var newUser User

	query := `INSERT INTO users (first_name, last_name, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id, first_name, last_name, email`

	err = h.DB.QueryRow(c.Request.Context(), query, req.FirstName, req.LastName, req.Email, string(passwordHash)).Scan(&newUser.ID, &newUser.FirstName, &newUser.LastName, &newUser.Email)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, newUser)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	var user User

	query := `SELECT id, first_name, last_name, email, password_hash, role FROM users WHERE email = $1`

	err := h.DB.QueryRow(c.Request.Context(), query, req.Email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid email or password",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to login",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid email or password",
		})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	response := LoginResponse{
		Token: token,
		User:  user,
	}

	c.JSON(http.StatusOK, response)

}

func (h *Handler) GetUsers(c *gin.Context) {
	query := `SELECT id, first_name, last_name, email, role FROM users ORDER BY created_at`

	rows, err := h.DB.Query(c.Request.Context(), query)
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get users",
		})
		return
	}

	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var user User

		err := rows.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to read user data",
			})
			return
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read user data",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated, user id not found",
		})
		return
	}

	var user User 

	query := `SELECT id, first_name, last_name, email, role FROM users WHERE id = $1`

	err := h.DB.QueryRow(c.Request.Context(), query, userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get user",
		})
		return
	}
	c.JSON(http.StatusOK, user)
}