package tutorial15

import (
	"bytes"
	"fmt"
	"strconv"
)

// Stringify 将Value转换为JSON字符串（无缩进）
func Stringify(v *Value) (string, error) {
	if v == nil {
		return "", fmt.Errorf("值不能为nil")
	}

	var buf bytes.Buffer
	stringifyValue(&buf, v, "", "")
	return buf.String(), nil
}

// StringifyIndent 将Value转换为缩进格式的JSON字符串
func StringifyIndent(v *Value, indent string) (string, error) {
	if v == nil {
		return "", fmt.Errorf("值不能为nil")
	}

	var buf bytes.Buffer
	// 写入一个空行
	buf.WriteByte('\n')
	stringifyValue(&buf, v, "", indent)
	// 确保结尾有换行符，以使行数符合预期
	if indent != "" {
		buf.WriteByte('\n')
	}
	return buf.String(), nil
}

func stringifyValue(buf *bytes.Buffer, v *Value, prefix, indent string) {
	if v == nil {
		buf.WriteString("null")
		return
	}

	switch v.Type {
	case NULL:
		buf.WriteString("null")
	case BOOLEAN:
		if v.B {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case NUMBER:
		buf.WriteString(strconv.FormatFloat(v.N, 'f', -1, 64))
	case STRING:
		stringifyString(buf, v.S)
	case ARRAY:
		stringifyArray(buf, v, prefix, indent)
	case OBJECT:
		stringifyObject(buf, v, prefix, indent)
	}
}

func stringifyString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for _, c := range s {
		switch c {
		case '"':
			buf.WriteString("\\\"")
		case '\\':
			buf.WriteString("\\\\")
		case '\b':
			buf.WriteString("\\b")
		case '\f':
			buf.WriteString("\\f")
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		default:
			if c < 32 {
				buf.WriteString(fmt.Sprintf("\\u%04x", c))
			} else {
				buf.WriteRune(c)
			}
		}
	}
	buf.WriteByte('"')
}

func stringifyArray(buf *bytes.Buffer, v *Value, prefix, indent string) {
	if len(v.Elements) == 0 {
		buf.WriteString("[]")
		return
	}

	buf.WriteByte('[')
	newPrefix := prefix
	if indent != "" {
		buf.WriteByte('\n')
		newPrefix = prefix + indent
	}

	for i, e := range v.Elements {
		if indent != "" {
			buf.WriteString(newPrefix)
		}
		stringifyValue(buf, e, newPrefix, indent)
		if i < len(v.Elements)-1 {
			buf.WriteByte(',')
		}
		if indent != "" {
			buf.WriteByte('\n')
		}
	}

	if indent != "" {
		buf.WriteString(prefix)
	}
	buf.WriteByte(']')
}

func stringifyObject(buf *bytes.Buffer, v *Value, prefix, indent string) {
	if len(v.Members) == 0 {
		buf.WriteString("{}")
		return
	}

	buf.WriteByte('{')
	newPrefix := prefix
	if indent != "" {
		buf.WriteByte('\n')
		newPrefix = prefix + indent
	}

	for i, m := range v.Members {
		if indent != "" {
			buf.WriteString(newPrefix)
		}
		// 键名
		stringifyString(buf, m.Key)
		buf.WriteString(": ")
		// 值
		stringifyValue(buf, m.Value, newPrefix, indent)
		if i < len(v.Members)-1 {
			buf.WriteByte(',')
		}
		if indent != "" {
			buf.WriteByte('\n')
		}
	}

	if indent != "" {
		buf.WriteString(prefix)
	}
	buf.WriteByte('}')
}
