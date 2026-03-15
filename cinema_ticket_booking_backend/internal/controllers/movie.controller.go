package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/peerapat012/CinemaTicketBooking/internal/database"
	"github.com/peerapat012/CinemaTicketBooking/internal/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var movieCollection *mongo.Collection = database.OpenCollection("movies")

func GetMovies() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movies []models.Movie

		cursor, err := movieCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies."})
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &movies); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode movies."})
		}

		c.JSON(http.StatusOK, gin.H{"message": "Movies fetched successfully.", "data": movies})
	}
}

func GetMovieByID() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		movieID := c.Params.ByName("_id")

		if movieID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Movie ID is required."})
			return
		}

		var movie models.Movie
		if err := movieCollection.FindOne(ctx, bson.M{"_id": movieID}).Decode(&movie); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movie."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Movie fetched successfully.", "data": movie})
	}
}

func CreateMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var movie models.Movie
		if err := c.ShouldBindJSON(&movie); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload."})
			return
		}

		if _, err := movieCollection.InsertOne(ctx, movie); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create movie."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Movie created successfully."})
	}
}

func UpdateMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	}
}

func DeleteMovie() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	}
}
