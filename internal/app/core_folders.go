package app

import (
	//"embed"
	//"errors"
	//"flag"
	"fmt"
	//"log"
	//"net"
	//"net/http"
	//"net/url"
	"os"
	"path"
	//"path/filepath"
	//"runtime"
	//"strconv"
	//"strings"
	"time"

	//"github.com/western/http-here/v2/internal/api"
	"github.com/western/http-here/v2/internal/conf"
	//"github.com/western/http-here/v2/internal/crypt"
	//"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/util"
	"github.com/western/http-here/v2/pkg/safefs"

	"github.com/gofiber/fiber/v3"
)

func CoreMakeFolders() {

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// FOLDERS

	if !fiber.IsChild() {

		if _, err := os.Stat(path.Join(conf.ConfigRoot)); err != nil {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot), os.ModePerm); err != nil {
				fmt.Println(err)
				//return
				os.Exit(0)
			}
		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "thumb")); err != nil {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot, "thumb"), os.ModePerm); err != nil {
				fmt.Println(err)
				//return
				os.Exit(0)
			}
		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "tls")); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot, "tls"), os.ModePerm); err != nil {
				fmt.Println(err)
				//return
				os.Exit(0)
			}
		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "temp")); err != nil {

			fmt.Println("")
			//fmt.Println("  Make temp folder")

			if err = os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {
				fmt.Println(err)
				//return
				os.Exit(0)
			}
		} else {

			fmt.Println("")
			//fmt.Println(yellow_clr("  Clear temp folder"))

			/*
				if err = os.RemoveAll(path.Join(conf.ConfigRoot, "temp")); err != nil {
					fmt.Println(err)
					return
				}

				if err = os.MkdirAll(path.Join(conf.ConfigRoot, "temp"), os.ModePerm); err != nil {
					fmt.Println(err)
					return
				}
			*/
			go util.WalkAndClearOld(path.Join(conf.ConfigRoot, "temp"), 1*time.Hour)

		}

		if _, err := os.Stat(path.Join(conf.ConfigRoot, "parts")); os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(conf.ConfigRoot, "parts"), os.ModePerm); err != nil {
				fmt.Println(err)
				//return
				os.Exit(0)
			}
		} else {

			go util.WalkAndClearOld(path.Join(conf.ConfigRoot, "parts"), 1*time.Hour)
		}

	}

	/*
		if !fiber.IsChild() {

			util.WalkAndClearZeroFile(path.Join(conf.ConfigRoot, "thumb"), 0)

		}
	*/

	thumbFS, err := safefs.NewSafeFS(path.Join(conf.ConfigRoot, "thumb"))
	if err != nil {
		fmt.Println("Safefs err: ", err.Error())
		//return
		os.Exit(0)
	}
	conf.ThumbRootFS = thumbFS

	tempFS, err := safefs.NewSafeFS(path.Join(conf.ConfigRoot, "temp"))
	if err != nil {
		fmt.Println("Safefs err: ", err.Error())
		//return
		os.Exit(0)
	}
	conf.TempRootFS = tempFS

}
