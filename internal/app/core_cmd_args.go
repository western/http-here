package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"
	"github.com/western/http-here/v2/pkg/lopool"
	"github.com/western/http-here/v2/pkg/safefs"

	"github.com/gofiber/fiber/v3"
	"github.com/jellydator/ttlcache/v3"
)

func CoreCmdArgs() {

	flag.Bool("help", false, "Show help")

	flag.String("host", "", "Listen host")
	flag.Int("port", 8000, "Port to use")
	flag.Bool("tls", false, "Start HTTPS")
	flag.Bool("tls-debug", false, "Start HTTPS with verbosity")

	flag.String("login", "", "Login for user basic auth")
	flag.String("password", "", "Password for user basic auth")
	flag.Bool("basic", false, "Set basic auth and generate several accounts every time")
	flag.Bool("basic2", false, "Set basic auth and use database accounts")
	flag.Bool("csrf-disable", false, "Disable CSRF check")

	flag.Bool("upload-disable", false, "Disable upload API and form controller")
	flag.Bool("folder-make-disable", false, "Disable make folder API and form controller")
	flag.Bool("share-only", false, "Set --upload-disable and --folder-make-disable")
	flag.Bool("index-disable", false, "Disable current folder read")

	flag.Bool("extend-mode", false, "Enable delete mechanics. Be very careful. It disabled by default.")
	flag.Bool("crypt", false, "Enable encrypt file support.")

	flag.Bool("prefork", false, "Enable spawn multiple processes")

	flag.Bool("silence", false, "Disable all console messages")
	flag.Bool("nolog", false, "Do not write any data to event_log table")

	flag.Bool("usedb", false, "Database enable")
	flag.Int("cache-dir", 30, "Cache timeout for readdir, seconds")

	flag.String("logto", "", "Filename for log")
	flag.String("jsonto", "", "Filename for json log")

	flag.Parse()

	// -----------------------------------------------------------------------------------------------------------------------------------------

	argsMap := make(map[string]any)

	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if g, ok := f.Value.(flag.Getter); ok {

			argsMap[f.Name] = g.Get()
		} else {
			argsMap[f.Name] = f.Value.String()
		}
	})

	conf.ArgsMap = make(map[string]any)
	conf.ArgsMap = argsMap

	if !fiber.IsChild() {
		//fmt.Printf("%#v\n", argsMap)
	}

	// -----------------------------------------------------------------------------------------------------------------------------------------

	arg_tls_debug := conf.ArgsMap["tls-debug"].(bool)

	arg_share_only := conf.ArgsMap["share-only"].(bool)

	arg_help := conf.ArgsMap["help"].(bool)

	// -----------------------------------------------------------------------------------------------------------------------------------------

	if arg_tls_debug {

		conf.ArgsMap["tls"] = true
	}

	if arg_share_only {

		conf.ArgsMap["upload-disable"] = true
		conf.ArgsMap["folder-make-disable"] = true
	}

	if arg_help {

		inf := []string{
			``,
			green_clr(conf.Version),
			``,
			`Simple zero-configuration command line http server with lightweight interface to work with files`,
			``,
			`usage: ` + green_clr(`http-here`) + ` [options] [path]`,
			``,
			`options:`,
			``,
			`     --host ` + white_clr(`[str]`) + `              Listen host [all interfaces]`,
			`     --port ` + white_clr(`[int]`) + `              Port to use [8000]`,
			`     --tls                     Start HTTPS`,
			``,
			``,
			`     --login ` + white_clr(`[str]`) + `             Login for basic authorization`,
			`     --password ` + white_clr(`[str]`) + `          Password for basic authorization`,
			``,
			`     --basic                   Set basic auth and generate several accounts every time`,
			`     --basic2                  Set basic auth and use database accounts`,
			``,
			`     --csrf-disable            Disable CSRF check`,
			``,
			``,
			`     --index-disable           Disable current folder read`,
			`     --share-only              Set --upload-disable and --folder-make-disable`,
			``,
			``,
			`     --extend-mode             Enable extended mode`,
			``,
			`     --prefork                 Enable spawn multiple processes`,
			``,
			`     --crypt                   Enable encrypt file support`,
			``,
			//`     --spa                     Enable frontend SPA (Single Page Application)`,
			``,
			``,
			`     --silence                 Disable all console messages`,
			``,
			`     --usedb                   Database enable`,
			`     --nolog                   Do not write any data to event_log table`,
			``,
			`     --cache-dir ` + white_clr(`[int]`) + `         Cache timeout for readdir, seconds [30]`,
			``,
			``,
			`     --logto ` + white_clr(`[str]`) + `             Filename for log`,
			`     --jsonto ` + white_clr(`[str]`) + `            Filename for json log`,
			``,
			``,
			`examples:`,
			``,
			`     The safest run`,
			`                        ` + green_clr(`http-here`) + ` --tls --basic ` + white_clr(`/some/path`),
			``,
			`     Only share`,
			`                        ` + green_clr(`http-here`) + ` --share-only ` + white_clr(`/tmp/fold`),
			``,
			`     For work`,
			`                        ` + green_clr(`http-here`) + ` --tls --login ` + white_clr(`login`+util.RandStringRunes(2)) + ` --password ` + white_clr(util.RandStringRunes(12)) + ` --prefork --extend-mode ` + white_clr(`/tmp/fold`),
			``,
			`     Maximum performance`,
			`                        ` + green_clr(`http-here`) + ` --prefork --silence ` + white_clr(`/tmp`),
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		//return
		os.Exit(0)
	}

	// -----------------------------------------------------------------------------------------------------------------------------------------

	arg_logto := conf.ArgsMap["logto"].(string)

	arg_jsonto := conf.ArgsMap["jsonto"].(string)

	arg_usedb := conf.ArgsMap["usedb"].(bool)

	arg_prefork := conf.ArgsMap["prefork"].(bool)

	arg_extend_mode := conf.ArgsMap["extend-mode"].(bool)

	arg_csrf_disable := conf.ArgsMap["csrf-disable"].(bool)

	arg_cache_dir := conf.ArgsMap["cache-dir"].(int)

	// -----------------------------------------------------------------------------------------------------------------------------------------

	if len(arg_logto) > 0 {

		model2.LogtoOpen(arg_logto)
	}

	if len(arg_jsonto) > 0 {

		model2.JsontoOpen(arg_jsonto)
	}

	if arg_usedb && arg_prefork {

		fmt.Println("")
		fmt.Println("You can not run --prefork with --usedb")
		//return
		os.Exit(0)
	}

	if arg_share_only && arg_extend_mode {

		fmt.Println("")
		fmt.Println("You can not run --share-only with --extend-mode")
		//return
		os.Exit(0)
	}

	if !arg_csrf_disable && arg_prefork {

		fmt.Println("")
		fmt.Println("You can not run --prefork with CSRF Secutiry. You can run --prefork with --csrf-disable together")
		//return
		os.Exit(0)
	}

	if len(os.Args) > 1 {

		if os.Args[1] == "user" {

			model2.Open()

			cmdSubcommandUser()
		}

		if os.Args[1] == "usermod" {

			model2.Open()

			cmdSubcommandUserMod()
		}

		if os.Args[1] == "userdel" {

			model2.Open()

			cmdSubcommandUserDel()
		}

		if os.Args[1] == "log" {

			model2.Open()

			cmdSubcommandLog()
		}

		if os.Args[1] == "file" {

			model2.Open()

			cmdSubcommandFile()
		}

		if os.Args[1] == "database" {

			model2.Open()

			cmdSubcommandDatabase()
		}
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// if you not changed cache-dir arg from default 30, the app expect operate with it

	if arg_prefork && arg_cache_dir == 30 {

		arg_cache_dir = 5

		if !fiber.IsChild() {
			fmt.Println()
			fmt.Println("  For prefork automatic set cache-dir=", arg_cache_dir)
		}
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)

		defer ticker.Stop()
		for range ticker.C {
			api.CleanupCacheDir(arg_cache_dir)
		}
	}()

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// arg_fold

	arg_fold, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		//return
		os.Exit(0)
	}

	if len(flag.Args()) > 0 {
		arg_fold = flag.Args()[0]
		if arg_fold, err = filepath.Abs(arg_fold); err != nil {
			fmt.Println(err)
			//return
			os.Exit(0)
		}
	}
	arg_fold = filepath.Clean(arg_fold)

	if runtime.GOOS == "windows" {
		arg_fold = util.RotateSlash(arg_fold)
	}

	if fileInfo, err := os.Stat(arg_fold); err == nil {

		if !fileInfo.IsDir() {
			fmt.Println(arg_fold + " is not folder")
			//return
			os.Exit(0)
		}

	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println(arg_fold + " is not exist")
		//return
		os.Exit(0)
	}

	conf.ArgFold = arg_fold

	rfs, err := safefs.NewSafeFS(arg_fold)
	if err != nil {
		fmt.Println("Safefs err: ", err.Error())
		//return
		os.Exit(0)
	}
	conf.ArgFoldRootFS = rfs

	// -------------------------------------------------------------------------------------------------------------------------------------------

	// -------------------------------------------------------------------------------------------------------------------------------------------

	api.CacheUniqueKeyTTL = ttlcache.New[string, string](
		ttlcache.WithTTL[string, string](30 * time.Minute),
	)
	go api.CacheUniqueKeyTTL.Start()

	if arg_usedb {

		model2.Open()
	}

	if !fiber.IsChild() {

		model2.EventLogAdd(nil, 0, "INIT", "run "+conf.ArgFold)

	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// Libreoffice pool

	//fmt.Println("DefaultPoolSize=", lopool.DefaultPoolSize())

	lopool.LibreOfficePool = lopool.NewLOPool(lopool.DefaultPoolSize())

	// -------------------------------------------------------------------------------------------------------------------------------------------

}
