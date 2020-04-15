package main

import (
	"file_check/tool"
	"fmt"
	"os"
)

const start string = "start"
const check string = "check"

var TargetPath = ""

func main() {
	// 校验传参是否正确
	args := os.Args
	if len(args) != 3 || (args[1] != start && args[1] != check) {
		fmt.Print(`
            Usage: 
                filecheck [command] [path]
            The commands are:
                start  重新初始化当前文件夹
                check  检测当前文件夹的改动
            `)
		os.Exit(0)
	}
	// 目标文件夹路径
	model := args[1]
	// 避免有些地方需要添加/,提前加上
	TargetPath = args[2] + "/"

	if model == start {
		_ = os.RemoveAll(TargetPath + tool.ProjectDir)
		// 目标文件夹的文件信息
		folder := tool.GetTargetList(TargetPath)
		// 保存到本地
		tool.SaveDataToLocal(folder, TargetPath)
		fmt.Println("初始化完成, 请勿删除'.filecheck'文件夹及其文件")
	}
	// 为了使用规范,还是不写else了吧?
	if model == check {
		// 获取上次校对的信息
		latest := tool.GetDataFromLocal(TargetPath)
		// 目标文件夹的当前信息
		current := tool.GetTargetList(TargetPath)
		tool.SaveDataToLocal(current, TargetPath)
		// 将新的文件进行对比
		tool.CompareData(&latest, &current)
		tool.SaveDataToLocal(current, ".")
		tool.SortFolder(&current)
		tool.PrintFolder(&current, "")
		// fmt.Printf("save: %#v\n", save)
	}
}
