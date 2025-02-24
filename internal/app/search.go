package app

import (
	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	s := c.Query("s")

	result_list := model.FileSearchResult(db, arg_fold, s)

	util.LogPrefix(c, "200", "Get search '"+s+"'")

	return c.Render("view/search", fiber.Map{

		"s":           s,
		"result_list": result_list,
	}, "view/layout/default")
}
