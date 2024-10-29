

package controller

import (
	"regexp"
	"time"
	"fmt"
	
	"github.com/gofiber/fiber/v2"
)



func CleanDirtyPath(p string) string {
    
    re := regexp.MustCompile("/+")
    p = re.ReplaceAllLiteralString(p, "/")
    
    
    re2 := regexp.MustCompile("\\.{2,}")
    p = re2.ReplaceAllLiteralString(p, ".")
    
    //p = filepath.Clean(p)
    
    return p
}


func LogPrefix(c *fiber.Ctx, status string, addition string)  {
    
    pref := ""
    
    pref += "[" + time.Now().UTC().Format("2006-01-02 15:04:05") + "] "
    
    pref += "[" + c.IP() + "] "
    
    
    
    _user := ""
    if c.Locals("username") != nil {
        _user = c.Locals("username").(string)
    }
    
    if len(_user) > 0 {
        pref += "[" + _user + "] "
    }
    
    
    
    pref += "[" + status + "] "
    
    
    
    pref += addition
    
    
    
    fmt.Println(pref)
    
    
    //return pref
}

