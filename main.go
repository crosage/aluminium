package main

import (
	"chain/database"
	"chain/handlers"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	database.InitDatabase()
	handlers.InitHandlers(app)
	app.Listen(":20000")
}
