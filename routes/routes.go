package routes

import (
	"bikincetak-api/controllers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App)  {
	api:= app.Group("/v1")

	api.Get("/items", controllers.GetItems)
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
}