package routes

import (
	"bikincetak-api/controllers"
	"bikincetak-api/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App)  {
	api:= app.Group("/v1")

	api.Get("/items", controllers.GetItems)
	api.Get("/items/:name", controllers.GetDetailItem)
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)

	api.Use(middleware.Protected())
	
	api.Post("/cart", controllers.AddToCart)
	api.Get("/cart", controllers.GetCart)

}