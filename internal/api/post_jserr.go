package api

import (
	//"io"
	"net/url"
	//"os"
	//"path"
	//"regexp"
	//"strings"
	//"fmt"

	//"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v2"
)

func PostJserr(c *fiber.Ctx) error {

	// -------------------------------------------------------------------------------------------------------------------------

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostJserr", "Error url parse "+referer+" "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Error url parse " + referer,
		}, "application/json")
	}

	u_path := util.CleanDirtyPath(u.Path)

	// -------------------------------------------------------------------------------------------------------------------------

	form_path := c.FormValue("path")
	form_path = util.CleanDirtyPath(form_path)
	if len(form_path) > 0 {
		u_path = form_path
	}

	//readTarget := path.Join(conf.ArgFold, u_path)

	// -------------------------------------------------------------------------------------------------------------------------

	/*
		form, err := c.MultipartForm()
		if err != nil {

			model2.EventLogAdd(c, 500, "PostJserr", "Error MultipartForm parse "+u_path+" "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error MultipartForm parse " + u_path,
			}, "application/json")
		}*/

	//msg := form.Value["msg"]
	msg := c.FormValue("msg")

	if len(msg) > 0 {
		model2.EventLogAdd(c, 200, "PostJserr", u_path+" | "+msg)
	} else {
		model2.EventLogAdd(c, 500, "PostJserr", u_path+" | msg value is empty")
	}

	// -------------------------------------------------------------------------------------------------------------------------

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")

}
