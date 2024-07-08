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
	db, err := config.ConnectDatabase()
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// Migrate the schemas
	err = db.AutoMigrate(&models.User{}, &models.Preferences{})
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

		// Preferences, only for authenticated users
		userPreferences := api.Group("/preferences")
		{
			userPreferences.Use(middleware.AuthMiddleware())
			userPreferences.GET("/", controllers.GetPreferences)
			userPreferences.PUT("/", controllers.UpdatePreferences)
		}

		// Themes TODO: protect using auth or consider other protections
		api.GET("/themes", controllers.GetThemes)

		// Layouts TODO: protect using auth or consider other protections
		api.GET("/layouts", controllers.GetLayouts)
		spotifyPlayback := api.Group("/spotify")
		{
			// spotifyPlayback.Use(middleware.AuthMiddleware())
			spotifyPlayback.GET("/auth/login", controllers.GetSpotifyAuthLogin)
			spotifyPlayback.GET("/auth/callback", controllers.GetSpotifyAuthCallback)
			spotifyPlayback.GET("/auth/token", controllers.GetSpotifyAuthToken)
		}
	}

	r.Run()
}
