package api

import (
	"github.com/western/http-here/internal/model"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func GetSearch(c *fiber.Ctx) error {

	arg_fold := GetStringFromLocals(c, "arg_fold", "")

	db := c.Locals("db").(*gorm.DB)

	// -------------------------------------------------------------------------------------------------------------------------

	s := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	result_list := model.FileSearchResult(db, arg_fold, s)

	model.EventLogAdd(db, c, "200", "GetSearch", "Get api search '"+s+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	}, "application/json")

}
