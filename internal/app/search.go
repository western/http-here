package app

import (
	"github.com/western/http-here/internal/model"
	_ "github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

func GetSearch(c *fiber.Ctx) error {

	db := c.Locals("db").(*gorm.DB)

	arg_usedb := ""
	if c.Locals("arg_usedb") != nil {
		arg_usedb = c.Locals("arg_usedb").(string)
	}

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	s := c.Query("s")

	result_list := model.FileSearchResult(db, arg_fold, s)

	model.EventLogAdd(db, c, "200", "GetSearch", "Get search '"+s+"'")

	if db == nil {
		model.EventLogAdd(db, c, "200", "GetSearch", "You should enable database for search")
	}

	return c.Render("view/search", fiber.Map{

		"s":           s,
		"result_list": result_list,
		"arg_usedb":   arg_usedb,
	}, "view/layout/default")
}
