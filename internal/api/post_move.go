package api

import (
	//"fmt"
	"net/url"
	//"os"
	//"path"
	"errors"
	"golang.org/x/sync/errgroup"
	"path/filepath"
	"strings"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/model2"
	"github.com/western/http-here/v2/internal/safefs"
	"github.com/western/http-here/v2/internal/util"

	"github.com/gofiber/fiber/v3"
)

func PostMove(c fiber.Ctx) error {

	// -------------------------------------------------------------------------------------------------------------------------

	from_path := c.FormValue("from_path")

	from_path, err := url.QueryUnescape(from_path)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostMove", err.Error())

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  err.Error(),
		})
	}

	from_path = util.CleanDirtyPath(from_path)

	if len(from_path) == 0 {

		model2.EventLogAdd(c, 500, "PostMove", "from_path is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "from_path is empty",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// check from_path

	sourceRoot := conf.ArgFoldRootFS

	if from_path != "/" {

		fileInfo, err := sourceRoot.Stat(strings.TrimPrefix(from_path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostMove", "sourceRoot.Stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostMove", "It is not directory. "+from_path)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}

		// REOPEN sourceRoot for use below
		sourceRoot, err = safefs.NewSafeFS(filepath.Join(conf.ArgFold, from_path))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostMove", "Open storage targetRoot err: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

	}

	// -------------------------------------------------------------------------------------------------------------------------

	to_path := c.FormValue("to_path")

	to_path, err = url.QueryUnescape(to_path)
	if err != nil {

		model2.EventLogAdd(c, 500, "PostMove", err.Error())

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  err.Error(),
		})
	}

	to_path = util.CleanDirtyPath(to_path)

	if len(to_path) == 0 {

		model2.EventLogAdd(c, 500, "PostMove", "to_path is empty")

		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "to_path is empty",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------
	// check to_path

	targetRoot := conf.ArgFoldRootFS

	if to_path != "/" {

		fileInfo, err := targetRoot.Stat(strings.TrimPrefix(to_path, "/"))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostMove", "targetRoot.Stat err "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

		if !fileInfo.IsDir() {

			model2.EventLogAdd(c, 500, "PostMove", "It is not directory. "+to_path)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "It is not directory",
			})
		}

		// REOPEN targetRoot for use below
		targetRoot, err = safefs.NewSafeFS(filepath.Join(conf.ArgFold, to_path))
		if err != nil {

			model2.EventLogAdd(c, 500, "PostMove", "Open storage targetRoot err: "+err.Error())

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "stat err",
			})
		}

	}

	// -------------------------------------------------------------------------------------------------------------------------

	if from_path == to_path {

		model2.EventLogAdd(c, 500, "PostMove", "Source and target paths '"+from_path+"' are identical")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Source and target paths '" + from_path + "' are identical",
		})
	}

	// -------------------------------------------------------------------------------------------------------------------------

	form, _ := c.MultipartForm()
	names := form.Value["name"]

	if len(names) == 0 {

		model2.EventLogAdd(c, 500, "PostMove", "form is empty")

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "form is empty",
		})
	}

	//var err_list []string

	errGroup1 := new(errgroup.Group)
	errGroup1.SetLimit(10)

	for _, name := range names {

		errGroup1.Go(func() error { return entityPostMove(c, sourceRoot, targetRoot, name) })
	}

	err = errGroup1.Wait()
	if err != nil {

		RemoveCacheDir(filepath.Join(conf.ArgFold, from_path))
		RemoveCacheDir(filepath.Join(conf.ArgFold, to_path))

		model2.EventLogAdd(c, 500, "PostMove", "Move error: "+err.Error())

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": 500,
			"msg":  "Move error",
		})
	}

	RemoveCacheDir(filepath.Join(conf.ArgFold, from_path))
	RemoveCacheDir(filepath.Join(conf.ArgFold, to_path))

	/*
		if len(err_list) > 0 {

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  err_list[0],
			})
		}
	*/

	return c.JSON(fiber.Map{
		"code": 200,
	})

}

func entityPostMove(c fiber.Ctx, sourceRoot *safefs.RootFS, targetRoot *safefs.RootFS, name_ string) error {

	//fmt.Println("entityPostMove start ", name_)

	name := strings.ReplaceAll(name_, "/", "")
	name = util.CleanDirtyPath(name)

	if len(name) == 0 {
		model2.EventLogAdd(c, 500, "PostMove", "name is empty")
		//err_list = append(err_list, "name is empty")
		//continue
		return errors.New("name is empty")
	}

	src_file_path := filepath.Join(sourceRoot.Name(), name)
	target_file_path := filepath.Join(targetRoot.Name(), name)

	//src_stat, err := os.Stat(src_file_path)
	src_stat, err := sourceRoot.Stat(name)
	if err != nil {
		model2.EventLogAdd(c, 500, "PostMove", "Source '"+src_file_path+"' not exists")
		//err_list = append(err_list, "Source '"+src_file_path+"' not exists")
		//continue
		return errors.New("Source '" + src_file_path + "' not exists")
	}

	//_, err = os.Stat(target_file_path)
	_, err = targetRoot.Stat(name)
	if err == nil {

		model2.EventLogAdd(c, 200, "PostMove", "Target '"+target_file_path+"' is exists. It will be rewrite.")

		//if err := os.RemoveAll(target_file_path); err != nil {
		if err := targetRoot.RemoveAll(name); err != nil {
			model2.EventLogAdd(c, 500, "PostMove", "'"+target_file_path+"' err "+err.Error())
			//err_list = append(err_list, "'"+target_file_path+"' err "+err.Error())
			//continue
			return errors.New("'" + target_file_path + "' err " + err.Error())
		}
	}

	if src_stat.IsDir() {

		err := util.CopyDir(src_file_path, target_file_path)

		if err != nil {
			model2.EventLogAdd(c, 500, "PostMove", "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
			//err_list = append(err_list, "CopyDir '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
			//continue
			return errors.New("CopyDir '" + src_file_path + "' to '" + target_file_path + "' err " + err.Error())
		}

		// remove source folder
		//if err := os.RemoveAll(src_file_path); err != nil {
		if err := sourceRoot.RemoveAll(name); err != nil {
			model2.EventLogAdd(c, 500, "PostMove", "rm folder '"+src_file_path+"' err "+err.Error())
			//err_list = append(err_list, "'"+src_file_path+"' err "+err.Error())
			//continue
			return errors.New("RemoveAll '" + src_file_path + "' err " + err.Error())
		}

		model2.EventLogAdd(c, 200, "PostMove", "Move dir '"+src_file_path+"' to "+target_file_path)

	} else {

		err := util.MoveFile(src_file_path, target_file_path)

		if err != nil {
			model2.EventLogAdd(c, 500, "PostMove", "MoveFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
			//err_list = append(err_list, "MoveFile '"+src_file_path+"' to '"+target_file_path+"' err "+err.Error())
			//continue
			return errors.New("MoveFile '" + src_file_path + "' to '" + target_file_path + "' err " + err.Error())
		}

		model2.EventLogAdd(c, 200, "PostMove", "Move file '"+src_file_path+"' to "+target_file_path)
	}

	return nil

}
