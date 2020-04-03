package main

import (
	"fmt"
	"web2/handler"
	"web2/middleware"

	gormadapter "github.com/casbin/gorm-adapter/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	// Initialize  casbin adapter
	adapter, err := gormadapter.NewAdapterByDB(handler.DB)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize casbin adapter: %v", err))
	}
	router := gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = true
	router.Use(cors.New(corsConfig)) // CORS configuraion
	router.POST("/user/login", handler.Login)
	v1 := router.Group("api/v1/web")
	{
		v1.POST("/", middleware.Authorize("admin", "POST", adapter), handler.CreateUser)
		v1.GET("/", middleware.Authorize("admin", "GET", adapter), handler.GetAllUser)
		v1.GET("/:id", middleware.Authorize("user", "GET", adapter), handler.GetUser)
		v1.PUT("/:id", handler.ModifyUser)
		v1.DELETE("/:id", handler.DeleteUser)
	}
	router.Run()
}
