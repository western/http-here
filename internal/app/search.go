package app

import (
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/conf"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	s := c.Query("s")

	//result_list := model2.FileSearchResult(s)
	result_list := model2.FileFastSearchResult(s)

	model2.EventLogAdd(c, 200, "GetSearch", "Get search '"+s+"'")

	return c.Render("view/search", fiber.Map{

		"s":           s,
		"result_list": result_list,
	}, "view/layout/default")
}
