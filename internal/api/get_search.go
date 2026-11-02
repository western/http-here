package api

import (
	"github.com/western/http-here/v2/internal/model2"

	"github.com/gofiber/fiber/v3"
)

// TODO: deprecated?
// It is work for SPA
func GetSearch(c fiber.Ctx) error {

	searchString := c.Query("s")

	// -------------------------------------------------------------------------------------------------------------------------

	result_list := FileFastSearchResult(searchString)

	model2.EventLogAdd(c, 200, "GetSearch", "Get api search '"+searchString+"'")

	return c.JSON(fiber.Map{
		"code":        200,
		"result_list": result_list,
	})

}
