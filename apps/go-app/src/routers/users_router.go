package routers

import (
	"github.com/gin-gonic/gin"
	"go-app/src/services"
	"net/http"
)

func RegisterUserRoutes(rg *gin.RouterGroup, userService *services.UserService) {
	rg.GET("/:ulid", func(c *gin.Context) {
		ulid := c.Params.ByName("ulid")
		value, err := userService.GetUserByID(ulid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"id": ulid, "status": "no value"})
		} else {
			c.JSON(http.StatusOK, gin.H{"id": ulid, "value": value})
		}
	})

	rg.GET("/all", func(c *gin.Context) {
		users, err := userService.GetAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "users": users})
	})
}
