package app

import (
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/conf"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	arg_usedb := GetBoolFromLocals(c, "arg_usedb")

	s := c.Query("s")

	result_list := model2.FileSearchResult(s)

	model2.EventLogAdd(c, 200, "GetSearch", "Get search '"+s+"'")

	/*
		if db == nil {
			model2.EventLogAdd( c, "200", "GetSearch", "You should enable database for search")
		}*/

	return c.Render("view/search", fiber.Map{

		"s":           s,
		"result_list": result_list,
		"arg_usedb":   arg_usedb,
	}, "view/layout/default")
}
