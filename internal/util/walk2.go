package util

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/charlievieth/fastwalk"
	//_ "github.com/western/http-here/internal/fastwalk2"
)

func WalkAndTreeBuild2(path_ string, deep int) []TreeRow {

	var folder_list []string

	conf := fastwalk.Config{
		Follow: true,
		Sort:   fastwalk.SortNone,
		//Sort: fastwalk.SortLexical,
	}

	walkFn := func(result_path string, d fs.DirEntry, err error) error {

		if err != nil {
			//fmt.Fprintf(os.Stderr, "%s: %v\n", result_path, err)
			// return error - stop iteration
			return nil
		}

		if d.IsDir() {

			c_path := strings.Replace(result_path, path_, "", 1)

			folder_list = append(folder_list, c_path)
		}

		// return error - stop iteration
		return nil
	}

	if err := fastwalk.Walk(&conf, path_, walkFn); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path_, err)
		os.Exit(1)
	}

	sort.Strings(folder_list)

	var node_list []TreeRow
	max_depth := 1

	for _, val := range folder_list {

		if len(val) == 0 {
			continue
		}

		folder_name := GetFileName(val)

		if len(folder_name) == 0 {
			folder_name = val
		}

		pt := strings.Split(val, "/")
		pt = pt[1:]
		if len(pt) > max_depth {
			max_depth = len(pt)
		}

		node_list = append(node_list, TreeRow{

			Text: folder_name,
			Path: val,

			Expanded: false,

			Nodes: []TreeRow{},

			PathArr:  pt,
			MaxDepth: len(pt),
		})

	}

	var node_list2 []TreeRow

	/*
	   for i := 0; i < len(node_list); i++ {
	       node_list2[i] = node_list[i]
	   }
	*/

	for _, node := range node_list {

		//PrintPrettify("node", node)

		node_list2 = findAndInsert(node_list2, node)
	}

	//PrintPrettify("node_list2", node_list2)

	return node_list2

}

func findAndInsert(node_list []TreeRow, nd TreeRow) []TreeRow {

	// at first - what the insert node
	// - root element
	// - non-root element

	if nd.MaxDepth == 1 {
		// it is root
		// looking similar

		is_found := false

		for _, node := range node_list {
			if node.Text == nd.Text && node.Path == nd.Path {
				is_found = true
			}
		}

		if !is_found {
			node_list = append(node_list, nd)
		}
	}

	fn_walk := func(nd_list []TreeRow) {}

	fn_walk = func(nd_list []TreeRow) {

		for indx1, node := range nd_list {

			if len(node.PathArr) == len(nd.PathArr)-1 {

				path_arr := nd.PathArr
				path_arr = path_arr[:len(path_arr)-1]

				is_match := true

				for indx2, pt := range path_arr {
					if pt != node.PathArr[indx2] {
						is_match = false
					}
				}

				if is_match {
					//fmt.Println("found! for fn_walk ", nd)
					//node.Nodes = append(node.Nodes, nd)
					nd_list[indx1].Nodes = append(node.Nodes, nd)
				}
			}

			if len(node.Nodes) > 0 {
				fn_walk(node.Nodes)
			}

		}
	}

	if nd.MaxDepth > 1 {
		//fmt.Println("looking for fn_walk ", nd)
		fn_walk(node_list)
	}

	return node_list
}
