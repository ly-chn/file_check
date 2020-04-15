package tool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
)

const (
	// 最新文件相对路径
	latestPath = ".filecheck/latest.json"
	// 最新文件保存的文件夹的相对路径
	ProjectDir = ".filecheck/"
	// 数据文件后缀
	suffix = ".json"
)

// 状态
const (
	statusUnchanged = "无变化"
	statusModified  = "已修改"
	statusDeleted   = "已删除"
	statusNewFile   = "新增"
)

// 颜色
var (
	greenBg      = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBg      = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBg     = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg        = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBg       = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magentaBg    = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyanBg       = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	green        = string([]byte{27, 91, 51, 50, 109})
	white        = string([]byte{27, 91, 51, 55, 109})
	yellow       = string([]byte{27, 91, 51, 51, 109})
	red          = string([]byte{27, 91, 51, 49, 109})
	blue         = string([]byte{27, 91, 51, 52, 109})
	magenta      = string([]byte{27, 91, 51, 53, 109})
	cyan         = string([]byte{27, 91, 51, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

var latest = ""

// 输出文件夹及其文件,包含状态
func PrintFolder(folder *Folder, indent string) {
	var text string
	if folder.Name == "" {
		text = "📁目标文件夹"
	} else {
		text = "📁" + folder.Name
	}
	prefix := "|" + indent

	printWithColor(folder.Status, prefix, text)
	for _, child := range folder.Children {
		PrintFolder(child, indent+"--")
	}
	for _, file := range folder.Files {
		printWithColor(file.Status, prefix, "🖹"+file.Name)
	}
}

// 根据状态输出带颜色的文字
func printWithColor(status string, prefix string, text string) {
	switch status {
	case statusDeleted:
		fmt.Println(prefix, redBg, text, reset)
		break
	case statusNewFile:
		fmt.Println(prefix, greenBg, text, reset)
		break
	case statusUnchanged:
		fmt.Println(prefix, text)
		break
	case statusModified:
		fmt.Println(prefix, blueBg, text, reset)
		break
	default:
		fmt.Println(prefix, text, "---解析异常")
		break
	}
}

// 排序文件夹和文件
func SortFolder(folder *Folder) {
	sort.Sort(FileSlice(folder.Files))
	sort.Sort(FolderSlice(folder.Children))
	for _, child := range folder.Children {
		SortFolder(child)
	}
}

// 对比文件
// @latest 上次的校对信息
// @current 本次的校对信息
func CompareData(oldFolder, currentFolder *Folder) {
	compareFolder(oldFolder, currentFolder)
	// 将未定义的文件夹设置为新增,省的再次遍历两个树了
	setFolderBeStatus(currentFolder, statusNewFile)
}

// 判断文件的状态,
func compareFile(oldFolder, currentFolder *Folder) {
	oldFiles := oldFolder.Files
	currentFiles := currentFolder.Files
	// 未删除的文件
	for _, currentFile := range currentFiles {
		currentFile.Status = statusNewFile
		for _, oldFile := range oldFiles {
			// 判断当前文件的状态
			if oldFile.Path == currentFile.Path {
				currentFile.Status = statusModified
				if oldFile.Md5 == currentFile.Md5 {
					currentFile.Status = statusUnchanged
				}
			}
		}
	}
	// 已删除的文件
Deleted:
	for _, oldFile := range oldFiles {
		for _, currentFile := range currentFiles {
			if currentFile.Path == oldFile.Path {
				continue Deleted
			}
		}
		oldFile.Status = statusDeleted
		currentFolder.Files = append(currentFiles, oldFile)
	}
}

// 判断文件夹的状态
func compareFolder(oldFolder, currentFolder *Folder) {
	oldFolderChildren := oldFolder.Children
	currentFolderChildren := currentFolder.Children
	currentFolder.Status = statusNewFile
	if currentFolder.Path == oldFolder.Path {
		currentFolder.Status = statusModified
		if currentFolder.Md5 == oldFolder.Md5 {
			currentFolder.Status = statusUnchanged
		}
	}
	// 对比文件
	compareFile(oldFolder, currentFolder)
	// 已删除
	var currentFolderChild = new(Folder)
Deleted:
	for _, oldFolderChild := range oldFolderChildren {
		for _, currentFolderChild = range currentFolderChildren {
			if oldFolderChild.Path == currentFolderChild.Path {
				continue Deleted
			}
		}
		oldFolderChild.Status = statusDeleted
		setFolderBeStatus(oldFolderChild, statusDeleted)
		currentFolder.Children = append(currentFolder.Children, oldFolderChild)
	}
	// 对比差异
	for _, currentFolderChild := range currentFolderChildren {
		if currentFolderChild.Status == statusNewFile || currentFolderChild.Status == statusDeleted {
			compareFolder(&Folder{}, currentFolderChild)
			continue
		}
		for _, oldFolderChild := range oldFolderChildren {
			if oldFolderChild.Path == currentFolderChild.Path {
				compareFolder(oldFolderChild, currentFolderChild)
			}
		}
	}
}

// 将一个文件夹及其子文件标记为已删除
func setFolderBeStatus(folder *Folder, status string) {
	if folder.Status == "" {
		folder.Status = status
	}
	for _, file := range folder.Files {
		if file.Status == "" {
			file.Status = status
		}
	}
	for _, child := range folder.Children {
		setFolderBeStatus(child, status)
	}
}

// 从本地读取记录
func GetDataFromLocal(targetPath string) (folder Folder) {
	bytes, err := ioutil.ReadFile(filepath.Join(targetPath, latestPath))
	if err != nil {
		fmt.Println(`读取文件发生错误: ` + err.Error() + `
            如果没有初始化,请尝试执行:
            filecheck start [path]
        `)
		os.Exit(0)
	}
	_ = json.Unmarshal(bytes, &folder)
	return
}

// 保存到本地文件夹
func SaveDataToLocal(folder Folder, targetPath string) {
	// 最新文件绝对路径
	latest = targetPath + latestPath
	// 判断文件是否存在
	info, err := os.Stat(latest)
	if err != nil {
		// 创建文件夹并赋予777权限
		_ = os.MkdirAll(targetPath+ProjectDir, os.ModePerm)
		SaveToFile(folder)
	} else {
		// 备份文件,此处没有检查文件是否被修改~命名以最后修改时间
		modTime := info.ModTime().Format("20160102150405")
		dir, _ := filepath.Split(latest)
		_ = os.Rename(latest, dir+modTime+suffix)
		// 保存文件
		SaveToFile(folder)
	}

	if runtime.GOOS == "windows" {
		// 隐藏文件夹
		HideWindowsFile(targetPath + ProjectDir)
	}
}

// todo: 尚未确定linux下的可行性, 参考代码:https://stackoverflow.com/questions/54139606/how-to-create-a-hidden-file-in-windows-mac-linux
// linux下隐藏文件据说可以直接加"."去除"."就行了
// 隐藏文件
func HideWindowsFile(filename string) {
	filenameW, _ := syscall.UTF16PtrFromString(filename)
	_ = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
}

// 将数据保存到latest文件
func SaveToFile(folder Folder) {
	// 保存folder对象到本地json
	marshal, _ := json.Marshal(folder)
	_ = ioutil.WriteFile(latest, marshal, os.ModePerm)
}

// 从文件列表中移除文件,防止文件太多时反向遍历占用空间,貌似还得遍历一遍,可能偷鸡不成蚀把米
// func removeFileFromFileList(fileList []*File, target *File) {
//     for index, file := range fileList {
//         if file == target {
//             fileList[index] = fileList[len(fileList)-1]
//             fileList = fileList[:len(fileList)-1]
//         }
//     }
// }

// 显示windows文件,如果文件不隐藏的话,这个也暂时用不到了
// func ShowWindowsFile(filename string) {
//     filenameW, _ := syscall.UTF16PtrFromString(filename)
//     _ = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_NORMAL)
// }
