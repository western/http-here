package app

import (
	"github.com/gofiber/fiber/v2"
)

// -------------------------------------------------------------------------------------------------------------------------------------------

func GetBoolFromLocals(c *fiber.Ctx, key string) bool {
	if val, ok := c.Locals(key).(bool); ok {
		return val
	}
	return false
}

func GetStringFromLocals(c *fiber.Ctx, key string, defaultValue string) string {
	if val, ok := c.Locals(key).(string); ok {
		return val
	}
	return defaultValue
}

func GetIntFromLocals(c *fiber.Ctx, key string, defaultValue int) int {
	if val, ok := c.Locals(key).(int); ok {
		return val
	}
	return defaultValue
}

// -------------------------------------------------------------------------------------------------------------------------------------------
