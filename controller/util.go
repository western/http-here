package controller

import (
	"fmt"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/fatih/color"
)

func CleanDirtyPath(p string) string {

	re := regexp.MustCompile("/+")
	p = re.ReplaceAllLiteralString(p, "/")

	re2 := regexp.MustCompile("\\.{2,}")
	p = re2.ReplaceAllLiteralString(p, ".")

	//p = filepath.Clean(p)

	return p
}

func LogPrefix(c *fiber.Ctx, status string, addition string) {
    
    green := color.New(color.FgGreen).SprintFunc()
    //magenta := color.New(color.FgMagenta).SprintFunc()
    cian := color.New(color.FgCyan).SprintFunc()
    yellow := color.New(color.FgYellow).SprintFunc()
    
    
    
	pref := ""

	pref += "[" + green(time.Now().UTC().Format("2006-01-02 15:04:05")) + "] "

	pref += "[" + green(c.IP()) + "] "

	_user := ""
	if c.Locals("username") != nil {
		_user = green(c.Locals("username").(string))
	}

	if len(_user) > 0 {
		pref += "[" + _user + "] "
	}else{
	    pref += "[] "
	}
	
    if status != "200" {
	    pref += "[" + yellow(status) + "] "
	}else{
	    pref += "[" + status + "] "
	}

	pref += cian(addition)

	fmt.Println(pref)
    
    
    
    
    
    
    // [2024-10-29 18:22:46] [192.168.0.101] [loginT] [200] Dir /tmp/fold1/fold3
    //fmt.Printf("[%s] [%s] [%s] [%s] %s\n", magenta("warning"), red("error"))
    
    
	
}
