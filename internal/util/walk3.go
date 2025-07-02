package util

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"

	//"github.com/wneessen/go-fileperm"
	//_ "github.com/western/http-here/internal/fastwalk2"
)

func WalkAndTreeBuild3(path_, prev_path string, deep int) []TreeRow {

	var node_list []TreeRow

	//fmt.Println("WalkAndTreeBuild3 ", path_, " ", prev_path)

	var wg sync.WaitGroup
	t := make(chan TreeRow, 10)
	cnt_i := 0
	var cnt_s_list []string

	if _, err := os.Stat(path_); err != nil {
		fmt.Println("os.Stat err:", path_, err)
		return node_list
	}

	files, err := os.ReadDir(path_)
	if err != nil {
		fmt.Println("os.ReadDir err:", err)
		return node_list
	}

	if len(files) == 0 {
		return node_list
	}

	if deep > 5 {
		return node_list
	}

	for _, file := range files {

		//full_path := path.Join(path_, file.Name())

		/*
			        var fileInfo os.FileInfo

			        if fileInfo, err = os.Stat(full_path); err != nil {
					    fmt.Println("os.Stat 2 err:", full_path,  err)
					    continue
				    }
		*/

/*
		up, err := fileperm.New(full_path)
		if err != nil {
			panic(err)
		}

		if !up.UserReadable() && file.IsDir() {
			fmt.Println("file is not UserReadable ", file.Name())
			continue
		}

		if !up.UserReadExecutable() && file.IsDir() {
			fmt.Println("file is not UserReadExecutable ", file.Name())
			continue
		}
*/

		// fileInfo.Mode().Type() == fs.ModeSocket

		/*
		   if fileInfo.Mode().Type() == os.ModeSocket {
		       fmt.Println("file is socket ", file.Name())
		       continue
		   }
		*/

		/*
		   if fileInfo.Mode().Type() == os.ModeIrregular {
		       fmt.Println("file/dir is irregular ", file.Name())
		       continue
		   }
		*/

		//fmt.Println("file/dir is=", file.Name(), " mode=", fileInfo.Mode())

		//Check 'others' permission
		/*
		   m := fileInfo.Mode()
		   if m&(1<<2) != 0 {
		       //other users have read permission
		       //fmt.Println("other users have read permission ", file.Name())
		   } else {
		       //other users don't have read permission
		       fmt.Println("other users don't have read permission ", file.Name())
		       continue
		   }
		*/

		// fileInfo.Mode().IsRegular()

		/*
			    if strings.Contains(file.Name(), "systemd-private") {

			        fmt.Println("systemd-private string match MODE ", fileInfo.Mode().Type())

			        fmt.Println("systemd-private string match ", file.Name())
				    continue
			    }


			    if strings.Contains(file.Name(), "plasma-csd") {
			        fmt.Println("plasma-csd string match ", file.Name())
				    continue
			    }


			    if strings.Contains(file.Name(), "ssh-XXXX") {
			        fmt.Println("ssh-XXXX string match ", file.Name())
				    continue
			    }
		*/

		if file.IsDir() {

			fmt.Println("wg.Add ", file.Name())

			wg.Add(1)
			cnt_i += 1
			cnt_s_list = append(cnt_s_list, file.Name())

			go func() {
				defer wg.Done()

				nodes := WalkAndTreeBuild3(path.Join(path_, file.Name()), path.Join(prev_path, file.Name()), deep+1)
				//nodes := node_list

				fmt.Println("go_func 1 ", file.Name())

				t <- TreeRow{

					Text: file.Name(),
					Path: path.Join(prev_path, file.Name()),

					Nodes: nodes,
				}

				fmt.Println("go_func 2 ", file.Name())

				//time.Sleep(5 * time.Second)
				//wg.Done()

				//fmt.Println("go_func 3 ", file.Name())

			}()

			/*
				nodes := WalkAndTreeBuild3(path.Join(path_, file.Name()), path.Join(prev_path, file.Name()), deep+1)

				node_list = append(node_list, TreeRow{

					Text: file.Name(),
					Path: path.Join(prev_path, file.Name()),

					Nodes: nodes,
				})
			*/
		}
	}

	//defer close(t)
	//close(t)

	if cnt_i == 0 {
		//fmt.Println("pre wg wait cnt_i=0 return []")
		fmt.Println("WalkAndTreeBuild3 RETURN[] FOR ", path_, " ", prev_path)

		return node_list
	}

	fmt.Println("pre wg wait ", cnt_i)
	fmt.Println("pre wg wait LIST= ", cnt_s_list)

	/*
			go func() {
		        time.Sleep(2 * time.Second)
		        panic("PANIC WITH WalkAndTreeBuild3 "+path_+ " "+ prev_path)
		    }()
	*/

	wg.Wait()

	//close(t)

	fmt.Println("after wg wait ", cnt_i)

	if cnt_i == 0 {
		//fmt.Println("node_list=", node_list)
		fmt.Println("WalkAndTreeBuild3 RETURN[] FOR ", path_, " ", prev_path)

		return node_list
	}

	for resp := range t {

		fmt.Println("resp=", resp)

		node_list = append(node_list, resp)

		//close(t)
		//fmt.Println(IsClosed(t))
		//break

		//fmt.Println("i=", i)
		cnt_i--

		if cnt_i == 0 {
			fmt.Println("WalkAndTreeBuild3 BREAK FOR ", path_, " ", prev_path)
			break
		}

	}

	/*
	   for {
	       resp, more := <-t

	       fmt.Println("resp=", resp)
	       fmt.Println("more=", more)

	       if more {
	           fmt.Println("received job", resp)
	       } else {
	           fmt.Println("received all t")
	           //done <- true
	           return node_list
	       }
	   }
	*/

	//fmt.Println("node_list=", node_list)

	return node_list

}

func IsClosed(ch <-chan TreeRow) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func DefaultNumWorkers() int {
	numCPU := runtime.GOMAXPROCS(-1)
	if numCPU < 4 {
		return 4
	}
	// Darwin IO performance on APFS can slow with increased parallelism.
	// Depending on CPU, stat(2) performance is best around 4-10 workers
	// and file IO is best around 4 workers. More workers only benefit CPU
	// intensive tasks.
	//
	// NB(Charlie): As of macOS 15, the parallel performance of readdir_r(3)
	// and stat(2) calls has improved and is now generally the number of
	// performance cores (on ARM Macs).
	//
	// TODO: Consider using the value of sysctl("hw.perflevel0.physicalcpu").
	// TODO: Find someone with a Mac Studio to test higher core counts.
	if runtime.GOOS == "darwin" {
		switch {
		case numCPU <= 8:
			return 4
		case numCPU <= 10:
			return 6
		default: // numCPU > 10
			return 10
		}
	}
	if numCPU > 32 {
		return 32
	}
	return numCPU
}
