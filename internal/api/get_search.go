package api

import (
	"github.com/western/http-here/internal/model"
	"github.com/western/http-here/internal/util"

	"github.com/gofiber/fiber/v2"
)

func GetSearch(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	db, err := model.ConnectToSQLite()
	if err != nil {
		panic(err)
	}

	// -------------------------------------------------------------------------------------------------------------------------

	s := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	result_list := model.FileSearchResult(db, arg_fold, s)

	util.LogPrefix(c, "200", "Get api search '"+s+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	}, "application/json")

}
