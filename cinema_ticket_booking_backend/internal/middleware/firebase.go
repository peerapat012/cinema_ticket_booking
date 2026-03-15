package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"google.golang.org/api/idtoken"
)

var jwtSecret []byte

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func init() {
	godotenv.Load()
	jwtSecret = []byte(os.Getenv("JWT_SECRET"))
	if string(jwtSecret) == "" {
		jwtSecret = []byte("my-secret-key")
	}
}

type GoogleAuthMiddleware struct {
	client         *http.Client
	userCollection *mongo.Collection
}

func NewGoogleAuthMiddleware() *GoogleAuthMiddleware {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := &http.Client{}

	return &GoogleAuthMiddleware{
		client:         client,
		userCollection: database.OpenCollection("users"),
	}
}

func (m *GoogleAuthMiddleware) VerifyIDToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
	payload, err := idtoken.Validate(ctx, idToken, "")
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func (m *GoogleAuthMiddleware) FindOrCreateUser(ctx context.Context, email, name, picture string) (*models.User, error) {
	var user models.User
	err := m.userCollection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == nil {
		return &user, nil
	}

	if err == mongo.ErrNoDocuments {
		newUser := models.User{
			ID:        bson.ObjectID{},
			Username:  name,
			Email:     email,
			Avatar:    picture,
			Role:      models.UserRoleUser,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result, err := m.userCollection.InsertOne(ctx, newUser)
		if err != nil {
			return nil, err
		}

		newUser.ID = result.InsertedID.(bson.ObjectID)
		log.Printf("[LOGIN] New user created: %s (%s)", newUser.Username, newUser.Email)
		return &newUser, nil
	}

	return nil, err
}

func (m *GoogleAuthMiddleware) GetUserByID(ctx context.Context, userID bson.ObjectID) (*models.User, error) {
	var user models.User
	err := m.userCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

var GoogleAuth *GoogleAuthMiddleware

func init() {
	GoogleAuth = NewGoogleAuthMiddleware()
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if idToken == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
			c.Abort()
			return
		}

		ctx := context.Background()

		token, err := jwt.ParseWithClaims(idToken, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err == nil && token != nil {
			if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
				userIDStr := claims.UserID
				email := claims.Email
				roleStr := claims.Role

				objID, err := bson.ObjectIDFromHex(userIDStr)
				if err == nil {
					user, err := GoogleAuth.GetUserByID(ctx, objID)
					if err == nil {
						c.Set("userID", user.ID)
						c.Set("userEmail", user.Email)
						c.Set("role", user.Role)
						c.Set("user", user)
						c.Next()
						return
					}
				}

				c.Set("userID", bson.ObjectID{})
				c.Set("userEmail", email)
				c.Set("role", models.UserRole(roleStr))
				c.Next()
				return
			}
		}

		payload, err := GoogleAuth.VerifyIDToken(ctx, idToken)
		if err != nil {
			log.Printf("Token validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
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
		if name == "" && email != "" {
			name = strings.Split(email, "@")[0]
		}

		user, err := GoogleAuth.FindOrCreateUser(ctx, email, name, picture)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user"})
			c.Abort()
			return
		}

		c.Set("userID", user.ID)
		c.Set("userEmail", user.Email)
		c.Set("role", user.Role)
		c.Set("user", user)

		c.Next()
	}
}
