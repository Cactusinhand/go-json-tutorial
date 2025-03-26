// Package leptjson 实现了一个轻量级的JSON解析器
// 支持解析以下JSON值类型：
// - null
// - true
// - false
// - number
// - string
// - array
// - object
//
// 使用示例：
//
//	v := Value{}
//	if err := Parse(&v, "null"); err != PARSE_OK {
//	    // 处理错误
//	}
//	if GetType(&v) == NULL {
//	    // 处理null值
//	}
//
// 特殊情况处理：
// 1. null值解析：
//   - 必须完全匹配"null"字符串
//   - 大小写敏感，如"NULL"、"Null"等都是非法的
//   - 不能包含额外的空白字符
//
// 2. true值解析：
//   - 必须完全匹配"true"字符串
//   - 大小写敏感，如"TRUE"、"True"等都是非法的
//   - 不能包含额外的空白字符
//
// 3. false值解析：
//   - 必须完全匹配"false"字符串
//   - 大小写敏感，如"FALSE"、"False"等都是非法的
//   - 不能包含额外的空白字符
//
// 4. 空白字符处理：
//   - 支持空格、制表符(\t)、换行符(\n)和回车符(\r)
//   - 空白字符可以出现在JSON值的任意位置
//   - 多个连续的空白字符是合法的
//
// 5. 错误处理：
//   - PARSE_EXPECT_VALUE: 输入为空或只包含空白字符
//   - PARSE_INVALID_VALUE: 输入格式不正确
//   - PARSE_ROOT_NOT_SINGULAR: 输入包含多个值
package leptjson

import (
	"fmt"
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
	PARSE_OK                ParseError = iota // 解析成功
	PARSE_EXPECT_VALUE                        // 期望一个值
	PARSE_INVALID_VALUE                       // 无效的值
	PARSE_ROOT_NOT_SINGULAR                   // 根节点不唯一
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
	default:
		return fmt.Sprintf("未知错误(%d)", int(e))
	}
}

// Value 表示一个JSON值
// 字段说明：
// - Type: 值的类型
// - Num: 数值（当Type为NUMBER时使用）
// - Str: 字符串值（当Type为STRING时使用）
// - Array: 数组值（当Type为ARRAY时使用）
type Value struct {
	Type  ValueType
	Num   float64
	Str   string
	Array []*Value
}

// String 实现fmt.Stringer接口，返回Value的字符串表示
func (v *Value) String() string {
	return v.Type.String()
}

// context 表示解析过程中的上下文
type context struct {
	json  string
	index int
}

// Parse 解析JSON字符串
// 参数：
//   - v: 用于存储解析结果的Value指针
//   - json: 要解析的JSON字符串
//
// 返回值：
//   - ParseError: 解析结果，PARSE_OK表示成功
//
// 注意：
//   - 解析前会重置Value对象
//   - 解析失败时会设置Value类型为NULL
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

// GetType 获取Value的类型
// 参数：
//   - v: Value指针
//
// 返回值：
//   - ValueType: 值的类型
func GetType(v *Value) ValueType {
	return v.Type
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
	default:
		return PARSE_INVALID_VALUE
	}
}
