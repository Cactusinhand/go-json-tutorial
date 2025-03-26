#!/bin/bash

# 编译示例程序
echo "编译示例程序..."
go build -o enhanced_error_demo enhanced_error_demo.go

# 检查编译是否成功
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi

# 运行示例程序
echo "运行示例程序..."
echo "========================================================"
./enhanced_error_demo

# 清理编译产物
rm -f enhanced_error_demo 