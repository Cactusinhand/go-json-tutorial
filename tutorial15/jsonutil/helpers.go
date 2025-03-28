// helpers.go - 提供辅助函数
package jsonutil

import (
	"fmt"
	"io/ioutil"
	"os"
)

// readJSONFile 从文件或标准输入读取JSON内容
func readJSONFile(filePath string) ([]byte, error) {
	if filePath == "" {
		// 从标准输入读取
		return ioutil.ReadAll(os.Stdin)
	}

	// 从文件读取
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法读取文件: %v", err)
	}
	return content, nil
}

// writeOutput 写入到文件或标准输出
func writeOutput(filePath string, content []byte) error {
	if filePath == "" {
		// 写入标准输出
		_, err := os.Stdout.Write(content)
		if err != nil {
			return fmt.Errorf("写入标准输出失败: %v", err)
		}
		return nil
	}

	// 写入文件
	err := ioutil.WriteFile(filePath, content, 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}
	return nil
}
