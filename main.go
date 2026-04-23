package main

import (
	"bikincetak-api/database"
	"bikincetak-api/routes"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Env gak kebaca")
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",}))

	database.ConnectDB()
	routes.SetupRoutes(app)
	

	fmt.Println("Server sedang berjalan di Port: 3000")
	log.Fatal(app.Listen(":3000"))
}