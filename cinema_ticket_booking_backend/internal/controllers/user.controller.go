package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/middleware"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TokenResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

var jwtSecret = []byte("my-secret-key")

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func generateJWT(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func LoginWithGoogle() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IDToken string `json:"idToken" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		ctx := context.Background()
		payload, err := middleware.GoogleAuth.VerifyIDToken(ctx, req.IDToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID token"})
			return
		}

		email := payload.Claims["email"].(string)
		name := ""
		picture := ""

		if nameVal, ok := payload.Claims["name"].(string); ok {
			name = nameVal
		}
		if pictureVal, ok := payload.Claims["picture"].(string); ok {
			picture = pictureVal
		}

		user, err := middleware.GoogleAuth.FindOrCreateUser(ctx, email, name, picture)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		token, err := generateJWT(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		log.Printf("[LOGIN] User logged in: %s (%s)", user.Username, user.Email)

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"data": TokenResponse{
				Token: token,
				User:  user,
			},
		})
	}
}

func GetCurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		u := user.(*models.User)
		c.JSON(http.StatusOK, gin.H{"data": u})
	}
}

func GetAllUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userCollection := database.OpenCollection("users")
		cursor, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}
		defer cursor.Close(ctx)

		var users []models.User
		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": users})
	}
}

func UpdateUserRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID := c.Param("id")
		objID, err := bson.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var req struct {
			Role models.UserRole `json:"role" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		userCollection := database.OpenCollection("users")
		_, err = userCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
			"$set": bson.M{
				"role":      req.Role,
				"updatedAt": time.Now(),
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
	}
}
