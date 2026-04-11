//go:build windows
// +build windows

package app

import (
	"github.com/western/http-here/v2/internal/util"
)

func MakeLibreofficeThumbnail(filepath_tmp, arg_fold_path string) {

	util.RunAnyCommandUnderWin(`"C:/Program Files/LibreOffice/program/soffice.exe" --headless --norestore --nologo --convert-to png --outdir "` + filepath_tmp + `" "` + arg_fold_path + `"`)

}
