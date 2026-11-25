package app

import (
	//"embed"
	//"errors"
	//"flag"
	//"fmt"
	//"net"
	//"net/http"
	//"net/url"
	//"os"
	//"path"
	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"
	//"syscall"
	//"time"

	//"github.com/western/http-here/v2/internal/api"
	//"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	//fis "github.com/western/http-here/v2/internal/fiberinterstate"
	//"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/util"
	//"github.com/western/http-here/v2/pkg/lopool"
	//"github.com/western/http-here/v2/pkg/safefs"

	"github.com/fatih/color"
	//"github.com/gofiber/fiber/v3"
	//"github.com/gofiber/fiber/v3/extractors"
	//"github.com/gofiber/fiber/v3/middleware/basicauth"
	//"github.com/gofiber/fiber/v3/middleware/compress"
	//"github.com/gofiber/fiber/v3/middleware/csrf"
	//"github.com/gofiber/fiber/v3/middleware/favicon"
	//"github.com/gofiber/fiber/v3/middleware/static"
	//"github.com/gofiber/contrib/v3/monitor"
	//"github.com/gofiber/template/html/v3"
	//"github.com/jedib0t/go-pretty/v6/table"
	//"github.com/jellydator/ttlcache/v3"
	//badgerstorage "github.com/gofiber/storage/badger/v2"
)

var (
	green_clr  = color.New(color.FgGreen).SprintFunc()
	white_clr  = color.New(color.Bold, color.FgWhite).SprintFunc()
	cian_clr   = color.New(color.FgCyan).SprintFunc()
	yellow_clr = color.New(color.FgYellow).SprintFunc()
)

func Core() {

	CoreCmdArgs()

	CoreMakeFolders()

	app := CoreFiberInit()

	CoreFiberMiddleware(app)

	CoreFiberRoute(app)

	CoreFiberListen(app)
}
