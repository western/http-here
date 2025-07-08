package app

import (
	"github.com/western/http-here/internal/conf"

	"github.com/gofiber/fiber/v2"
)

func GetSpa(c *fiber.Ctx) error {

	arg_upload_disable := GetBoolFromLocals(c, "arg_upload_disable")
	arg_folder_make_disable := GetBoolFromLocals(c, "arg_folder_make_disable")
	arg_extend_mode := GetBoolFromLocals(c, "arg_extend_mode")

	return c.Render("view/spa", fiber.Map{

		"files_count_max":     conf.Files_count_max,
		"fieldSize_max":       conf.FieldSize_max,
		"fieldSize_max_human": conf.FieldSize_max_human,

		"arg_upload_disable":      arg_upload_disable,
		"arg_folder_make_disable": arg_folder_make_disable,
		"arg_extend_mode":         arg_extend_mode,
	})

}
