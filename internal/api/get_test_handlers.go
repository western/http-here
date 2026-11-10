package api

import (
	//"os"
	//"path"
	//"path/filepath"
	//"strings"

	//"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/util"
	//"github.com/western/http-here/v2/pkg/safefs"
	fis "github.com/western/http-here/v2/internal/fiberinterstate"

	"github.com/gofiber/fiber/v3"
)

func GetTestGet(c fiber.Ctx) error {

	key := c.Query("key")

	// -------------------------------------------------------------------------------------------------------------------------

	res, err := fis.Get(key)
	if err != nil {

		model2.EventLogAdd(c, 500, "GetTestGet", "Error: "+err.Error())

		return c.JSON(fiber.Map{
			"code": 500,
		})
	}

	model2.EventLogAdd(c, 200, "GetTestGet", "res: "+res)

	return c.JSON(fiber.Map{
		"res": res,
	})
}

func GetTestSet(c fiber.Ctx) error {

	key := c.Query("key")

	value := c.Query("value")

	// -------------------------------------------------------------------------------------------------------------------------

	err := fis.Set(key, value)
	if err != nil {

		model2.EventLogAdd(c, 500, "GetTestSet", "Error: "+err.Error())

		return c.JSON(fiber.Map{
			"code": 500,
		})
	}

	model2.EventLogAdd(c, 200, "GetTestSet", "key: "+key+" value: "+value)

	return c.JSON(fiber.Map{
		"code": 200,
	})
}

func GetTestDel(c fiber.Ctx) error {

	key := c.Query("key")

	// -------------------------------------------------------------------------------------------------------------------------

	err := fis.Delete(key)
	if err != nil {

		model2.EventLogAdd(c, 500, "GetTestDel", "Error: "+err.Error())

		return c.JSON(fiber.Map{
			"code": 500,
		})
	}

	model2.EventLogAdd(c, 200, "GetTestDel", "key: "+key)

	return c.JSON(fiber.Map{
		"code": 200,
	})
}
