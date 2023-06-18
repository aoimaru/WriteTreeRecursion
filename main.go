package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

func (fs FileStatus) GetType() string {
	if _, err := os.Stat(fs.Name); err != nil {
		return "blob"
	} else {
		return "tree"
	}

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

func Compress(buffer []byte) []uint8 {
	var compressed bytes.Buffer
	zlib_writer := zlib.NewWriter(&compressed)
	zlib_writer.Write(buffer)
	zlib_writer.Close()
	return compressed.Bytes()
}

func Walking(tree *Tree) (hash string) {
	buffer := make([]byte, 0)
	// buffer = append(buffer, []byte("tree")...)
	header := []byte{116, 114, 101, 101, 32, 51, 53, 51}
	buffer = append(buffer, header...)
	// buffer = append(buffer, 0)

	for _, child_tree := range (*tree).Children {
		file_status, _ := GetFileStatus((*child_tree).Path)
		entry_buffer := make([]byte, 0)
		entry_buffer = append(entry_buffer, 0)
		if file_status.GetType() == "blob" {
			entry_buffer = append(entry_buffer, []byte("100644"+" ")...)
			entry_buffer = append(entry_buffer, []byte(file_status.Name+" ")...)
			// entry_buffer = append(entry_buffer, 0)
			entry_buffer = append(entry_buffer, file_status.Hash[0:20]...)
		} else {
			entry_buffer = append(entry_buffer, []byte("40000"+" ")...)
			entry_buffer = append(entry_buffer, []byte(file_status.Name+" ")...)
			// entry_buffer = append(entry_buffer, 0)
			child_hash := Walking(child_tree)
			entry_buffer = append(entry_buffer, []byte(child_hash)...)

		}
		buffer = append(buffer, entry_buffer...)
	}
	compressed_buffer := Compress(buffer)
	sha1 := sha1.New()
	sha1.Write(compressed_buffer)

	new_hash := hex.EncodeToString(sha1.Sum(nil))

	object_path := "/home/aoimaru/document/go_project/Recursion/.bakibaki/objects/"
	fmt.Println(object_path, new_hash[:2], new_hash[2:])
	if _, err := os.Stat(object_path + new_hash[:2]); err != nil {
		if err := os.MkdirAll(object_path+new_hash[:2], 1755); err != nil {
			return "sssssss"
		}
	}

	new_writer, _ := os.Create(object_path + new_hash[:2] + "/" + new_hash[2:])
	defer new_writer.Close()

	count, err := new_writer.Write(compressed_buffer)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("write %d bytes\n", count)

	return new_hash
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

	// for _, tree := range trees {
	// 	if (*tree).Path == "root" {
	// 		Walk(tree, "")
	// 	}
	// }

	for _, tree := range trees {
		if (*tree).Path == "root" {
			_ = Walking(tree)
		}
	}

}

func CatFile(hash string) {
	fmt.Println()
	fmt.Println("hash:", hash)
	root_dir := "/home/aoimaru/document/go_project/Recursion/.bakibaki/objects/"
	tree_path := root_dir + hash[:2] + "/" + hash[2:]
	f, _ := os.Open(tree_path)
	defer f.Close()

	buffer := make([]byte, 0)
	buf := make([]byte, 64)
	for {
		n, _ := (*f).Read(buf)
		if n == 0 {
			break
		}
		buffer = append(buffer, buf...)
	}
	// fmt.Println(buffer)

	extracting_buffer := bytes.NewBuffer(buffer)
	zlib_f, _ := zlib.NewReader(extracting_buffer)

	zlib_buffer, _ := ioutil.ReadAll(zlib_f)
	// fmt.Println(zlib_buffer)

	entries := make([][]byte, 0)
	entry := make([]byte, 0)

	for _, zlib_buf := range zlib_buffer {
		if zlib_buf == 0 {
			entries = append(entries, entry)
			entry = make([]byte, 0)
		}
		entry = append(entry, zlib_buf)
	}
	entries = append(entries, entry)

	for _, entry := range entries {
		fmt.Println(string(entry))
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
	fmt.Println(samples)
	// CreateTree(samples)
	CatFile("17557b5615e7e9a05a2fd598c5d3fd07791f0f0a")
	CatFile("18f885e413a0a63f12dfc2655b69a9c716ef7d1d")
	CatFile("2bd8b99210a3c17aa5e54bb1e95d3311048b0447")
	CatFile("d056969fd6da5e11bc43b9afb0c539d61351ad5c")

}
