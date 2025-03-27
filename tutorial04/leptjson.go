// leptjson.go - Go语言版JSON库实现
package leptjson

import (
	"fmt"
	"math"
	"strconv"
	"unicode/utf8"
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
	PARSE_INVALID_UNICODE_HEX                         // 无效的Unicode十六进制
	PARSE_INVALID_UNICODE_SURROGATE                   // 无效的Unicode代理对
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
		return "无效的Unicode十六进制"
	case PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的Unicode代理对"
	default:
		return fmt.Sprintf("未知错误(%d)", int(e))
	}
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
func GetNumber(v *Value) float64 {
	return v.N
}

// GetString 获取JSON字符串值
func GetString(v *Value) string {
	return v.S
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

	// 处理整数部分
	if c.index < len(c.json) && c.json[c.index] == '0' {
		c.index++
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) {
			ch := c.json[c.index]
			// 检查0后面是否跟着x或X（十六进制表示法）
			if ch == 'x' || ch == 'X' {
				return PARSE_INVALID_VALUE
			}
			if ch >= '0' && ch <= '9' {
				return PARSE_INVALID_VALUE
			}
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

// parseString 解析字符串值
func parseString(c *context, v *Value) ParseError {
	return parseStringRaw(c, &v.S, v)
}

// parseStringRaw 解析字符串，可选择是否设置值
func parseStringRaw(c *context, s *string, v *Value) ParseError {
	var sb []byte
	c.index++ // 跳过开头的引号

	for c.index < len(c.json) {
		ch := c.json[c.index]
		if ch == '"' {
			c.index++
			if v != nil {
				v.Type = STRING
			}
			*s = string(sb)
			return PARSE_OK
		} else if ch == '\\' {
			c.index++
			if c.index >= len(c.json) {
				return PARSE_INVALID_STRING_ESCAPE
			}

			switch c.json[c.index] {
			case '"', '\\', '/':
				sb = append(sb, c.json[c.index])
			case 'b':
				sb = append(sb, '\b')
			case 'f':
				sb = append(sb, '\f')
			case 'n':
				sb = append(sb, '\n')
			case 'r':
				sb = append(sb, '\r')
			case 't':
				sb = append(sb, '\t')
			case 'u':
				// 解析4位十六进制数字
				if c.index+4 >= len(c.json) {
					return PARSE_INVALID_UNICODE_HEX
				}

				var codepoint int
				for i := 1; i <= 4; i++ {
					c.index++
					ch := c.json[c.index]
					codepoint <<= 4
					if ch >= '0' && ch <= '9' {
						codepoint |= int(ch - '0')
					} else if ch >= 'A' && ch <= 'F' {
						codepoint |= int(ch - 'A' + 10)
					} else if ch >= 'a' && ch <= 'f' {
						codepoint |= int(ch - 'a' + 10)
					} else {
						return PARSE_INVALID_UNICODE_HEX
					}
				}

				// 处理UTF-16代理对
				if codepoint >= 0xD800 && codepoint <= 0xDBFF {
					// 高代理项，需要后面跟着低代理项
					if c.index+6 >= len(c.json) ||
						c.json[c.index+1] != '\\' ||
						c.json[c.index+2] != 'u' {
						return PARSE_INVALID_UNICODE_SURROGATE
					}

					c.index += 2 // 跳过 \u
					var codepoint2 int
					for i := 1; i <= 4; i++ {
						c.index++
						ch := c.json[c.index]
						codepoint2 <<= 4
						if ch >= '0' && ch <= '9' {
							codepoint2 |= int(ch - '0')
						} else if ch >= 'A' && ch <= 'F' {
							codepoint2 |= int(ch - 'A' + 10)
						} else if ch >= 'a' && ch <= 'f' {
							codepoint2 |= int(ch - 'a' + 10)
						} else {
							return PARSE_INVALID_UNICODE_HEX
						}
					}

					// 检查是否为有效的低代理项
					if codepoint2 < 0xDC00 || codepoint2 > 0xDFFF {
						return PARSE_INVALID_UNICODE_SURROGATE
					}

					// 计算Unicode码点
					codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (codepoint2 - 0xDC00)
				}

				// 将Unicode码点编码为UTF-8
				buf := make([]byte, 4)
				n := utf8.EncodeRune(buf, rune(codepoint))
				sb = append(sb, buf[:n]...)

			default:
				return PARSE_INVALID_STRING_ESCAPE
			}
		} else if ch < 0x20 {
			// 控制字符必须转义
			return PARSE_INVALID_STRING_CHAR
		} else {
			sb = append(sb, ch)
		}
		c.index++
	}

	return PARSE_MISS_QUOTATION_MARK
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
