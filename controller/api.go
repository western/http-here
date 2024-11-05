package controller

import (
	_ "errors"
	_ "fmt"

	"path/filepath"
	"regexp"
	"strings"
	"strconv"
	"archive/zip"

	"github.com/gofiber/fiber/v2"

	"os"

	"io"
	"log"
	"net/url"
	"time"
)

func PostUpload(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	form, _ := c.MultipartForm()
	files := form.File["fileBlob"]

	for _, file := range files {
		fileExt := filepath.Ext(file.Filename)
		fileExt = strings.ToLower(fileExt)

		originalFileName := strings.TrimSuffix(
			filepath.Base(file.Filename),
			filepath.Ext(file.Filename),
		)

		originalFileName = strings.ReplaceAll(originalFileName, "/", "")
		re := regexp.MustCompile("\\s+")
		originalFileName = re.ReplaceAllLiteralString(originalFileName, "-")

		filename := originalFileName + fileExt
		filename = CleanDirtyPath(filename)

		if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, filename)); err == nil {

			if !fileInfo.IsDir() {
				LogPrefix(c, "200", "'"+filepath.Join(arg_fold, u_path, filename)+"' already exists. It will be rewrite.")
			}
		}

		out, err := os.Create(filepath.Join(arg_fold, u_path, filename))
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error create "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}
		defer out.Close()

		LogPrefix(c, "200", "Save '"+filepath.Join(arg_fold, u_path, filename)+"'")

		readerFile, _ := file.Open()
		_, err = io.Copy(out, readerFile)
		if err != nil {

			log.Println(err)
			LogPrefix(c, "500", "Error copy "+filepath.Join(arg_fold, u_path, filename))
			return c.JSON(fiber.Map{
				"code": 500,
			}, "application/json")
		}
	}

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

func PostFolder(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)

	name := c.FormValue("name")

	name = strings.ReplaceAll(name, "/", "")
	re := regexp.MustCompile("\\s+")
	name = re.ReplaceAllLiteralString(name, " ")

	name = CleanDirtyPath(name)

	if len(name) == 0 {
		LogPrefix(c, "500", "name is empty")
		return c.JSON(fiber.Map{
			"code": 500,
			"msg":  "name is empty",
		}, "application/json")
	}

	if fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name)); err == nil {

		if fileInfo.IsDir() {
			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' already exists")
			return c.JSON(fiber.Map{
				"code": 500,
				"msg":  filepath.Join(u_path, name) + " already exists",
			}, "application/json")
		}
	}

	if err := os.Mkdir(filepath.Join(arg_fold, u_path, name), os.ModePerm); err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error mkdir "+filepath.Join(arg_fold, u_path, name))
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	LogPrefix(c, "200", "Mkdir '"+filepath.Join(arg_fold, u_path, name)+"'")

	return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
}

func PostDelete(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)
    
    
    
    
    
    for i:=1 ; i<50 ; i++ {
        
        key := "name["+strconv.Itoa(i)+"]"
        val := c.FormValue(key)
        
        if len(val) > 0 {
            //fmt.Println("val="+val)
            
            
            name := strings.ReplaceAll(val, "/", "")
        	//re := regexp.MustCompile("\\s+")
        	//name = re.ReplaceAllLiteralString(name, " ")

        	name = CleanDirtyPath(name)

        	if len(name) == 0 {
        		LogPrefix(c, "500", "name is empty")
        		continue
        		/*
        		return c.JSON(fiber.Map{
        			"code": 500,
        			"msg":  "name is empty",
        		}, "application/json")
        		*/
        	}

        	fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

        	if err != nil {
        		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
        		continue
        		/*
        		return c.JSON(fiber.Map{
        			"code": 500,
        			"msg":  filepath.Join(u_path, name) + " not exists",
        		}, "application/json")
        		*/
        	}

        	if fileInfo.IsDir() {

        		// remove fold and all inside data

        		if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

        			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
        			continue
        			/*
        			return c.JSON(fiber.Map{
        				"code": 500,
        				"msg":  filepath.Join(u_path, name) + " err",
        			}, "application/json")
        			*/
        		}

        		LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")

        	} else {

        		// remove one file

        		if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

        			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
        			continue
        			/*
        			return c.JSON(fiber.Map{
        				"code": 500,
        				"msg":  filepath.Join(u_path, name) + " err",
        			}, "application/json")
        			*/
        		}

        		LogPrefix(c, "200", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")

        	}
            
            
        }
    }
    
    
    
    
    return c.JSON(fiber.Map{
		"code": 200,
	}, "application/json")
    
    
    
    
}




func PostZip(c *fiber.Ctx) error {

	arg_fold := ""
	arg_fold = c.Locals("arg_fold").(string)
	
	
	
	
	
	homepath, err2 := os.UserHomeDir()
	if err2 != nil {
		log.Fatal(err2)
	}
	//fmt.Println( homepath )
	
	
	if _, err3 := os.Stat(filepath.Join(homepath, ".httphere", "temp")); err3 != nil {

		if err4 := os.MkdirAll(filepath.Join(homepath, ".httphere", "temp"), os.ModePerm); err4 != nil {
			log.Fatal(err4)
		}
	}
	
	
	

	referer := c.Get("Referer")

	// already decoded
	u, err := url.Parse(referer)
	if err != nil {

		log.Println(err)
		LogPrefix(c, "500", "Error url parse "+referer)
		return c.JSON(fiber.Map{
			"code": 500,
		}, "application/json")
	}

	u_path := CleanDirtyPath(u.Path)
    
    
    
    
    
    
    
    
    
    
    
    //archive_name := "archive-"+time.Now().Format("2006-01-02T15:04:05")+".zip"
    archive_name := "archive-"+time.Now().Format("20060102-150405")+".zip"
    
    
    
    //archive, err := os.Create(filepath.Join(arg_fold, u_path, "archive.zip"))
    archive, err := os.Create(filepath.Join(homepath, ".httphere", "temp", archive_name))
    
    if err != nil {
        panic(err)
    }
    defer archive.Close()
    
    
    
    zipWriter:=zip.NewWriter(archive)
    
    
    
    
    
    for i:=1 ; i<50 ; i++ {
        
        key := "name["+strconv.Itoa(i)+"]"
        val := c.FormValue(key)
        
        if len(val) > 0 {
            //fmt.Println("val="+val)
            
            
            name := strings.ReplaceAll(val, "/", "")
        	//re := regexp.MustCompile("\\s+")
        	//name = re.ReplaceAllLiteralString(name, " ")

        	name = CleanDirtyPath(name)

        	if len(name) == 0 {
        		LogPrefix(c, "500", "name is empty")
        		continue
        		/*
        		return c.JSON(fiber.Map{
        			"code": 500,
        			"msg":  "name is empty",
        		}, "application/json")
        		*/
        	}

        	fileInfo, err := os.Stat(filepath.Join(arg_fold, u_path, name))

        	if err != nil {
        		LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' not exists")
        		continue
        		/*
        		return c.JSON(fiber.Map{
        			"code": 500,
        			"msg":  filepath.Join(u_path, name) + " not exists",
        		}, "application/json")
        		*/
        	}

        	if fileInfo.IsDir() {

        		/*

        		if err := os.RemoveAll(filepath.Join(arg_fold, u_path, name)); err != nil {

        			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
        			continue
        			
        		}

        		LogPrefix(c, "200", "Remove fold '"+filepath.Join(arg_fold, u_path, name)+"'")
        		*/
        		
        		addFiles(zipWriter, filepath.Join(arg_fold, u_path, name),  name)

        	} else {
                
                /*
                zipWriter := zip.NewWriter(archive)
                
                
                f1, err := os.Open(filepath.Join(arg_fold, u_path, name))
                if err != nil {
                    panic(err)
                }
                defer f1.Close()
                
                
                w1, err := zipWriter.Create( name)
                if err != nil {
                    panic(err)
                }
                if _, err := io.Copy(w1, f1); err != nil {
                    panic(err)
                }
                
                zipWriter.Close()
                */
                
                
                
                
                
                
                
                
                
                
                f1, err:=os.Open(filepath.Join(arg_fold, u_path, name))
                if err!=nil{
                    panic(err)
                }
                defer f1.Close()

                
                w1,err:=zipWriter.Create( name)
                if err!=nil{
                    panic(err)
                }
                if _,err:=io.Copy(w1,f1); err != nil{
                    panic(err)
                }
                
                
                
                
                
                
        		
                /*
        		if err := os.Remove(filepath.Join(arg_fold, u_path, name)); err != nil {

        			LogPrefix(c, "500", "'"+filepath.Join(arg_fold, u_path, name)+"' err")
        			continue
        			
        		}

        		LogPrefix(c, "200", "Remove '"+filepath.Join(arg_fold, u_path, name)+"'")
        		*/

        	}
            
            
        }
    }
    
    zipWriter.Close()
    archive.Close()
    
    
    //return c.SendFile(filepath.Join(arg_fold, u_path, "archive.zip"), false)
    //return c.Download(filepath.Join(arg_fold, u_path, "archive.zip"), "archive.zip");
    
    LogPrefix(c, "200", "Temp file create "+filepath.Join(homepath, ".httphere", "temp", archive_name))
    
    return c.JSON(fiber.Map{
		"code": 200,
		"file": filepath.Join( "/__temp/", archive_name),
	}, "application/json")
    
    
    
    
}




