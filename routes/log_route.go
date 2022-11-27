package routes

import (
	"backend-log-api/controllers"

	"github.com/gin-gonic/gin"
)

func LogRoutes(router *gin.Engine) {
	router.POST("/log", controllers.PostLog())
	router.GET("/log", controllers.GetAllLog())
	router.GET("/count/connected", controllers.GetConnected())
	router.GET("/count/failed", controllers.GetFailed())
	router.POST("/connected-user", controllers.PostCurrentUser())
	router.GET("/connected-user", controllers.GetAllCurrentUser())
	router.POST("/login", controllers.PostLogin())
}
