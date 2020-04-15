package tool

import (
	"bytes"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
)

type FolderSlice []*Folder

func (s FolderSlice) Len() int {
	return len(s)
}

func (s FolderSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s FolderSlice) Less(i, j int) bool {
	a, _ := UTF82GBK(s[i].Path)
	b, _ := UTF82GBK(s[j].Path)
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

type FileSlice []*File

func (s FileSlice) Len() int {
	return len(s)
}

func (s FileSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s FileSlice) Less(i, j int) bool {
	a, _ := UTF82GBK(s[i].Path)
	b, _ := UTF82GBK(s[j].Path)
	bLen := len(b)
	for idx, chr := range a {
		if idx > bLen-1 {
			return false
		}
		if chr != b[idx] {
			return chr < b[idx]
		}
	}
	return true
}

// 对中文的排序支持
func UTF82GBK(src string) ([]byte, error) {
	GB18030 := simplifiedchinese.All[0]
	return ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(src)), GB18030.NewEncoder()))
}
