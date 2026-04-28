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

	auth:= api.Group("auth")
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Get("google/login", controllers.GoogleLogin)
	auth.Get("google/callback", controllers.GoogleCallback)

	api.Use(middleware.Protected())
	
	api.Post("/cart", controllers.AddToCart)
	api.Get("/cart", controllers.GetCart)
	api.Put("cart/:id", controllers.UpdateCartItem)     
	api.Delete("cart/:id", controllers.DeleteCartItem)

}