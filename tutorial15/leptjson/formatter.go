// formatter.go - JSON格式化工具
package leptjson

import (
	"fmt"
	"strconv"
	"strings"
)

// Stringify 将JSON值转换为紧凑字符串格式
func Stringify(v *Value) (string, error) {
	return StringifyIndent(v, "")
}

// StringifyIndent 将JSON值转换为带缩进的格式化字符串
// indent 参数指定缩进字符串，如果为空则生成紧凑格式
func StringifyIndent(v *Value, indent string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("输入的JSON值不能为空")
	}

	var sb strings.Builder
	err := stringifyWithIndent(v, &sb, indent, 0)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// stringifyWithIndent 递归地将值格式化为带缩进的JSON字符串
func stringifyWithIndent(v *Value, sb *strings.Builder, indent string, level int) error {
	if v == nil {
		return fmt.Errorf("JSON值不能为空")
	}

	isCompact := indent == ""

	switch v.Type {
	case JSON_NULL:
		sb.WriteString("null")
	case JSON_FALSE:
		sb.WriteString("false")
	case JSON_TRUE:
		sb.WriteString("true")
	case JSON_NUMBER:
		formatNumber(v.N, sb)
	case JSON_STRING:
		escapeString(v.S, sb)
	case JSON_ARRAY:
		if len(v.A) == 0 {
			sb.WriteString("[]")
			return nil
		}

		sb.WriteString("[")
		if !isCompact {
			sb.WriteString("\n")
		}

		for i, elem := range v.A {
			if !isCompact {
				// 缩进
				for j := 0; j < level+1; j++ {
					sb.WriteString(indent)
				}
			}

			if err := stringifyWithIndent(elem, sb, indent, level+1); err != nil {
				return err
			}

			if i != len(v.A)-1 {
				if isCompact {
					sb.WriteString(",")
				} else {
					sb.WriteString(",\n")
				}
			} else if !isCompact {
				sb.WriteString("\n")
			}
		}

		if !isCompact {
			// 缩进
			for j := 0; j < level; j++ {
				sb.WriteString(indent)
			}
		}
		sb.WriteString("]")
	case JSON_OBJECT:
		if len(v.M) == 0 {
			sb.WriteString("{}")
			return nil
		}

		sb.WriteString("{")
		if !isCompact {
			sb.WriteString("\n")
		}

		for i, pair := range v.M {
			if !isCompact {
				// 缩进
				for j := 0; j < level+1; j++ {
					sb.WriteString(indent)
				}
			}

			// 键名
			escapeString(pair[0], sb)
			if isCompact {
				sb.WriteString(":")
			} else {
				sb.WriteString(": ")
			}

			// 值
			valueObj := &Value{}
			if err := Parse(valueObj, pair[1]); err != PARSE_OK {
				return fmt.Errorf("无法解析对象值: %v", err)
			}

			if err := stringifyWithIndent(valueObj, sb, indent, level+1); err != nil {
				return err
			}

			if i != len(v.M)-1 {
				if isCompact {
					sb.WriteString(",")
				} else {
					sb.WriteString(",\n")
				}
			} else if !isCompact {
				sb.WriteString("\n")
			}
		}

		if !isCompact {
			// 缩进
			for j := 0; j < level; j++ {
				sb.WriteString(indent)
			}
		}
		sb.WriteString("}")
	default:
		return fmt.Errorf("无效的JSON值类型: %d", v.Type)
	}

	return nil
}

// formatNumber 格式化数字
func formatNumber(n float64, sb *strings.Builder) {
	// 检查整数
	if n == float64(int64(n)) {
		sb.WriteString(strconv.FormatInt(int64(n), 10))
	} else {
		// 浮点数，避免科学计数法
		sb.WriteString(strconv.FormatFloat(n, 'f', -1, 64))
	}
}

// escapeString 转义字符串并写入
func escapeString(s string, sb *strings.Builder) {
	sb.WriteByte('"')
	for _, c := range s {
		switch c {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\b':
			sb.WriteString("\\b")
		case '\f':
			sb.WriteString("\\f")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			if c < 0x20 {
				// 控制字符使用 \u 转义
				fmt.Fprintf(sb, "\\u%04x", c)
			} else {
				sb.WriteRune(c)
			}
		}
	}
	sb.WriteByte('"')
}
