package util

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// openssl aes-256-cbc -a -salt -in file.txt -out file.txt.cr -pass pass:123
func EncryptFile(path_, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path_) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	from_file := filepath.Base(path_)
	to_file := filepath.Base(path_) + ".cr"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -a -salt -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path_)
	//out, _ := cmd.Output()
	//_ = out

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		//panic(err)
		return false, err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		//fmt.Println("openssl:", scanner.Text())
	}

	//fmt.Println("--------------------------------------------------------------------------------------------------")
	//fmt.Println("out=", out)

	if err = os.Remove(path_); err != nil {
		return false, err
	}

	if err = os.Rename(path.Join(filepath.Dir(path_), to_file), path_); err != nil {
		return false, err
	}

	return true, nil
}

// openssl aes-256-cbc -d -a -in file.txt.cr -out file.txt.new -pass pass:123
func DecryptFile(path_, pass string) (bool, error) {

	_, err := exec.Command("bash", "-c", "openssl --help").Output()
	if err != nil {
		return false, err
	}

	if len(path_) == 0 {
		return false, errors.New("Path of file is empty")
	}

	if len(pass) == 0 {
		return false, errors.New("Pass is empty")
	}

	from_file := filepath.Base(path_)
	to_file := filepath.Base(path_)
	to_file = strings.Replace(to_file, ".cr", "", 1)
	to_file += ".decrypt"

	cmd := exec.Command("bash", "-c", "openssl aes-256-cbc -d -a  -in "+from_file+" -out "+to_file+" -pass pass:"+pass)
	cmd.Dir = filepath.Dir(path_)
	//out, _ := cmd.Output()
	//_ = out

	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		//panic(err)
		return false, err
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		//fmt.Println("openssl:", scanner.Text())
	}

	//fmt.Println("--------------------------------------------------------------------------------------------------")
	//fmt.Println("out=", out)

	if err = os.Remove(path_); err != nil {
		return false, err
	}

	if err = os.Rename(path.Join(filepath.Dir(path_), to_file), path_); err != nil {
		return false, err
	}

	return true, nil
}
