package controller

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"time"

	"github.com/fatih/color"
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

func LogPrefix(c *fiber.Ctx, status string, addition string) {

	green := color.New(color.FgGreen).SprintFunc()
	//magenta := color.New(color.FgMagenta).SprintFunc()
	cian := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	pref := ""

	pref += "[" + green(time.Now().Format("2006-01-02 15:04:05")) + "] "

	pref += "[" + green(c.IP()) + "] "

	_user := ""
	if c.Locals("username") != nil {
		_user = green(c.Locals("username").(string))
	}

	if len(_user) > 0 {
		pref += "[" + _user + "] "
	} else {
		pref += "[] "
	}

	if status != "200" {
		pref += "[" + yellow(status) + "] "
	} else {
		pref += "[" + status + "] "
	}

	pref += cian(addition)

	fmt.Println(pref)

	// [2024-10-29 18:22:46] [192.168.0.101] [loginT] [200] Dir /tmp/fold1/fold3
	//fmt.Printf("[%s] [%s] [%s] [%s] %s\n", magenta("warning"), red("error"))

}

func prettyByteSize(b int64) string {
	bf := float64(b)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
