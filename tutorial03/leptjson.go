// leptjson.go - Go语言版JSON库实现
//
// 这个包实现了一个简单的JSON解析器，支持基本的JSON值类型
// 包括null、true、false、数字和字符串等。后续版本将支持数组和对象。
//
// 字符串解析的特殊情况处理：
// 1. 基本规则：
//   - 字符串必须以双引号开始和结束
//   - 单引号是非法的
//   - 字符串中的双引号必须转义
//   - 空字符串""是合法的
//
// 2. 转义序列：
//   - 标准转义序列：\"、\\、\/、\b、\f、\n、\r、\t
//   - 其他转义序列都是非法的（如\v、\a等）
//   - 转义序列必须完整（如"\n"是合法的，但"\n"后面跟着其他字符是非法的）
//
// 3. Unicode处理：
//   - 支持\uXXXX格式的Unicode转义
//   - 支持UTF-16代理对（\uD800-\uDFFF）
//   - 代理对必须成对出现（高代理项后必须跟低代理项）
//   - 单独的低代理项是非法的
//
// 4. 控制字符处理：
//   - 所有小于0x20的控制字符必须转义
//   - 未转义的控制字符是非法的
//
// 5. 字符串长度限制：
//   - 理论上没有长度限制
//   - 但实际实现中可能受限于内存大小
//
// 6. 错误处理：
//   - PARSE_MISS_QUOTATION_MARK: 缺少右引号
//   - PARSE_INVALID_STRING_ESCAPE: 无效的转义序列
//   - PARSE_INVALID_STRING_CHAR: 无效的字符
//   - PARSE_INVALID_UNICODE_HEX: 无效的Unicode十六进制
//   - PARSE_INVALID_UNICODE_SURROGATE: 无效的Unicode代理对
package leptjson

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ValueType 表示JSON值的类型
type ValueType int

// String 返回ValueType的字符串表示
func (t ValueType) String() string {
	switch t {
	case NULL:
		return "null"
	case FALSE:
		return "false"
	case TRUE:
		return "true"
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case ARRAY:
		return "array"
	case OBJECT:
		return "object"
	default:
		return "unknown"
	}
}

// JSON值类型常量
const (
	NULL   ValueType = iota // null值
	FALSE                   // false值
	TRUE                    // true值
	NUMBER                  // 数字
	STRING                  // 字符串
	ARRAY                   // 数组
	OBJECT                  // 对象
)

// ParseError 表示解析过程中可能出现的错误
type ParseError int

// 解析错误常量
const (
	PARSE_OK                        ParseError = iota // 解析成功
	PARSE_EXPECT_VALUE                                // 期望一个值
	PARSE_INVALID_VALUE                               // 无效的值
	PARSE_ROOT_NOT_SINGULAR                           // 根节点不唯一
	PARSE_NUMBER_TOO_BIG                              // 数字太大
	PARSE_MISS_QUOTATION_MARK                         // 缺少引号
	PARSE_INVALID_STRING_ESCAPE                       // 无效的转义序列
	PARSE_INVALID_STRING_CHAR                         // 无效的字符
	PARSE_INVALID_UNICODE_HEX                         // 无效的 Unicode 十六进制数字
	PARSE_INVALID_UNICODE_SURROGATE                   // 无效的 Unicode 代理对
)

// Error 实现error接口，返回错误描述
func (e ParseError) Error() string {
	switch e {
	case PARSE_OK:
		return "解析成功"
	case PARSE_EXPECT_VALUE:
		return "期望一个值"
	case PARSE_INVALID_VALUE:
		return "无效的值"
	case PARSE_ROOT_NOT_SINGULAR:
		return "根节点不唯一"
	case PARSE_NUMBER_TOO_BIG:
		return "数字太大"
	case PARSE_MISS_QUOTATION_MARK:
		return "缺少引号"
	case PARSE_INVALID_STRING_ESCAPE:
		return "无效的转义序列"
	case PARSE_INVALID_STRING_CHAR:
		return "无效的字符"
	case PARSE_INVALID_UNICODE_HEX:
		return "无效的 Unicode 十六进制数字"
	case PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的 Unicode 代理对"
	default:
		return fmt.Sprintf("未知错误(%d)", int(e))
	}
}

func (e ParseError) String() string {
	switch e {
	case PARSE_INVALID_STRING_ESCAPE:
		return "无效的转义序列"
	case PARSE_INVALID_UNICODE_HEX:
		return "无效的 Unicode 十六进制数字"
	case PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的 Unicode 代理对"
	}
	return "未知错误"
}

// Value 表示一个JSON值
type Value struct {
	Type ValueType `json:"type"` // 值类型
	N    float64   `json:"n"`    // 数字值（当Type为NUMBER时有效）
	S    string    `json:"s"`    // 字符串值（当Type为STRING时有效）
}

// String 实现fmt.Stringer接口，返回Value的字符串表示
func (v *Value) String() string {
	switch v.Type {
	case NUMBER:
		return fmt.Sprintf("%g", v.N)
	case STRING:
		return v.S
	default:
		return v.Type.String()
	}
}

// context 表示解析过程中的上下文
type context struct {
	json  string
	index int
}

// Parse 解析JSON文本
//
// 接收一个Value指针和JSON字符串，解析JSON并将结果存储在Value中
// 返回解析过程中可能出现的错误
func Parse(v *Value, json string) ParseError {
	var c context
	c.json = json
	c.index = 0
	v.Type = NULL

	parseWhitespace(&c)
	if err := parseValue(&c, v); err != PARSE_OK {
		return err
	}
	parseWhitespace(&c)
	if c.index < len(c.json) {
		return PARSE_ROOT_NOT_SINGULAR
	}
	return PARSE_OK
}

// GetType 获取JSON值的类型
func GetType(v *Value) ValueType {
	return v.Type
}

// GetNumber 获取JSON数字值
//
// 当且仅当值类型为NUMBER时，返回值有效
func GetNumber(v *Value) float64 {
	return v.N
}

// GetString 获取JSON字符串值
//
// 当且仅当值类型为STRING时，返回值有效
func GetString(v *Value) string {
	return v.S
}

// GetStringLength 获取JSON字符串长度
//
// 当且仅当值类型为STRING时，返回值有效
func GetStringLength(v *Value) int {
	return len(v.S)
}

// SetString 设置JSON值为字符串
func SetString(v *Value, s string) {
	v.Type = STRING
	v.S = s
}

// parseWhitespace 跳过空白字符
func parseWhitespace(c *context) {
	for c.index < len(c.json) {
		if c.json[c.index] == ' ' || c.json[c.index] == '\t' || c.json[c.index] == '\n' || c.json[c.index] == '\r' {
			c.index++
		} else {
			break
		}
	}
}

// parseNull 解析null值
func parseNull(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) ||
		c.json[c.index+1] != 'u' ||
		c.json[c.index+2] != 'l' ||
		c.json[c.index+3] != 'l' {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = NULL
	return PARSE_OK
}

// parseTrue 解析true值
func parseTrue(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) ||
		c.json[c.index+1] != 'r' ||
		c.json[c.index+2] != 'u' ||
		c.json[c.index+3] != 'e' {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = TRUE
	return PARSE_OK
}

// parseFalse 解析false值
func parseFalse(c *context, v *Value) ParseError {
	if c.index+4 >= len(c.json) ||
		c.json[c.index+1] != 'a' ||
		c.json[c.index+2] != 'l' ||
		c.json[c.index+3] != 's' ||
		c.json[c.index+4] != 'e' {
		return PARSE_INVALID_VALUE
	}
	c.index += 5
	v.Type = FALSE
	return PARSE_OK
}

// parseNumber 解析数字值
func parseNumber(c *context, v *Value) ParseError {
	// 记录起始位置
	start := c.index

	// 处理负号
	if c.index < len(c.json) && c.json[c.index] == '-' {
		c.index++
	}

	// 检查十六进制数字
	if c.json[0] == '0' && len(c.json) > 1 {
		if c.json[1] == 'x' || c.json[1] == 'X' {
			return PARSE_INVALID_VALUE
		}
	}

	// 处理整数部分
	if c.index < len(c.json) && c.json[c.index] == '0' {
		c.index++
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			return PARSE_INVALID_VALUE
		}
	} else if c.index < len(c.json) && c.json[c.index] >= '1' && c.json[c.index] <= '9' {
		c.index++
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 处理小数部分
	if c.index < len(c.json) && c.json[c.index] == '.' {
		c.index++
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 处理指数部分
	if c.index < len(c.json) && (c.json[c.index] == 'e' || c.json[c.index] == 'E') {
		c.index++
		if c.index < len(c.json) && (c.json[c.index] == '+' || c.json[c.index] == '-') {
			c.index++
		}
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 转换为数字
	n, err := strconv.ParseFloat(c.json[start:c.index], 64)
	if err != nil {
		if numError, ok := err.(*strconv.NumError); ok {
			if numError.Err == strconv.ErrRange {
				return PARSE_NUMBER_TOO_BIG
			}
		}
		return PARSE_INVALID_VALUE
	}

	// 检查是否溢出
	if math.IsInf(n, 0) || math.IsNaN(n) {
		return PARSE_NUMBER_TOO_BIG
	}

	v.N = n
	v.Type = NUMBER
	return PARSE_OK
}

// parseValue 解析JSON值
func parseValue(c *context, v *Value) ParseError {
	if c.index >= len(c.json) {
		return PARSE_EXPECT_VALUE
	}

	switch c.json[c.index] {
	case 'n':
		return parseNull(c, v)
	case 't':
		return parseTrue(c, v)
	case 'f':
		return parseFalse(c, v)
	case '"':
		return parseString(c, v)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return parseNumber(c, v)
	default:
		return PARSE_INVALID_VALUE
	}
}

// parseString 解析字符串值
func parseString(c *context, v *Value) ParseError {
	return parseStringRaw(c, &v.S, v)
}

// parseStringRaw 解析字符串，不设置值类型
func parseStringRaw(c *context, str *string, v *Value) ParseError {
	if c.index >= len(c.json) || c.json[c.index] != '"' {
		return PARSE_MISS_QUOTATION_MARK
	}
	c.index++ // 跳过开始的引号

	var sb strings.Builder
	for c.index < len(c.json) {
		ch := c.json[c.index]
		if ch == '"' {
			c.index++ // 跳过结束的引号
			*str = sb.String()
			if v != nil {
				v.Type = STRING
			}
			return PARSE_OK
		}

		if ch == '\\' { // 处理转义字符
			c.index++
			if c.index >= len(c.json) {
				return PARSE_INVALID_STRING_ESCAPE
			}

			switch c.json[c.index] {
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			case '/':
				sb.WriteByte('/')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'u':
				if err := parseUnicode(c, &sb); err != PARSE_OK {
					return err
				}
				continue // parseUnicode已经处理了index的增加
			default:
				return PARSE_INVALID_STRING_ESCAPE
			}
		} else if ch < 0x20 { // 不允许未转义的控制字符
			return PARSE_INVALID_STRING_CHAR
		} else {
			sb.WriteByte(ch) // 普通字符直接写入
		}
		c.index++
	}
	return PARSE_MISS_QUOTATION_MARK // 字符串没有结束引号
}

// parseUnicode 解析 Unicode 转义序列
func parseUnicode(c *context, sb *strings.Builder) ParseError {
	// 确保有足够的字符
	if c.index+4 >= len(c.json) {
		return PARSE_INVALID_UNICODE_HEX
	}

	// 解析4位十六进制数字
	var u rune = 0
	c.index++ // 跳过 'u'

	for i := 0; i < 4; i++ {
		ch := c.json[c.index+i]
		u <<= 4
		switch {
		case ch >= '0' && ch <= '9':
			u |= rune(ch - '0')
		case ch >= 'A' && ch <= 'F':
			u |= rune(ch - 'A' + 10)
		case ch >= 'a' && ch <= 'f':
			u |= rune(ch - 'a' + 10)
		default:
			return PARSE_INVALID_UNICODE_HEX
		}
	}

	c.index += 4 // 移动索引越过这4个十六进制数字

	// 处理代理对
	if u >= 0xD800 && u <= 0xDBFF { // 高代理项
		if c.index+6 >= len(c.json) || // 确保后面有足够的字符
			c.json[c.index] != '\\' ||
			c.json[c.index+1] != 'u' {
			return PARSE_INVALID_UNICODE_SURROGATE
		}

		// 解析低代理项
		var u2 rune = 0
		c.index += 2 // 跳过 \u

		for i := 0; i < 4; i++ {
			ch := c.json[c.index+i]
			u2 <<= 4
			switch {
			case ch >= '0' && ch <= '9':
				u2 |= rune(ch - '0')
			case ch >= 'A' && ch <= 'F':
				u2 |= rune(ch - 'A' + 10)
			case ch >= 'a' && ch <= 'f':
				u2 |= rune(ch - 'a' + 10)
			default:
				return PARSE_INVALID_UNICODE_HEX
			}
		}
		c.index += 4

		if u2 < 0xDC00 || u2 > 0xDFFF { // 验证低代理项范围
			return PARSE_INVALID_UNICODE_SURROGATE
		}

		// 计算实际的 Unicode 码点
		u = 0x10000 + (u-0xD800)*0x400 + (u2 - 0xDC00)
	}

	// 将 Unicode 码点转换为 UTF-8 并写入缓冲区
	sb.WriteRune(u)
	return PARSE_OK
}

func isValidUnicode(s string) bool {
	if len(s) != 4 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
