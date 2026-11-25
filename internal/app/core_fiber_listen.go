package app

import (
	//"embed"
	//"errors"
	//"flag"
	"fmt"
	//"log"
	"net"
	//"net/http"
	//"net/url"
	"os"
	"path"
	//"path/filepath"
	"runtime"
	"strconv"
	//"strings"
	//"time"

	//"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/crypt"
	//"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func CoreFiberListen(app *fiber.App) {

	// -------------------------------------------------------------------------------------------------------------------------------------------

	arg_tls := conf.ArgsMap["tls"].(bool)

	arg_port := conf.ArgsMap["port"].(int)

	arg_host := conf.ArgsMap["host"].(string)

	arg_prefork := conf.ArgsMap["prefork"].(bool)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	crt_filename := path.Join(conf.ConfigRoot, "tls", "server.pem")
	key_filename := path.Join(conf.ConfigRoot, "tls", "server.key")

	if runtime.GOOS == "windows" {
		crt_filename = util.RotateSlash(crt_filename)
		key_filename = util.RotateSlash(key_filename)
	}

	crt_is_exists := false

	if _, err3 := os.Stat(crt_filename); err3 == nil {
		crt_is_exists = true
	}

	if arg_tls && crt_is_exists && !fiber.IsChild() {

		fmt.Println("")
		fmt.Println("  Crt: " + yellow_clr(crt_filename))
		fmt.Println("  Key: " + yellow_clr(key_filename))
		fmt.Println("")
	}

	if arg_tls && !crt_is_exists && !fiber.IsChild() {

		crypt.MakeRSAKeyPairAndSave(path.Join(conf.ConfigRoot, "tls"), "server")

		fmt.Println(yellow_clr("  Generate new TLS keys"))
		fmt.Println("")
		fmt.Println("  Crt: " + yellow_clr(crt_filename))
		fmt.Println("  Key: " + yellow_clr(key_filename))
		fmt.Println("")
	}

	if !fiber.IsChild() {
		fmt.Println("")
		if arg_tls {
			fmt.Println(yellow_clr("  TLS Server port " + strconv.Itoa(arg_port)))
		} else {
			fmt.Println("  Server port " + cian_clr(strconv.Itoa(arg_port)))
		}

		if len(arg_host) > 0 {

			if arg_tls {
				fmt.Println("     https://" + arg_host + ":" + cian_clr(strconv.Itoa(arg_port)))
			} else {
				fmt.Println("     http://" + arg_host + ":" + cian_clr(strconv.Itoa(arg_port)))
			}

		} else {

			fmt.Println("")
			ifaces, err := net.Interfaces()
			if err != nil {
				fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
				//return
				os.Exit(0)
			}
			for _, i := range ifaces {
				addrs, err := i.Addrs()
				if err != nil {
					fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
					continue
				}
				for _, a := range addrs {
					switch v := a.(type) {

					case *net.IPNet:

						if v.IP.To4() != nil {

							if arg_tls {
								fmt.Println("     https://" + v.IP.String() + ":" + cian_clr(strconv.Itoa(arg_port)))
							} else {
								fmt.Println("     http://" + v.IP.String() + ":" + cian_clr(strconv.Itoa(arg_port)))
							}
						}
					}

				}
			}

		}

		fmt.Println("")

		fmt.Println("  Serve folder: " + white_clr(conf.ArgFold))
		fmt.Println("")
		fmt.Println(cian_clr("  [ Control + C ] ") + "Break Server")
		fmt.Println("")
	}

	if arg_tls {

		err := app.Listen(arg_host+":"+strconv.Itoa(arg_port), fiber.ListenConfig{
			CertFile:    crt_filename,
			CertKeyFile: key_filename,

			EnablePrefork:         arg_prefork,
			DisableStartupMessage: true,
		})
		if err != nil {
			fmt.Print(fmt.Errorf("app.ListenTLS: %+v\n", err.Error()))
			//return
			os.Exit(0)
		}

	} else {

		err := app.Listen(arg_host+":"+strconv.Itoa(arg_port), fiber.ListenConfig{

			EnablePrefork:         arg_prefork,
			DisableStartupMessage: true,
		})
		if err != nil {
			fmt.Print(fmt.Errorf("app.Listen: %+v\n", err.Error()))
			//return
			os.Exit(0)
		}
	}

}
