package main

import (
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fatih/color"

	"github.com/western/http-here/controller"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	_ "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

//go:embed view/*
var view_fs embed.FS

//go:embed assets/*
var embedDirStatic embed.FS

func main() {

	arg_port := flag.Int("port", 8000, "Change default listen port")
	
	arg_user := flag.String("user", "", "Login for user basic auth")
	arg_password := flag.String("password", "", "Password for user basic auth")
	arg_basic := flag.Bool("basic", false, "Set basic auth and generate several accounts every time")
	
	arg_help := flag.Bool("help", false, "Show help")
	
	arg_upload_disable := flag.Bool("upload-disable", false, "Disable upload API and form controller")
	arg_folder_make_disable := flag.Bool("folder-make-disable", false, "Disable make folder API and form controller")
	arg_index_disable := flag.Bool("index-disable", false, "Disable current folder read")
	
	arg_tls := flag.Bool("tls", false, "Start HTTPS (need easyrsa linux package)")

	flag.Parse()

	if *arg_help {

		inf := []string{
			``,
			`v1.0.9`,
			``,
			`usage: http-here [options] [path]`,
			``,
			`options:`,
			`     --port                    Port to use [8000]`,
			``,
			`     --user                    Login for basic authorization.`,
			`     --password                Password for basic authorization.`,
			``,
			`     --basic                   Set basic auth and generate several accounts every time.`,
			``,
			`     --upload-disable          Disable upload API and form controller.`,
			`     --folder-make-disable     Disable make folder API and form controller.`,
			`     --index-disable           Disable current folder read.`,
			``,
			`     --tls                     Start HTTPS (need easyrsa linux package).`,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		return
	}

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	arg_fold := filepath.Dir(ex)

	if len(flag.Args()) > 0 {
		arg_fold = flag.Args()[0]
	}
	arg_fold = filepath.Clean(arg_fold)

	if fileInfo, err := os.Stat(arg_fold); err == nil {

		if !fileInfo.IsDir() {
			fmt.Println(arg_fold + " is not folder")
			return
		}

	} else if errors.Is(err, os.ErrNotExist) {
		fmt.Println(arg_fold + " is not exist")
		return
	}

	

	//engine := html.New("./view", ".html")
	engine := html.NewFileSystem(http.FS(view_fs), ".html")

	config := fiber.Config{
		Prefork:               false,
		DisableStartupMessage: true,
		ServerHeader:          "",
		Views:                 engine,
		BodyLimit:             7 * 1024 * 1024 * 1024,
	}

	app := fiber.New(config)

	app.Use(func(c *fiber.Ctx) error {

		
		c.Locals("arg_fold", arg_fold)

		if *arg_upload_disable {
			
			c.Locals("arg_upload_disable", "1")
		}
		if *arg_folder_make_disable {
			
			c.Locals("arg_folder_make_disable", "1")
		}

		return c.Next()
	})

	/*
	   app.Use([]string{"/api", "/"}, func(c *fiber.Ctx) error {

	       c.Locals("arg_fold", arg_fold)


	       if *arg_upload_disable {
	   		c.Locals("arg_upload_disable", "1")
	   	}
	   	if *arg_folder_make_disable {
	   		c.Locals("arg_folder_make_disable", "1")
	   	}

	       return c.Next()
	   })*/

	app.Use(cors.New(cors.Config{
		//AllowOrigins: "*",
		AllowCredentials: true,
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
		AllowMethods:  "*",
		AllowHeaders:  "*",
		ExposeHeaders: "*",
	}))

	cian := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	
	if *arg_basic {
	    
	    fmt.Println("")
		fmt.Println("  Basic auth set: ")
	    
	    var data = map[string]string{}
	    
	    for i := range 10 {
	        
	        login := "login" + strconv.Itoa(i) + controller.RandStringRunes(2)
	        password := controller.RandStringRunes(16)
	        
	        data[login] = password
	        
	        fmt.Println("         "+login+"    "+password)
	    }
	    
	    app.Use(basicauth.New(basicauth.Config{

			Users: data,
			
			Unauthorized: func(c *fiber.Ctx) error {

				controller.LogPrefix(c, "401", filepath.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout")
			},
		}))
	    
	}

	if len(*arg_user) > 0 && len(*arg_password) > 0 {

		app.Use(basicauth.New(basicauth.Config{

			Authorizer: func(user, pass string) bool {
				if user == *arg_user && pass == *arg_password {
					return true
				}

				return false
			},
			Unauthorized: func(c *fiber.Ctx) error {

				controller.LogPrefix(c, "401", filepath.Join(arg_fold, c.Path()))

				c.Set(fiber.HeaderWWWAuthenticate, "Basic realm='Restricted'")
				return c.Status(fiber.StatusUnauthorized).Render("view/401", fiber.Map{}, "view/layout")
			},
		}))

		fmt.Println("")
		fmt.Println("  Basic auth set: " + cian(*arg_user) + " " + cian(*arg_password))
	}

	app.Use("/__assets", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDirStatic),
		PathPrefix: "",
		Browse:     false,
	}))

	app.Options("/*", controller.OptionsAll)

	if !*arg_index_disable {
		app.Get("/*", controller.GetAll)
	}

	if !*arg_upload_disable {
		app.Post("/api/upload", controller.PostUpload)
	}

	if !*arg_folder_make_disable {
		app.Post("/api/folder", controller.PostFolder)
	}

	app.Use(func(c *fiber.Ctx) error {

		
		controller.LogPrefix(c, "404", filepath.Join(arg_fold, c.Path()))

		return c.Status(fiber.StatusNotFound).Render("view/404", fiber.Map{}, "view/layout")
	})

	
	

	
	
	
	
	// /home/andrew/.httphere/easyrsa/pki/issued/server1.crt
	// /home/andrew/.httphere/easyrsa/pki/private/server1.key
	
	homepath, err2 := os.UserHomeDir()
    if err2 != nil {
        log.Fatal( err2 )
    }
    //fmt.Println( homepath )
	
	crt_filename := filepath.Join(homepath, ".httphere", "easyrsa", "pki", "issued", "server1.crt")
	key_filename := filepath.Join(homepath, ".httphere", "easyrsa", "pki", "private", "server1.key")
	
	crt_is_exists := false
	
	
	if _, err3 := os.Stat(crt_filename); err3 == nil {
	    crt_is_exists = true
	}
	
	
	if *arg_tls && crt_is_exists {
	    
	    fmt.Println("")
	    fmt.Println("     Crt: " + yellow(crt_filename))
	    fmt.Println("     Key: " + yellow(key_filename))
	    fmt.Println("")
	}
	
	
	if *arg_tls && !crt_is_exists {
	    
	    //log.Fatal(crt_is_exists)
	    
	    _, err1 := exec.Command("bash", "-c", "easyrsa --help").Output()
    	if err1 != nil {
    		log.Fatal(err1)
    	}
    	//fmt.Printf("The date is %s\n", out)
    	
    	
    	
    	
	    
	    if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "easyrsa")); err3 != nil {
	        
	        if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "easyrsa"), os.ModePerm); err4 != nil {
        		log.Fatal( err4 )
        	}
	    }
	    
	    
	    _, err5 := exec.Command("bash", "-c", "cd "+filepath.Join(homepath, ".httphere", "easyrsa")).Output()
    	if err5 != nil {
    		log.Fatal(err5)
    	}
    	//fmt.Printf("The date is %s\n", out)
    	
    	
    	
    	
    	cmd := exec.Command("bash", "-c", "easyrsa init-pki")
        cmd.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
        out3, _ := cmd.Output()
    	
    	fmt.Println(yellow("--------------------------------------------------------------------------------------------------"))
    	fmt.Printf(" %s\n", out3)
    	
    	
    	var_data := `
set_var EASYRSA_DN "cn_only"
set_var EASYRSA_KEY_SIZE 2048
set_var EASYRSA_REQ_CN   "ca@desec.example.com"
set_var EASYRSA_BATCH    "yes"
set_var EASYRSA_CA_EXPIRE 3650
set_var EASYRSA_CERT_EXPIRE 3650
        `;
        
        f, err7 := os.OpenFile(filepath.Join(homepath, ".httphere", "easyrsa", "pki", "vars"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
        if err7 != nil {
            log.Fatal(err7)
        }
        defer f.Close()
    	
    	f.WriteString(var_data)
    	f.Close()
    	
    	
    	
    	
    	cmd2 := exec.Command("bash", "-c", "easyrsa build-ca nopass")
        cmd2.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
        out4, _ := cmd2.Output()
    	
    	fmt.Println(yellow("--------------------------------------------------------------------------------------------------"))
    	fmt.Printf(" %s\n", out4)
    	
    	
    	
    	cmd3 := exec.Command("bash", "-c", "easyrsa --req-cn=ChangeMe build-client-full server1 nopass")
        cmd3.Dir = filepath.Join(homepath, ".httphere", "easyrsa")
        out5, _ := cmd3.Output()
    	
    	fmt.Println(yellow("--------------------------------------------------------------------------------------------------"))
    	fmt.Printf(" %s\n", out5)
	    
	    
	    
	    fmt.Println(yellow("--------------------------------------------------------------------------------------------------"))
	    fmt.Println("")
	    fmt.Println("     Crt: " + yellow(crt_filename))
	    fmt.Println("     Key: " + yellow(key_filename))
	    fmt.Println("")
	}
	
	
	
	
	
	fmt.Println("")
	if *arg_tls {
	    fmt.Println( yellow("  TLS Server port " + strconv.Itoa(*arg_port)) )
	}else{
	    fmt.Println("  Server port " + cian(strconv.Itoa(*arg_port)))
    }

	fmt.Println("")
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return
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
				//fmt.Printf("%v : %s [%v/%v]\n", i.Name, v, v.IP, v.Mask)
				//fmt.Printf("%v \n", v.IP)
				if v.IP.To4() != nil {
					//fmt.Println( "yes, ipv4" )
					
					if *arg_tls {
					    fmt.Println("     https://" + v.IP.String() + ":" + cian(strconv.Itoa(*arg_port)))
					}else{
					    fmt.Println("     http://" + v.IP.String() + ":" + cian(strconv.Itoa(*arg_port)))
				    }
				}
			}

		}
	}
	fmt.Println("")
	
	
	
	
	
	
	
	fmt.Println("  Serve folder: " + cian(arg_fold))
	fmt.Println("")
	fmt.Println(cian("  [ Control + C ] ") + "Break Server")
	fmt.Println("")
	
	
	if *arg_tls {
        
        
        
        log.Fatal(   app.ListenTLS(":" + strconv.Itoa(*arg_port), crt_filename, key_filename)    )
        
    } else {
        
	    log.Fatal(app.Listen(":" + strconv.Itoa(*arg_port)))
    }
}






