package api

import (
	"github.com/western/http-here/internal/model"
	_ "github.com/western/http-here/internal/util"

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

	
	model.EventLogAdd(db, c, "200", "GetSearch", "Get api search '"+s+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	}, "application/json")

}
