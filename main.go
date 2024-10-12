package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/aseiilkhan/virtual-study-room/config"
	"github.com/aseiilkhan/virtual-study-room/controllers"
	"github.com/aseiilkhan/virtual-study-room/middleware"
	"github.com/aseiilkhan/virtual-study-room/models"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Failed to load environment variables from .env: " + err.Error())
	}

	// Connect to the database, retry up to 3 times
	config.DB_init()
	db := config.DB

	// Migrate the schemas
	err = db.AutoMigrate(&models.User{}, &models.Preferences{}, &models.State{})
	if err != nil {
		panic("Failed to migrate database: " + err.Error())
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Router setup

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Access-Control-Allow-Origin"}

	r.Use(cors.New(config))
	// Error handling
	r.Use(middleware.ErrorHandler())

	api := r.Group("/api")
	{
		// Authentication
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)
		api.POST("/refresh", controllers.RefreshToken)

		authorized := api.Group("/")
		authorized.Use(controllers.ValidateJWT())
		{
			authorized.POST("/protected", controllers.Protected)
		}

		spotifyPlayback := api.Group("/spotify")
		spotifyPlayback.Use(controllers.ValidateJWT())
		{
			spotifyPlayback.GET("/auth/login", controllers.GetSpotifyAuthLogin)
			spotifyPlayback.GET("/auth/token", controllers.GetSpotifyAuthToken)
			spotifyPlayback.GET("/auth/refresh", controllers.GetSpotifyRefreshToken)
		}

		api.GET("/spotify/auth/callback", controllers.GetSpotifyAuthCallback)
	}

	r.Run()
}
