package util

import (
	"os"
	"path"
)

type TreeRow struct {
	Text string `json:"text"`
	Path string `json:"path"`

	Expanded bool `json:"expanded"`

	Nodes []TreeRow `json:"nodes"`

	PathArr  []string `json:"path_arr"`
	MaxDepth int      `json:"max_depth"`
}

func WalkAndTreeBuild(path_ string, prev_path string, deep int) []TreeRow {

	//fmt.Println("WalkAndTreeBuild ", path)

	var node_list []TreeRow

	files, err := os.ReadDir(path_)
	if err != nil {
		//fmt.Println(err)
		return node_list
	}

	if deep > 5 {
		return node_list
	}

	for _, file := range files {

		if file.IsDir() {

			/*
			   fmt.Println("first: ", path.Join(path_, file.Name()))
			   fmt.Println("secon: ", path.Join(prev_path, file.Name()))
			   fmt.Println()
			*/

			nodes := WalkAndTreeBuild(path.Join(path_, file.Name()), path.Join(prev_path, file.Name()), deep+1)

			node_list = append(node_list, TreeRow{

				Text: file.Name(),
				Path: path.Join(prev_path, file.Name()),

				Nodes: nodes,
			})

		}
	}

	//fmt.Println("node_list=", node_list)

	return node_list
}
