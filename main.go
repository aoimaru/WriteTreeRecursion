package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

type Tree struct {
	// 実際はnodeオブジェクト
	Path     string
	Children []*Tree
}

type FileStatus struct {
	// TODO: Nameの型をどうするか, 文字列の長さを要素として持ち, 動的にメモリを確保するのか. それとも決め打ちするのか
	Name string
	Hash [20]byte
	Size uint32
	Mode uint32
}

func (fs FileStatus) AsByte() []byte {
	buffer := make([]byte, 0)
	if fs.Mode == 2147484141 {
		mode := []byte("40000")
		buffer = append(buffer, mode...)
	}
	buffer = append(buffer, []byte("100644")...)
	buffer = append(buffer, []byte(fs.Name)...)
	buffer = append(buffer, 0)
	buffer = append(buffer, fs.Hash[0:20]...)

	return buffer
}

func Press(buffer []byte) []uint8 {
	var Pressed bytes.Buffer
	zWriter := zlib.NewWriter(&Pressed)
	zWriter.Write(buffer)
	zWriter.Close()

	return Pressed.Bytes()
}

func WriteTreeObject(root FileStatus, buffer []byte) {
	current_dir, _ := os.Getwd()
	file_path := current_dir + "/.bakibaki/objects/" + string(root.Hash[:2])
	write_object, err := os.Create(file_path)
	if err != nil {
		return
	}
	defer write_object.Close()
	count, err := write_object.Write(Press(buffer))
	if err != nil {
		return
	}
	fmt.Printf("write %d bytes\n", count)

}

func Walk(tree *Tree, log string) {
	if len((*tree).Children) <= 0 {
		return
	}

	fmt.Println()
	root_file_status, err := GetFileStatus((*tree).Path)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", root_file_status)

	buffers := make([]byte, 0)
	buffers = append(buffers, []byte("tree")...)
	buffers = append(buffers, []byte("42342")...)

	for _, child_tree := range (*tree).Children {
		// fmt.Println((*child_tree).Path)
		file_stauts, err := GetFileStatus((*child_tree).Path)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%+v\n", file_stauts)
		buffers = append(buffers, file_stauts.AsByte()...)
	}

	WriteTreeObject(root_file_status, buffers)

	for _, child_tree := range (*tree).Children {
		Walk(child_tree, log+" "+(*tree).Path)
	}
}

func Compress() (hash string) {

}

func Walking(tree *Tree) (hash string) {
	for _, child_tree := range (*tree).Children {
		child_hash := Walking(child_tree)
		print(child_hash)
	}
}

func RelPath2AbsPath(rel_path string) string {
	current_dir, _ := os.Getwd()
	abs_path := strings.Replace(rel_path, "root", current_dir, 1)
	// fmt.Println("rel_path:", rel_path, "--> abs_path:", abs_path)
	return abs_path
}

func GetFileStatus(rel_path string) (FileStatus, error) {
	abs_path := RelPath2AbsPath(rel_path)
	name := strings.Replace(rel_path, "root/", "", 1)
	hash := sha1.Sum([]byte(abs_path))

	f, _ := os.Open(abs_path)
	defer f.Close()

	var file_status FileStatus
	file_status.Name = name
	file_status.Hash = hash
	if fi, err := f.Stat(); err == nil {
		file_status.Size = uint32(fi.Size())
		file_status.Mode = uint32(fi.Mode())
	}

	return file_status, nil

}

func GetParentName(tree *Tree) string {
	tmp := strings.Split((*tree).Path, "/")
	return strings.Join(tmp[:len(tmp)-1], "/")
}

func CreateTree(originals []string) {

	var names []string

	for _, original := range originals {
		if _, err := os.Stat(original); err != nil {
			continue
		}
		original = "root/" + original
		elements := strings.Split(original, "/")
		for i := 1; i <= len(elements); i++ {
			new_name := strings.Join(elements[:i], "/")
			flag := true
			for _, name := range names {
				if name == new_name {
					flag = false
					break
				}
			}
			if flag {
				names = append(names, new_name)
			}
		}
	}

	var trees []*Tree
	for _, name := range names {
		var tree Tree
		tree.Path = name
		trees = append(trees, &tree)
	}

	for _, tree := range trees {

		parent_path := GetParentName(tree)
		for _, parent_tree := range trees {
			if (*parent_tree).Path == parent_path {
				(*parent_tree).Children = append((*parent_tree).Children, tree)
			}
		}
	}

	for _, tree := range trees {
		if (*tree).Path == "root" {
			Walk(tree, "")
		}
	}

}

func main() {
	var samples = []string{
		"ABC/123/A.py",
		"ABC/123/B.py",
		"ABC/sample.py",
		"123.py",
		"ABC/124/C.py",
		"ABC/124/Answer/C.py",
		"ABC/124/D.py",
		"ARC/C.py",
		"ARC/1/D.py",
		"ARC/124/Answer/E.py",
	}
	CreateTree(samples)

}
