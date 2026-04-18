package conf

import (
	"os"
	"path"
)

const Version = "v2.3.0"

const FilesCountMax = 200
const FieldSizeMax = 14 * 1024 * 1024 * 1024
const FieldSizeMaxHuman = "14 Gb"

var ConfigRoot string
var ArgFold string

func init() {

	homepath, err := os.UserHomeDir()
	if err != nil {

		panic("User homepath detect error: " + err.Error())
	}
	ConfigRoot = path.Join(homepath, ".httphere")

}
