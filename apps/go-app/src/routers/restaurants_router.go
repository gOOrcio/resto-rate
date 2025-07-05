package routers

import (
	"github.com/gin-gonic/gin"
	"go-app/src/services"
	"net/http"
)

func RegisterRestaurantsRoutes(rg *gin.RouterGroup, restaurantService *services.RestaurantsService) {
	rg.GET("/:ulid", func(c *gin.Context) {
		ulid := c.Params.ByName("ulid")
		value, err := restaurantService.GetRestaurantByID(ulid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"id": ulid, "status": "no value"})
		} else {
			c.JSON(http.StatusOK, gin.H{"id": ulid, "value": value})
		}
	})

	rg.GET("/all", func(c *gin.Context) {
		restaurants, err := restaurantService.GetAllRestaurants()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "restaurants": restaurants})
	})
}
