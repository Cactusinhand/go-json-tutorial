package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Cactusinhand/go-json-tutorial/tutorial15/jsonutil"
)

func main() {
	// 获取程序名和命令行参数
	args := os.Args
	progName := filepath.Base(args[0])

	// 如果没有参数，打印帮助信息
	if len(args) < 2 {
		jsonutil.PrintUsage(progName)
		os.Exit(1)
	}

	// 将命令行参数转发给命令处理函数
	if err := jsonutil.DispatchCommand(progName, args[1:]); err != nil {
		// 处理错误信息，确保它是一个干净、格式化的错误输出
		errMsg := strings.TrimSpace(err.Error())
		if !strings.HasPrefix(errMsg, "错误:") {
			errMsg = "错误: " + errMsg
		}
		os.Stderr.WriteString(errMsg + "\n")
		os.Exit(1)
	}
}
