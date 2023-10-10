package main

import (
	"deployer/database"
	"deployer/handlers"
	_ "embed"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

func main() {
	godotenv.Load()

	r := gin.Default()
	s := database.NewStore()
	h := handlers.NewHandler(s, VueFiles)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	maxAge := 86400 * 30 // 30 days
	isProd := false      // TODO: Set to true when serving over https
	store := sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd
	gothic.Store = store

	goth.UseProviders(github.New(os.Getenv("GITHUB_CLIENT_KEY"), os.Getenv("GITHUB_SECRET"), os.Getenv("GITHUB_AUTH_CALLBACK_URL")))

	config := cors.DefaultConfig()
	config.AllowOrigins = append(config.AllowOrigins, "http://localhost:3000")
	config.AllowCredentials = true
	config.AllowHeaders = append(config.AllowHeaders, "X-Access-Token")

	r.Use(cors.New(config))
	r.Use(gin.Recovery())
	r.Use(static.Serve("/", EmbedFolder(VueFiles, "frontend/.output/public")))

	r.GET("/info", h.HandleGETInfo)
	r.POST("v1/project", h.HandlePOSTProject)
	r.GET("v1/projects", h.HandleGETProjects)
	r.GET("v1/hook/:unique_project_name", h.HandleGETHook)
	r.GET("v1/auth/:provider", h.HandleGETAuthenticate)
	r.GET("v1/auth/:provider/callback", h.HandleGETAuthenticateCallback)
	r.GET("v1/auth/logout/:provider", h.HandleGETLogout)

	r.Run()
}
