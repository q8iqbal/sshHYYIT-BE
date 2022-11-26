package routes

import (
	"backend-log-api/controllers"

	"github.com/gin-gonic/gin"
)

func LogRoutes(router *gin.Engine) {
	router.POST("/log", controllers.PostLog())
	router.GET("/log/:logId", controllers.GetALog())
	router.GET("/logs", controllers.GetAllLog())
}
