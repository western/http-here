package conf

import (
	//"fmt"
	"os"
	"path"

	"github.com/gofiber/fiber/v3"
)

const Version = "v2.4.1"

const FilesCountMax = 200
const FieldSizeMax = 14 * 1024 * 1024 * 1024
const FieldSizeMaxHuman = "14 Gb"

var ConfigRoot string
var ArgFold string

var MasterPID int

func init() {

	homepath, err := os.UserHomeDir()
	if err != nil {

		panic("User homepath detect error: " + err.Error())
	}
	ConfigRoot = path.Join(homepath, ".httphere")

	if !fiber.IsChild() {
		MasterPID = os.Getppid()
		//fmt.Println("conf.MasterPID=", MasterPID)
	}

}
