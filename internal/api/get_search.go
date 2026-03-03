package api

import (
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	// -------------------------------------------------------------------------------------------------------------------------

	s := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	result_list := model2.FileSearchResult(s)

	model2.EventLogAdd(c, 200, "GetSearch", "Get api search '"+s+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	}, "application/json")

}
