package api

import (
	"github.com/gofiber/fiber/v2"
)

func OptionsAll(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"code":   200,
		"method": "OPTIONS",
	}, "application/json")
}
