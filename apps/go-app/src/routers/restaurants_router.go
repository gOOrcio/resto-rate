package routers

import (
	"github.com/gin-gonic/gin"
	"go-app/src/services"
	"net/http"
)

func RegisterRestaurantsRoutes(rg *gin.RouterGroup, restaurantService *services.RestaurantsService) {
	restaurants := rg.Group("/restaurants")
	restaurants.GET("/:ulid", getRestaurantByULIDHandler(restaurantService))
	restaurants.GET("/all", getAllRestaurantsHandler(restaurantService))
}

// @Summary Get restaurant by ULID
// @Description Get restaurant details by ULID
// @Tags restaurants
// @Accept json
// @Produce json
// @Param ulid path string true "ULID of the restaurant"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 404 {object} map[string]string "Not Found"
// @Router /restaurants/{ulid} [get]
func getRestaurantByULIDHandler(restaurantService *services.RestaurantsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		ulid := c.Params.ByName("ulid")
		value, err := restaurantService.GetRestaurantByID(ulid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"id": ulid, "status": "no value"})
		} else {
			c.JSON(http.StatusOK, gin.H{"id": ulid, "value": value})
		}
	}
}

// @Summary Get all restaurants
// @Description Retrieve a list of all restaurants
// @Tags restaurants
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 500 {object} map[string]string "Internal Server Error"
// @Router /restaurants/all [get]
func getAllRestaurantsHandler(restaurantService *services.RestaurantsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		restaurants, err := restaurantService.GetAllRestaurants()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "restaurants": restaurants})
	}
}
