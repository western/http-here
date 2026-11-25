package app

import (
	"embed"
	"net/http"
	"time"

	//"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	//"github.com/western/http-here/v2/internal/model2"
	//"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v3"
)

//go:embed view/*
var view_fs embed.FS

func CoreFiberInit() *fiber.App {

	// -----------------------------------------------------------------------------------------------------------------------------------------

	arg_upload_disable := conf.ArgsMap["upload-disable"].(bool)
	arg_folder_make_disable := conf.ArgsMap["folder-make-disable"].(bool)
	arg_extend_mode := conf.ArgsMap["extend-mode"].(bool)
	arg_crypt := conf.ArgsMap["crypt"].(bool)

	arg_silence := conf.ArgsMap["silence"].(bool)
	arg_nolog := conf.ArgsMap["nolog"].(bool)
	arg_usedb := conf.ArgsMap["usedb"].(bool)
	arg_cache_dir := conf.ArgsMap["cache-dir"].(int)
	arg_csrf_disable := conf.ArgsMap["csrf-disable"].(bool)

	// -----------------------------------------------------------------------------------------------------------------------------------------

	fiberConfig := fiber.Config{

		ServerHeader: "",
		BodyLimit:    conf.FieldSizeMax,

		// Timeouts
		ReadTimeout:  30 * time.Second, // nil, The amount of time allowed to read the full request, including the body. The default timeout is unlimited.
		WriteTimeout: 30 * time.Second, // nil, The maximum duration before timing out writes of the response. The default timeout is unlimited.
		//IdleTimeout:     120 * time.Second, // nil, The maximum amount of time to wait for the next request when keep-alive is enabled. If IdleTimeout is zero, the value of ReadTimeout is used.
		IdleTimeout:     30 * time.Second,
		ReadBufferSize:  8192, // 4096
		WriteBufferSize: 8192, // 4096

		// Enable request/response pooling
		//EnableTrustedProxyCheck: false,
		//ProxyHeader:            "X-Forwarded-For",

		// Optimize for high concurrency
		//Concurrency:           256 * 1024, // 256K concurrent connections

		// Disable features for better performance
		//DisablePreParseMultipartForm: true, // false
		//StreamRequestBody:            true, // false

		//SharedStorage:     redisStorage,
	}

	fiberConfig.Views = html.NewFileSystem(http.FS(view_fs), ".html")

	app := fiber.New(fiberConfig)

	app.Use(func(c fiber.Ctx) error {

		c.Locals("arg_upload_disable", arg_upload_disable)
		c.Locals("arg_folder_make_disable", arg_folder_make_disable)
		c.Locals("arg_extend_mode", arg_extend_mode)
		c.Locals("arg_crypt", arg_crypt)

		c.Locals("arg_silence", arg_silence)
		c.Locals("arg_nolog", arg_nolog)
		c.Locals("arg_usedb", arg_usedb)
		c.Locals("arg_cache_dir", arg_cache_dir)
		c.Locals("arg_csrf_disable", arg_csrf_disable)

		return c.Next()
	})

	return app
}
