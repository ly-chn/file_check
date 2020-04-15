package tool

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Folder struct {
	Files    []*File   `json:"files"`    // 该目录下的文件
	Name     string    `json:"name"`     // 文件夹名
	Path     string    `json:"path"`     // 相对路径
	Md5      string    `json:"md5"`      // 文件夹名+文件夹下所有文件的md5值
	Status   string    `json:"status"`   // 状态: 新增/修改/删除/没变
	Children []*Folder `json:"children"` // 子文件夹
}

type File struct {
	Name   string `json:"name"`   // 文件名
	Path   string `json:"path"`   // 相对路径(相对于控制台输入的文件夹)
	Md5    string `json:"md5"`    // 文件的md5值
	Status string `json:"status"` // 状态: 新增/修改/删除/没变
}

// func GBK2UTF8(src []byte) (string, error) {
//     GB18030 := simplifiedchinese.All[0]
//     bytes, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(src), GB18030.NewDecoder()))
//     return string(bytes), err
// }

// 获取文件/文件夹列表
// @newFile: 添加文件是新增文件的标记
func GetFileOrDirList(path string, folder *Folder, relativePath string) {
	// 读取当前文件夹下的所有文件和文件夹
	fileInfos, _ := ioutil.ReadDir(path)
	for _, info := range fileInfos {
		// 将文件和文件夹进行分类
		if info.IsDir() {
			// 如果是目录名为ProjectDir则忽略
			if filepath.Base(info.Name()) == filepath.Base(ProjectDir) {
				continue
			}
			// 相对目录
			subPath := filepath.Join(relativePath, info.Name())
			f := Folder{
				Name: info.Name(),
				Path: subPath,
			}
			folder.Children = append(folder.Children, &f)
			GetFileOrDirList(filepath.Join(path, info.Name()), &f, subPath)
		} else {
			// 追加文件
			file := File{
				Name: info.Name(),
				Path: filepath.Join(relativePath, info.Name()),
				Md5:  getMd5ForFile(filepath.Join(path, info.Name())),
			}
			folder.Files = append(folder.Files, &file)
		}
	}
}

// 为目录设置 md5 值(目录名称+文件的 md5 值排序后二次md5的结果)
func SetFolderMd5(folder *Folder) {
	// 子文件夹的md5的集合
	var subFolderMd5s []string
	// 子文件的 md5 集合
	var subFileMd5s []string
	// 计算子文件夹的 md5
	for _, subFolder := range folder.Children {
		SetFolderMd5(subFolder)
		// 整合md5结果
		subFolderMd5s = append(subFolderMd5s, subFolder.Md5)
	}
	// 整合子文件的 md5
	for _, file := range folder.Files {
		subFileMd5s = append(subFileMd5s, file.Md5)
	}

	// 排序
	sort.Strings(subFolderMd5s)
	sort.Strings(subFileMd5s)
	target := folder.Name + strings.Join(subFolderMd5s, "") + strings.Join(subFileMd5s, "")

	// md5(md5)
	hash := md5.New()
	_, _ = io.WriteString(hash, target)
	folder.Md5 = hex.EncodeToString(hash.Sum(nil))
}

// 获取文件的 md5,忽略了异常处理
func getMd5ForFile(path string) string {
	// 打开文件
	file, _ := os.Open(path)
	defer file.Close()
	// 计算md5
	hash := md5.New()
	_, _ = io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil))
}

// 收集数据
func GetTargetList(targetPath string) (folder Folder) {
	// 判断文件夹是否存在
	info, err := os.Stat(targetPath)
	if err != nil {
		panic("所选文件夹不存在, 程序终止运行...")
	}

	// 当前文件夹名称, 暂时忽略
	_ = info.Name()

	// 整理文件树
	GetFileOrDirList(targetPath, &folder, "")
	// 处理 md5
	SetFolderMd5(&folder)
	return
}
