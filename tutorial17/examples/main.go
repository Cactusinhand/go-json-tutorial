package main

import (
	"fmt"
	"os"
	"strings"

	"./security"
)

func main() {
	// 显示标题
	printHeader("Go JSON 库 (tutorial17) 示例程序")

	// 检查命令行参数
	if len(os.Args) > 1 {
		example := strings.ToLower(os.Args[1])

		switch example {
		case "security":
			security.RunSecurityExample()
		// 可以添加其他示例
		default:
			fmt.Printf("未知的示例: %s\n", example)
			showUsage()
		}
	} else {
		// 默认运行所有示例
		fmt.Println("运行所有示例...")
		fmt.Println()

		// 安全功能示例
		security.RunSecurityExample()

		// 运行完毕
		fmt.Println("\n所有示例已完成。")
	}
}

// printHeader 打印漂亮的标题
func printHeader(title string) {
	width := len(title) + 10
	border := strings.Repeat("=", width)

	fmt.Println(border)
	fmt.Printf("%s %s %s\n", "==", title, "==")
	fmt.Println(border)
	fmt.Println()
}

// showUsage 显示使用说明
func showUsage() {
	fmt.Println("\n用法:")
	fmt.Println("  go run main.go [示例名称]")
	fmt.Println("\n可用示例:")
	fmt.Println("  security  - 安全功能示例")
	// 其他示例...
	fmt.Println("  (无参数)  - 运行所有示例")
}
