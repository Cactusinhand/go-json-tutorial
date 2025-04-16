package leptjson

import (
	"fmt"
	"strings"
)

// EnhancedError 定义了一个增强的错误类型，包含详细信息
type EnhancedError struct {
	Code          ParseError // 错误码
	Message       string     // 错误消息
	Line          int        // 行号
	Column        int        // 列号
	Context       string     // 错误发生的上下文
	Pointer       string     // 错误位置指针（比如 "----^"）
	SourceInput   string     // 输入源
	IsRecoverable bool       // 是否可恢复
}

// Error 实现error接口
func (e *EnhancedError) Error() string {
	result := fmt.Sprintf("[错误码: %d] %s，位置：第%d行，第%d列\n",
		e.Code, e.Message, e.Line, e.Column)

	if e.Context != "" {
		result += e.Context + "\n"
		if e.Pointer != "" {
			result += e.Pointer + "\n"
		}
	}

	if e.IsRecoverable {
		result += "注意：此错误可恢复，解析将继续\n"
	}

	return result
}

// ParseOptions 定义解析选项
type ParseOptions struct {
	MaxDepth          int  // 最大嵌套深度
	AllowComments     bool // 是否允许注释
	AllowTrailing     bool // 是否允许尾随逗号
	StrictMode        bool // 严格模式（更严格的检查）
	RecoverFromErrors bool // 是否从非致命错误恢复

	// 新增安全选项
	MaxStringLength int     // 最大字符串长度
	MaxArraySize    int     // 最大数组元素数量
	MaxObjectSize   int     // 最大对象成员数量
	MaxTotalSize    int     // 最大输入字节数
	MaxNumberValue  float64 // 最大数字值
	MinNumberValue  float64 // 最小数字值
	EnabledSecurity bool    // 是否启用安全检查
}

// DefaultParseOptions 返回默认解析选项
func DefaultParseOptions() ParseOptions {
	return ParseOptions{
		MaxDepth:          1000,
		AllowComments:     false,
		AllowTrailing:     false,
		StrictMode:        true,
		RecoverFromErrors: false,

		MaxStringLength: 8192,    // 8KB
		MaxArraySize:    10000,   // 最多1万个元素
		MaxObjectSize:   10000,   // 最多1万个成员
		MaxTotalSize:    1048576, // 1MB
		MaxNumberValue:  1e308,   // 接近float64最大值
		MinNumberValue:  -1e308,  // 接近float64最小值
		EnabledSecurity: true,    // 默认启用安全检查
	}
}

// 新增安全相关错误码
const (
	// 原有错误码后追加
	PARSE_MAX_STRING_LENGTH_EXCEEDED ParseError = 16
	PARSE_MAX_ARRAY_SIZE_EXCEEDED    ParseError = 17
	PARSE_MAX_OBJECT_SIZE_EXCEEDED   ParseError = 18
	PARSE_MAX_TOTAL_SIZE_EXCEEDED    ParseError = 19
	PARSE_NUMBER_RANGE_EXCEEDED      ParseError = 20
	PARSE_SECURITY_VIOLATION         ParseError = 21
)

// createEnhancedError 创建详细的错误信息
func createEnhancedError(code ParseError, json string, index int, linePositions []int, message string, isRecoverable bool) *EnhancedError {
	// 计算行号和列号
	line := 1
	column := index + 1

	// 找到当前行
	for i, pos := range linePositions {
		if index >= pos {
			line = i + 2 // 行号从1开始，且首行不在linePositions中
			column = index - pos + 1
		} else {
			break
		}
	}

	// 提取上下文
	lineStart := 0
	lineEnd := len(json)

	// 找到当前行的开始位置
	if line > 1 && len(linePositions) >= line-1 {
		lineStart = linePositions[line-2]
	}

	// 找到当前行的结束位置
	if line <= len(linePositions) {
		lineEnd = linePositions[line-1] - 1
	} else {
		for i := lineStart; i < len(json); i++ {
			if json[i] == '\n' {
				lineEnd = i
				break
			}
		}
	}

	// 提取当前行内容
	contextLine := ""
	if lineStart < len(json) && lineEnd <= len(json) {
		contextLine = json[lineStart:lineEnd]
	}

	// 创建指针字符串
	pointerLine := ""
	if len(contextLine) > 0 {
		spaces := strings.Repeat(" ", column-1)
		pointerLine = spaces + "^"
	}

	return &EnhancedError{
		Code:          code,
		Message:       message,
		Line:          line,
		Column:        column,
		Context:       contextLine,
		Pointer:       pointerLine,
		SourceInput:   json,
		IsRecoverable: isRecoverable,
	}
}

// GetErrorMessage 根据错误码获取错误消息
func GetErrorMessage(code ParseError) string {
	switch code {
	case PARSE_OK:
		return "解析成功"
	case PARSE_EXPECT_VALUE:
		return "期望一个值"
	case PARSE_INVALID_VALUE:
		return "无效的值"
	case PARSE_ROOT_NOT_SINGULAR:
		return "根节点后有多余内容"
	case PARSE_NUMBER_TOO_BIG:
		return "数字值过大"
	case PARSE_MISS_QUOTATION_MARK:
		return "缺少引号"
	case PARSE_INVALID_STRING_ESCAPE:
		return "无效的转义字符"
	case PARSE_INVALID_STRING_CHAR:
		return "无效的字符"
	case PARSE_INVALID_UNICODE_HEX:
		return "无效的Unicode十六进制数字"
	case PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的Unicode代理对"
	case PARSE_MISS_COMMA_OR_SQUARE_BRACKET:
		return "缺少逗号或方括号"
	case PARSE_MISS_KEY:
		return "缺少键"
	case PARSE_MISS_COLON:
		return "缺少冒号"
	case PARSE_MISS_COMMA_OR_CURLY_BRACKET:
		return "缺少逗号或大括号"
	case PARSE_COMMENT_NOT_CLOSED:
		return "注释未闭合"
	case PARSE_MAX_DEPTH_EXCEEDED:
		return "超过最大嵌套深度"
	case PARSE_MAX_STRING_LENGTH_EXCEEDED:
		return "超过最大字符串长度"
	case PARSE_MAX_ARRAY_SIZE_EXCEEDED:
		return "超过最大数组元素数量"
	case PARSE_MAX_OBJECT_SIZE_EXCEEDED:
		return "超过最大对象成员数量"
	case PARSE_MAX_TOTAL_SIZE_EXCEEDED:
		return "超过最大输入大小"
	case PARSE_NUMBER_RANGE_EXCEEDED:
		return "数值超出允许范围"
	case PARSE_SECURITY_VIOLATION:
		return "安全策略违规"
	default:
		return "未知错误"
	}
}
