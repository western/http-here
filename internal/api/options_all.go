

package api

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

func OptionsAll(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"code":   200,
		"method": "OPTIONS",
	}, "application/json")
}

