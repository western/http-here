package app

import (
	"embed"
	//"errors"
	//"flag"
	"fmt"
	//"log"
	//"net"
	//"net/http"
	"net/url"
	"os"
	"path"
	//"path/filepath"
	//"runtime"
	"strconv"
	//"strings"
	//"time"

	//"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	//"github.com/gofiber/contrib/v3/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jedib0t/go-pretty/v6/table"
)

//go:embed assets/*
var embedDirStatic embed.FS

func CoreFiberMiddleware(app *fiber.App) {

	// -------------------------------------------------------------------------------------------------------------------------------------------

	arg_csrf_disable := conf.ArgsMap["csrf-disable"].(bool)

	arg_basic := conf.ArgsMap["basic"].(bool)

	arg_basic2 := conf.ArgsMap["basic2"].(bool)

	arg_prefork := conf.ArgsMap["prefork"].(bool)

	arg_usedb := conf.ArgsMap["usedb"].(bool)

	arg_login := conf.ArgsMap["login"].(string)

	arg_password := conf.ArgsMap["password"].(string)

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// MIDDLEWARE

	/*
		app.Use(fis.New(fis.Config{

			App: app,
			//BadgerDirectory: path.Join(conf.ConfigRoot, "db", "registry.bd."+conf.Version),
			BadgerDirectory: path.Join(os.TempDir(), "db_test"),
			//LogEnable: true,
		}))
	*/

	app.Use(favicon.New(favicon.Config{
		File: "",
		URL:  "/favicon.ico",
		Data: []byte{},
	}))

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // 1
	}))

	if !arg_csrf_disable {

		app.Use(csrf.New(csrf.Config{
			CookieName: "csrf_",
			//CookieName:        "__Host-csrf_",
			//CookieSecure:      true,
			//CookieHTTPOnly:    true,       // Required for JavaScript access to tokens
			//CookieSameSite:    "Lax",
			//CookieSessionOnly: true,
			Extractor: extractors.FromHeader("csrf-token"),
			//Session:           sessionStore,

			ErrorHandler: func(c fiber.Ctx, err error) error {

				c_Path, err := url.QueryUnescape(c.Path())
				if err != nil {

					model2.EventLogAdd(c, 500, "CSRF_ERR", "Error "+c_Path+" "+err.Error())

					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{})
				}

				model2.EventLogAdd(c, 403, "CSRF_ERR", "Cross-Site Request Forgery Error "+c_Path)

				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"code": 403,
					"msg":  "Forbidden",
				})
			},
		}))
	}

	//app.Get("/metrics", monitor.New())

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// ASSETS

	app.Get("/__assets*", static.New("", static.Config{
		FS:     embedDirStatic,
		Browse: false,
	}))

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// BASIC AUTH

	if arg_basic && arg_prefork {

		fmt.Println("")
		fmt.Println("You can not run --prefork with --basic. Use --user and --password instead.")
		//return
		os.Exit(0)
	}

	if arg_basic && !arg_prefork {

		fmt.Println("")
		fmt.Println("  Basic auth set: ")

		var password_list = map[string]string{}

		for i := range 10 {

			login := "login" + strconv.Itoa(i) + util.RandStringRunes(2)
			password := util.RandStringRunes(16)

			password_list[login] = crypt.GenerateHashPassword(password)

			fmt.Println("         " + login + "    " + password)
		}

		app.Use(basicauth.New(basicauth.Config{

			Users: password_list,

			Unauthorized: func(c fiber.Ctx) error {

				c_Path, err := url.QueryUnescape(c.Path())
				if err != nil {

					model2.EventLogAdd(c, 500, "basicauth", "Error "+c_Path+" "+err.Error())
					return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
				}

				model2.EventLogAdd(c, 401, "basicauth", path.Join(conf.ArgFold, c_Path))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
			},
		}))

	}

	if arg_basic2 {

		if !arg_usedb {

			fmt.Println("")
			fmt.Println("You can not run --basic2 without --usedb")
			//return
			os.Exit(0)
		}

		fmt.Println("")
		fmt.Println("  Basic auth set: ")

		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.AppendHeader(table.Row{"#", "Login", "Password", "Label"})

		filt := []map[string]interface{}{
			//{"Login": "login2"},
			{"Enabled": true},
		}
		user_list := model2.UserSelect(filt)

		var password_list = map[string]string{}
		for _, el := range user_list {

			password_list[el.Login] = crypt.GenerateHashPassword(el.Password)

			t.AppendRows([]table.Row{
				{el.ID, el.Login, el.Password, el.Label},
			})
		}
		fmt.Println(t.Render())

		app.Use(basicauth.New(basicauth.Config{

			Users: password_list,

			Unauthorized: func(c fiber.Ctx) error {

				c_Path, err := url.QueryUnescape(c.Path())
				if err != nil {

					model2.EventLogAdd(c, 500, "basicauth", "Error "+c_Path+" "+err.Error())
					return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
				}

				model2.EventLogAdd(c, 401, "basicauth", path.Join(conf.ArgFold, c_Path))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
			},
		}))
	}

	if len(arg_login) > 0 && len(arg_password) > 0 {

		app.Use(basicauth.New(basicauth.Config{

			Authorizer: func(user, pass string, c fiber.Ctx) bool {
				if user == arg_login && pass == arg_password {
					return true
				}

				return false
			},
			Unauthorized: func(c fiber.Ctx) error {

				c_Path, err := url.QueryUnescape(c.Path())
				if err != nil {

					model2.EventLogAdd(c, 500, "basicauth", "Error "+c_Path+" "+err.Error())
					return c.Status(fiber.StatusInternalServerError).Render("view/500", fiber.Map{}, "view/layout/error")
				}

				model2.EventLogAdd(c, 401, "basicauth", path.Join(conf.ArgFold, c_Path))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout/error")
			},
		}))

		if !fiber.IsChild() {
			fmt.Println("")
			fmt.Println("  Basic auth set: " + cian_clr(arg_login) + " " + cian_clr(arg_password))
		}
	}

}
