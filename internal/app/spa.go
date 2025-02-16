package app

import (
	_ "encoding/json"
	_ "errors"
	_ "fmt"
	_ "html/template"
	_ "io"
	_ "log"
	_ "net/url"
	_ "os"
	_ "path/filepath"
	_ "reflect"
	_ "regexp"
	_ "sort"
	_ "strings"
	_ "time"

	_ "github.com/western/http-here/internal/model"

	"github.com/gofiber/fiber/v2"
)

func GetSpa(c *fiber.Ctx) error {

	return c.Render("view/spa", fiber.Map{

		"files_count_max":     100,
		"fieldSize_max":       14 * 1024 * 1024 * 1024,
		"fieldSize_max_human": "14 Gb",

		//"arg_upload_disable":      arg_upload_disable,
		//"arg_folder_make_disable": arg_folder_make_disable,
	})

}
