package routers

import (
	"github.com/gin-gonic/gin"
	"go-app/src/services"
	"net/http"
)

func RegisterUserRoutes(rg *gin.RouterGroup, userService *services.UserService) {
	users := rg.Group("/users")
	users.GET("/:ulid", getUserByULIDHandler(userService))
	users.GET("/all", getAllUsersHandler(userService))
}

// @Summary Get user by ULID
// @Description Get user details by ULID
// @Tags users
// @Accept json
// @Produce json
// @Param ulid path string true "ULID of the user"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /users/{ulid} [get]
func getUserByULIDHandler(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ulid := c.Params.ByName("ulid")
		value, err := userService.GetUserByID(ulid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"id": ulid, "status": "no value"})
		} else {
			c.JSON(http.StatusOK, gin.H{"id": ulid, "value": value})
		}
	}
}

// @Summary Get all users
// @Description Retrieve a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /users/all [get]
func getAllUsersHandler(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userService.GetAllUsers()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "users": users})
	}
}
