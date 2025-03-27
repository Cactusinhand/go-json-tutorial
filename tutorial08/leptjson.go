// leptjson.go - Go语言版JSON库
package leptjson

import (
	// "math"

	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// ValueType 表示JSON值的类型
type ValueType int

// JSON值类型常量
const (
	NULL ValueType = iota
	FALSE
	TRUE
	NUMBER
	STRING
	ARRAY
	OBJECT
)

// ParseError 表示解析错误
type ParseError int

// 解析错误常量
const (
	PARSE_OK                           ParseError = iota // 解析成功
	PARSE_EXPECT_VALUE                                   // 期望一个值
	PARSE_INVALID_VALUE                                  // 无效的值
	PARSE_ROOT_NOT_SINGULAR                              // 根节点不唯一
	PARSE_NUMBER_TOO_BIG                                 // 数字太大
	PARSE_MISS_QUOTATION_MARK                            // 缺少引号
	PARSE_INVALID_STRING_ESCAPE                          // 无效的转义序列
	PARSE_INVALID_STRING_CHAR                            // 无效的字符
	PARSE_INVALID_UNICODE_HEX                            // 无效的Unicode十六进制
	PARSE_INVALID_UNICODE_SURROGATE                      // 无效的Unicode代理对
	PARSE_MISS_COMMA_OR_SQUARE_BRACKET                   // 缺少逗号或方括号
	PARSE_MISS_KEY                                       // 缺少键
	PARSE_MISS_COLON                                     // 缺少冒号
	PARSE_MISS_COMMA_OR_CURLY_BRACKET                    // 缺少逗号或花括号
)

// StringifyError 表示字符串化错误
type StringifyError int

// 字符串化错误常量
const (
	STRINGIFY_OK StringifyError = iota // 字符串化成功
)

// Member 表示对象的成员（键值对）
type Member struct {
	K string // 键
	V *Value // 值
}

// Value 表示一个JSON值
type Value struct {
	Type ValueType `json:"type"` // 值类型
	N    float64   `json:"n"`    // 数字值（当Type为NUMBER时有效）
	S    string    `json:"s"`    // 字符串值（当Type为STRING时有效）
	A    []*Value  `json:"a"`    // 数组值（当Type为ARRAY时有效）
	O    []Member  `json:"o"`    // 对象值（当Type为OBJECT时有效）
}

// String 返回Value的字符串表示
func (v Value) String() string {
	switch v.Type {
	case NULL:
		return "null"
	case FALSE:
		return "false"
	case TRUE:
		return "true"
	case NUMBER:
		return strconv.FormatFloat(v.N, 'f', -1, 64)
	case STRING:
		return "\"" + v.S + "\""
	case ARRAY:
		var sb strings.Builder
		sb.WriteString("[")
		for i, elem := range v.A {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(elem.String())
		}
		sb.WriteString("]")
		return sb.String()
	case OBJECT:
		var sb strings.Builder
		sb.WriteString("{")
		for i, member := range v.O {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString("\"")
			sb.WriteString(member.K)
			sb.WriteString("\":")
			sb.WriteString(member.V.String())
		}
		sb.WriteString("}")
		return sb.String()
	default:
		return "unknown"
	}
}

// context 表示解析上下文
type context struct {
	json  string // JSON文本
	index int    // 当前解析位置
}

// Parse 解析JSON文本
//
// 解析步骤：
// 1. 跳过前导空白字符
// 2. 解析JSON值
// 3. 跳过后续空白字符
// 4. 检查是否还有额外内容（这将导致PARSE_ROOT_NOT_SINGULAR错误）
func Parse(v *Value, json string) ParseError {
	c := context{json: json, index: 0}
	v.Type = NULL // 初始化为NULL类型
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

// parseWhitespace 跳过空白字符
//
// 根据JSON规范，空白字符包括：空格、制表符(\t)、换行符(\n)和回车符(\r)
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
//
// 检查是否匹配"null"字符串，并设置值类型为NULL
func parseNull(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) || c.json[c.index:c.index+4] != "null" {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = NULL
	return PARSE_OK
}

// parseTrue 解析true值
//
// 检查是否匹配"true"字符串，并设置值类型为TRUE
func parseTrue(c *context, v *Value) ParseError {
	if c.index+3 >= len(c.json) || c.json[c.index:c.index+4] != "true" {
		return PARSE_INVALID_VALUE
	}
	c.index += 4
	v.Type = TRUE
	return PARSE_OK
}

// parseFalse 解析false值
//
// 检查是否匹配"false"字符串，并设置值类型为FALSE
func parseFalse(c *context, v *Value) ParseError {
	if c.index+4 >= len(c.json) || c.json[c.index:c.index+5] != "false" {
		return PARSE_INVALID_VALUE
	}
	c.index += 5
	v.Type = FALSE
	return PARSE_OK
}

// parseNumber 解析数字值
//
// JSON数字语法规则：
// number = [ "-" ] int [ frac ] [ exp ]
// int = "0" / digit1-9 *digit
// frac = "." 1*digit
// exp = ("e" / "E") ["-" / "+"] 1*digit
//
// 特殊情况处理：
// 1. 不允许有前导0后跟数字（如"01"是非法的）
// 2. 不支持十六进制表示（如"0x1"是非法的）
// 3. 处理数字溢出情况
func parseNumber(c *context, v *Value) ParseError {
	startIndex := c.index
	if c.json[c.index] == '-' {
		c.index++
	}

	// 整数部分
	if c.index < len(c.json) && c.json[c.index] == '0' {
		c.index++
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) {
			ch := c.json[c.index]
			// 检查0后面是否跟着x或X（十六进制表示法）或数字，都是非法的
			if (ch >= '0' && ch <= '9') || ch == 'x' || ch == 'X' {
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

	// 小数部分
	if c.index < len(c.json) && c.json[c.index] == '.' {
		c.index++
		if c.index >= len(c.json) || c.json[c.index] < '0' || c.json[c.index] > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.json[c.index] >= '0' && c.json[c.index] <= '9' {
			c.index++
		}
	}

	// 指数部分
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

	// 转换为float64
	num, err := strconv.ParseFloat(c.json[startIndex:c.index], 64)
	if err != nil {
		if err.(*strconv.NumError).Err == strconv.ErrRange {
			return PARSE_NUMBER_TOO_BIG
		}
		return PARSE_INVALID_VALUE
	}

	v.Type = NUMBER
	v.N = num
	return PARSE_OK
}

// parseString 解析字符串值
//
// 将字符串解析结果设置到值中
func parseString(c *context, v *Value) ParseError {
	var s string
	if err := parseStringRaw(c, &s, nil); err != PARSE_OK {
		return err
	}
	v.Type = STRING
	v.S = s
	return PARSE_OK
}

// parseStringRaw 解析字符串，可选择将结果存储到s或直接追加到sb
//
// JSON字符串语法：
// string = quotation-mark *char quotation-mark
// char = unescaped / escape sequence
// escape = \"、\\、\/、\b、\f、\n、\r、\t、\uXXXX
//
// 处理方式：
// 1. 使用strings.Builder高效构建字符串
// 2. 处理所有标准转义字符
// 3. 支持Unicode代理对(\uXXXX\uYYYY)解析
// 4. 检查非法控制字符(小于0x20的字符)
// 5. 检查缺少右引号的情况
func parseStringRaw(c *context, s *string, sb *strings.Builder) ParseError {
	c.index++ // 跳过开头的引号
	var result strings.Builder
	for c.index < len(c.json) {
		ch := c.json[c.index]
		if ch == '"' {
			c.index++
			if s != nil {
				*s = result.String()
			}
			if sb != nil {
				sb.WriteString(result.String())
			}
			return PARSE_OK
		} else if ch == '\\' {
			c.index++
			if c.index >= len(c.json) {
				return PARSE_INVALID_STRING_ESCAPE
			}
			switch c.json[c.index] {
			case '"':
				result.WriteByte('"')
			case '\\':
				result.WriteByte('\\')
			case '/':
				result.WriteByte('/')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case 'u':
				if c.index+4 >= len(c.json) {
					return PARSE_INVALID_UNICODE_HEX
				}
				// 解析4位十六进制数字
				var codepoint int
				for i := 0; i < 4; i++ {
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
				// UTF-16代理对用于表示超出BMP（基本多语言平面）的Unicode字符
				// 高代理项：0xD800-0xDBFF，低代理项：0xDC00-0xDFFF
				// 解码公式：codepoint = 0x10000 + ((高代理 - 0xD800) << 10) + (低代理 - 0xDC00)
				if codepoint >= 0xD800 && codepoint <= 0xDBFF {
					// 高代理项，需要后面跟着低代理项
					if c.index+6 >= len(c.json) || c.json[c.index+1] != '\\' || c.json[c.index+2] != 'u' {
						return PARSE_INVALID_UNICODE_SURROGATE
					}
					c.index += 2 // 跳过 \u
					var codepoint2 int
					for i := 0; i < 4; i++ {
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
					if codepoint2 < 0xDC00 || codepoint2 > 0xDFFF {
						return PARSE_INVALID_UNICODE_SURROGATE
					}
					// 计算Unicode码点
					codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (codepoint2 - 0xDC00)
				} else if codepoint >= 0xDC00 && codepoint <= 0xDFFF {
					// 单独出现的低代理项是无效的
					return PARSE_INVALID_UNICODE_SURROGATE
				}
				// 将Unicode码点转换为UTF-8
				result.WriteRune(rune(codepoint))
			default:
				return PARSE_INVALID_STRING_ESCAPE
			}
		} else if ch < 0x20 {
			// 控制字符必须转义
			return PARSE_INVALID_STRING_CHAR
		} else {
			// 普通字符
			result.WriteByte(ch)
		}
		c.index++
	}
	return PARSE_MISS_QUOTATION_MARK
}

// parseArray 解析数组值
//
// JSON数组语法：
// array = [ [ value [ , value ] * ] ]
//
// 解析步骤：
// 1. 跳过'['和前导空白
// 2. 处理空数组情况
// 3. 循环解析数组元素，直到遇到']'或错误
// 4. 处理元素之间的逗号
// 5. 检查数组是否正确结束（带有']'）
//
// 特殊情况处理：
// 1. 处理空数组 []
// 2. 不允许数组末尾有多余的逗号，如[1,2,]
// 3. 检查未完成的数组，如[1,2（缺少右方括号）
// 4. 处理嵌套数组
func parseArray(c *context, v *Value) ParseError {
	c.index++ // 跳过开头的 [
	parseWhitespace(c)

	// 处理空数组的情况
	if c.index < len(c.json) && c.json[c.index] == ']' {
		c.index++
		v.Type = ARRAY
		v.A = make([]*Value, 0)
		return PARSE_OK
	}

	var elements []*Value

	for {
		// 解析数组元素
		element := &Value{}

		// parseValue 解析数组元素
		if err := parseValue(c, element); err != PARSE_OK {
			return err
		}
		elements = append(elements, element)

		parseWhitespace(c)
		if c.index >= len(c.json) {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}

		// 检查是否到达数组结束或需要继续解析
		if c.json[c.index] == ']' {
			c.index++
			v.Type = ARRAY
			v.A = elements
			return PARSE_OK
		} else if c.json[c.index] == ',' {
			c.index++
			parseWhitespace(c)

			// 关键修改：检查逗号后面是否已经到达字符串末尾
			if c.index >= len(c.json) {
				return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
			}

			// 关键修改：检查逗号后面是否直接是右方括号，这是不允许的
			if c.json[c.index] == ']' {
				return PARSE_INVALID_VALUE
			}
		} else {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}
	}
}

// parseObject 解析对象值
//
// JSON对象语法：
// object = { [ member [ , member ] * ] }
// member = string : value
//
// 解析步骤：
// 1. 跳过'{'和前导空白
// 2. 处理空对象情况
// 3. 循环解析每个键值对，直到遇到'}'或错误
// 4. 处理键值对之间的逗号
// 5. 检查对象是否正确结束（带有'}'）
//
// 特殊情况处理：
// 1. 处理空对象 {}
// 2. 确保对象的键是字符串
// 3. 检查键和值之间是否有冒号
// 4. 不允许对象末尾有多余的逗号，如{"a":1,}
// 5. 检查未完成的对象
func parseObject(c *context, v *Value) ParseError {
	c.index++ // 跳过开头的 {
	parseWhitespace(c)

	// 处理空对象的情况
	if c.index < len(c.json) && c.json[c.index] == '}' {
		c.index++
		v.Type = OBJECT
		v.O = make([]Member, 0)
		return PARSE_OK
	}

	var members []Member

	for {
		// 解析键（必须是字符串）
		if c.index >= len(c.json) || c.json[c.index] != '"' {
			return PARSE_MISS_KEY
		}

		var key string
		if err := parseStringRaw(c, &key, nil); err != PARSE_OK {
			return err
		}

		// 解析冒号
		parseWhitespace(c)
		if c.index >= len(c.json) || c.json[c.index] != ':' {
			return PARSE_MISS_COLON
		}
		c.index++
		parseWhitespace(c)

		// 解析值
		value := &Value{}
		if err := parseValue(c, value); err != PARSE_OK {
			return err
		}

		// 添加键值对
		members = append(members, Member{K: key, V: value})

		parseWhitespace(c)
		if c.index >= len(c.json) {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}

		// 检查是否到达对象结束或需要继续解析
		if c.json[c.index] == '}' {
			c.index++
			v.Type = OBJECT
			v.O = members
			return PARSE_OK
		} else if c.json[c.index] == ',' {
			c.index++
			parseWhitespace(c)
			// 检查逗号后面是否有内容
			if c.index >= len(c.json) {
				return PARSE_MISS_COMMA_OR_CURLY_BRACKET
			}
		} else {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}
	}
}

// parseValue 解析JSON值
//
// 根据当前字符确定JSON值的类型，并调用相应的解析函数
// JSON值可以是以下几种类型之一：
// - null: 以'n'开头
// - true: 以't'开头
// - false: 以'f'开头
// - string: 以'"'开头
// - array: 以'['开头
// - object: 以'{'开头
// - number: 以'-'或数字开头
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
	case '[':
		return parseArray(c, v)
	case '{':
		return parseObject(c, v)
	case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return parseNumber(c, v)
	default:
		return PARSE_INVALID_VALUE
	}
}

// GetType 获取JSON值的类型
func GetType(v *Value) ValueType {
	return v.Type
}

// GetBoolean 获取JSON布尔值
func GetBoolean(v *Value) bool {
	return v.Type == TRUE
}

// SetBoolean 设置JSON布尔值
func SetBoolean(v *Value, b bool) {
	if b {
		v.Type = TRUE
	} else {
		v.Type = FALSE
	}
}

// GetNumber 获取JSON数字值
func GetNumber(v *Value) float64 {
	return v.N
}

// SetNumber 设置JSON数字值
func SetNumber(v *Value, n float64) {
	v.Type = NUMBER
	v.N = n
}

// GetString 获取JSON字符串值
func GetString(v *Value) string {
	return v.S
}

// SetString 设置JSON字符串值
func SetString(v *Value, s string) {
	v.Type = STRING
	v.S = s
}

// GetArraySize 获取JSON数组的大小
func GetArraySize(v *Value) int {
	return len(v.A)
}

// GetArrayElement 获取JSON数组的元素
func GetArrayElement(v *Value, index int) *Value {
	if index < 0 || index >= len(v.A) {
		return nil
	}
	return v.A[index]
}

// GetObjectSize 获取JSON对象的大小
func GetObjectSize(v *Value) int {
	return len(v.O)
}

// GetObjectKey 获取JSON对象的键
func GetObjectKey(v *Value, index int) string {
	if index < 0 || index >= len(v.O) {
		return ""
	}
	return v.O[index].K
}

// GetObjectValue 获取JSON对象的值
func GetObjectValue(v *Value, index int) *Value {
	if index < 0 || index >= len(v.O) {
		return nil
	}
	return v.O[index].V
}

// FindObjectIndex 查找JSON对象中指定键的索引
func FindObjectIndex(v *Value, key string) int {
	for i, member := range v.O {
		if member.K == key {
			return i
		}
	}
	return -1
}

// GetObjectValueByKey 根据键获取JSON对象的值
func GetObjectValueByKey(v *Value, key string) *Value {
	index := FindObjectIndex(v, key)
	if index == -1 {
		return nil
	}
	return v.O[index].V
}

// FindObjectKey 根据键名在对象中查找对应值，如果找到返回值和true，否则返回nil和false
func FindObjectKey(v *Value, key string) (*Value, bool) {
	// 检查是否为空或非对象类型
	if v == nil || v.Type != OBJECT {
		return nil, false
	}

	// 遍历对象成员查找匹配的键
	for i := 0; i < len(v.O); i++ {
		if v.O[i].K == key {
			return v.O[i].V, true
		}
	}

	// 未找到匹配的键
	return nil, false
}

// Stringify 将Value转换为JSON字符串
func Stringify(v *Value) (string, StringifyError) {
	if v == nil {
		return "", STRINGIFY_OK
	}

	var buffer bytes.Buffer
	stringifyValue(v, &buffer)
	return buffer.String(), STRINGIFY_OK
}

// stringifyValue 将Value写入Buffer
func stringifyValue(v *Value, buffer *bytes.Buffer) {
	switch v.Type {
	case NULL:
		buffer.WriteString("null")
	case TRUE:
		buffer.WriteString("true")
	case FALSE:
		buffer.WriteString("false")
	case NUMBER:
		// 使用 -1 精度以获得最短的表示形式
		buffer.WriteString(strconv.FormatFloat(v.N, 'g', -1, 64))
	case STRING:
		stringifyString(v.S, buffer)
	case ARRAY:
		stringifyArray(v, buffer)
	case OBJECT:
		stringifyObject(v, buffer)
	}
}

// stringifyString 将字符串写入Buffer，处理转义字符
func stringifyString(s string, buffer *bytes.Buffer) {
	buffer.WriteByte('"')
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '"':
			buffer.WriteString("\\\"")
		case '\\':
			buffer.WriteString("\\\\")
		case '/':
			buffer.WriteString("\\/")
		case '\b':
			buffer.WriteString("\\b")
		case '\f':
			buffer.WriteString("\\f")
		case '\n':
			buffer.WriteString("\\n")
		case '\r':
			buffer.WriteString("\\r")
		case '\t':
			buffer.WriteString("\\t")
		default:
			if ch < 0x20 {
				// 对于其他控制字符，使用 \u00xx 形式
				buffer.WriteString(fmt.Sprintf("\\u%04x", ch))
			} else {
				buffer.WriteByte(ch)
			}
		}
	}
	buffer.WriteByte('"')
}

// stringifyArray 将数组写入Buffer
func stringifyArray(v *Value, buffer *bytes.Buffer) {
	buffer.WriteByte('[')
	for i, elem := range v.A {
		if i > 0 {
			buffer.WriteByte(',')
		}
		stringifyValue(elem, buffer)
	}
	buffer.WriteByte(']')
}

// stringifyObject 将对象写入Buffer
func stringifyObject(v *Value, buffer *bytes.Buffer) {
	buffer.WriteByte('{')
	for i, member := range v.O {
		if i > 0 {
			buffer.WriteByte(',')
		}
		stringifyString(member.K, buffer)
		buffer.WriteByte(':')
		stringifyValue(member.V, buffer)
	}
	buffer.WriteByte('}')
}

// Equal 判断两个JSON值是否相等
func Equal(lhs, rhs *Value) bool {
	// 首先检查指针是否相同
	if lhs == rhs {
		return true
	}

	// 检查两个值是否都为nil
	if lhs == nil || rhs == nil {
		return lhs == nil && rhs == nil
	}

	// 检查类型是否相同
	if lhs.Type != rhs.Type {
		return false
	}

	// 根据类型进行比较
	switch lhs.Type {
	case NULL, FALSE, TRUE:
		return true // 这些类型只要类型相同就相等
	case NUMBER:
		return lhs.N == rhs.N
	case STRING:
		return lhs.S == rhs.S
	case ARRAY:
		// 数组长度必须相同
		if len(lhs.A) != len(rhs.A) {
			return false
		}
		// 递归比较每个元素
		for i := 0; i < len(lhs.A); i++ {
			if !Equal(lhs.A[i], rhs.A[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		// 对象成员数必须相同
		if len(lhs.O) != len(rhs.O) {
			return false
		}

		// 对于对象，键值对的顺序可能不同，所以需要通过键来查找
		for _, m1 := range lhs.O {
			// 在rhs中查找对应的键
			found := false
			for _, m2 := range rhs.O {
				if m1.K == m2.K {
					found = true
					// 递归比较值
					if !Equal(m1.V, m2.V) {
						return false
					}
					break
				}
			}
			if !found {
				return false // rhs中没有找到m1的键
			}
		}
		return true
	default:
		return false
	}
}

// Copy 深度复制一个JSON值
func Copy(dst, src *Value) {
	if dst == nil || src == nil || dst == src {
		return
	}

	// 先释放目标值
	Free(dst)

	// 根据源值类型进行复制
	switch src.Type {
	case NULL:
		dst.Type = NULL
	case FALSE:
		dst.Type = FALSE
	case TRUE:
		dst.Type = TRUE
	case NUMBER:
		dst.Type = NUMBER
		dst.N = src.N
	case STRING:
		SetString(dst, src.S)
	case ARRAY:
		// 设置为数组类型并预分配空间
		SetArray(dst, len(src.A))
		// 递归复制每个元素
		for i := 0; i < len(src.A); i++ {
			element := &Value{}
			Copy(element, src.A[i])
			dst.A = append(dst.A, element)
		}
	case OBJECT:
		// 设置为对象类型并预分配空间
		SetObject(dst)
		// 递归复制每个成员
		for i := 0; i < len(src.O); i++ {
			v := &Value{}
			Copy(v, src.O[i].V)
			dst.O = append(dst.O, Member{K: src.O[i].K, V: v})
		}
	}
}

// Move 将源值移动到目标值，并将源值设为null
func Move(dst, src *Value) {
	if dst == nil || src == nil || dst == src {
		return
	}

	// 先释放目标值
	Free(dst)

	// 直接复制src的所有内容
	*dst = *src

	// 将src设为null
	src.Type = NULL
	src.A = nil // 确保数组引用被清空
	src.O = nil // 确保对象引用被清空
}

// Swap 交换两个JSON值
func Swap(lhs, rhs *Value) {
	if lhs == nil || rhs == nil || lhs == rhs {
		return
	}

	// 使用临时变量交换
	temp := *lhs
	*lhs = *rhs
	*rhs = temp
}

// SetArray 设置值为数组类型，可以预分配容量
func SetArray(v *Value, capacity int) {
	if v == nil {
		return
	}

	// 先释放原来的资源
	Free(v)

	// 设置为数组类型
	v.Type = ARRAY
	if capacity > 0 {
		v.A = make([]*Value, 0, capacity)
	} else {
		v.A = nil
	}
}

// SetObject 设置值为对象类型，可以预分配容量
func SetObject(v *Value) {
	if v == nil {
		return
	}

	// 先释放原来的资源
	Free(v)

	// 设置为对象类型
	v.Type = OBJECT
	v.O = []Member{}
}

// Free 释放JSON值占用的资源
func Free(v *Value) {
	if v == nil {
		return
	}

	switch v.Type {
	case STRING:
		// Go中字符串是不可变的，不需要手动释放内存
		v.S = ""
	case ARRAY:
		// 递归释放数组中的每个元素
		for i := 0; i < len(v.A); i++ {
			Free(v.A[i])
		}
		v.A = nil
	case OBJECT:
		// 递归释放对象中的每个值
		for i := 0; i < len(v.O); i++ {
			Free(v.O[i].V)
		}
		v.O = nil
	}

	// 将值类型设为NULL
	v.Type = NULL
}

// GetArrayCapacity 获取数组当前的容量
func GetArrayCapacity(v *Value) int {
	if v == nil || v.Type != ARRAY {
		return 0
	}
	return cap(v.A)
}

// ReserveArray 扩充数组容量
func ReserveArray(v *Value, capacity int) {
	if v == nil || v.Type != ARRAY || capacity <= cap(v.A) {
		return
	}

	// 创建新的更大容量的切片
	newArray := make([]*Value, len(v.A), capacity)
	// 复制原有元素
	copy(newArray, v.A)
	v.A = newArray
}

// ShrinkArray 缩小数组容量至实际大小
func ShrinkArray(v *Value) {
	if v == nil || v.Type != ARRAY || len(v.A) == cap(v.A) {
		return
	}

	// 创建新的切片，容量与长度相同
	newArray := make([]*Value, len(v.A))
	copy(newArray, v.A)
	v.A = newArray
}

// PushBackArrayElement 在数组末尾添加一个新元素，并返回该元素
func PushBackArrayElement(v *Value) *Value {
	if v == nil || v.Type != ARRAY {
		return nil
	}

	// 当容量不足时扩容
	if len(v.A) == cap(v.A) {
		// 使用自定义的max函数
		newCapacity := maxInt(1, cap(v.A)*2)
		ReserveArray(v, newCapacity)
	}

	// 创建新元素
	newElement := &Value{Type: NULL}
	v.A = append(v.A, newElement)
	return newElement
}

// PopBackArrayElement 移除数组末尾的元素
func PopBackArrayElement(v *Value) {
	if v == nil || v.Type != ARRAY || len(v.A) == 0 {
		return
	}

	// 获取最后一个元素并释放其资源
	lastIndex := len(v.A) - 1
	Free(v.A[lastIndex])

	// 截断数组
	v.A = v.A[:lastIndex]
}

// InsertArrayElement 在指定位置插入元素，并返回该元素
func InsertArrayElement(v *Value, index int) *Value {
	if v == nil || v.Type != ARRAY || index < 0 || index > len(v.A) {
		return nil
	}

	// 相当于在末尾追加
	if index == len(v.A) {
		return PushBackArrayElement(v)
	}

	// 当容量不足时扩容
	if len(v.A) == cap(v.A) {
		// 使用自定义的max函数
		newCapacity := maxInt(1, cap(v.A)*2)
		ReserveArray(v, newCapacity)
	}

	// 创建新元素
	newElement := &Value{Type: NULL}

	// 扩展数组并移动元素
	v.A = append(v.A, nil) // 追加一个nil占位
	// 从后向前移动元素
	for i := len(v.A) - 1; i > index; i-- {
		v.A[i] = v.A[i-1]
	}
	v.A[index] = newElement

	return newElement
}

// EraseArrayElement 删除数组中从index开始的count个元素
func EraseArrayElement(v *Value, index, count int) {
	if v == nil || v.Type != ARRAY || index < 0 || index >= len(v.A) || count <= 0 {
		return
	}

	// 调整count，确保不会超出数组范围
	if index+count > len(v.A) {
		count = len(v.A) - index
	}

	// 释放要删除的元素
	for i := 0; i < count; i++ {
		Free(v.A[index+i])
	}

	// 移动元素
	copy(v.A[index:], v.A[index+count:])

	// 调整数组大小
	v.A = v.A[:len(v.A)-count]
}

// ClearArray 清空数组的所有元素
func ClearArray(v *Value) {
	if v == nil || v.Type != ARRAY {
		return
	}

	// 释放所有元素
	for i := 0; i < len(v.A); i++ {
		Free(v.A[i])
	}

	// 清空数组但保留容量
	v.A = v.A[:0]
}

// GetObjectCapacity 获取对象的容量
func GetObjectCapacity(v *Value) int {
	if v == nil || v.Type != OBJECT {
		return 0
	}
	return cap(v.O)
}

// ReserveObject 扩充对象容量
func ReserveObject(v *Value, capacity int) {
	if v == nil || v.Type != OBJECT || capacity <= cap(v.O) {
		return
	}

	// 创建更大容量的切片
	newObject := make([]Member, len(v.O), capacity)
	copy(newObject, v.O)
	v.O = newObject
}

// ShrinkObject 缩小对象容量至实际大小
func ShrinkObject(v *Value) {
	if v == nil || v.Type != OBJECT || len(v.O) == cap(v.O) {
		return
	}

	// 创建新的切片，容量与长度相同
	newObject := make([]Member, len(v.O))
	copy(newObject, v.O)
	v.O = newObject
}

// SetObjectValue 设置对象的键值对，如果键已存在则返回其值指针，否则添加新的键值对并返回新值指针
func SetObjectValue(v *Value, key string) *Value {
	if v == nil || v.Type != OBJECT {
		return nil
	}

	// 先查找是否已存在该键
	for i := 0; i < len(v.O); i++ {
		if v.O[i].K == key {
			return v.O[i].V
		}
	}

	// 当容量不足时扩容
	if len(v.O) == cap(v.O) {
		newCapacity := maxInt(1, cap(v.O)*2)
		ReserveObject(v, newCapacity)
	}

	// 创建新值
	newValue := &Value{Type: NULL}

	// 添加新的键值对
	v.O = append(v.O, Member{K: key, V: newValue})

	return newValue
}

// RemoveObjectValue 移除对象中指定索引的成员
func RemoveObjectValue(v *Value, index int) {
	if v == nil || v.Type != OBJECT || index < 0 || index >= len(v.O) {
		return
	}

	// 释放值的资源
	Free(v.O[index].V)

	// 移动元素
	copy(v.O[index:], v.O[index+1:])

	// 调整对象大小
	v.O = v.O[:len(v.O)-1]
}

// ClearObject 清空对象的所有成员
func ClearObject(v *Value) {
	if v == nil || v.Type != OBJECT {
		return
	}

	// 释放所有值
	for i := 0; i < len(v.O); i++ {
		Free(v.O[i].V)
	}

	// 清空对象但保留容量
	v.O = v.O[:0]
}

// Error 返回解析错误的描述
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
	case PARSE_MISS_COMMA_OR_SQUARE_BRACKET:
		return "缺少逗号或方括号"
	case PARSE_MISS_KEY:
		return "缺少键"
	case PARSE_MISS_COLON:
		return "缺少冒号"
	case PARSE_MISS_COMMA_OR_CURLY_BRACKET:
		return "缺少逗号或花括号"
	default:
		return "未知错误"
	}
}

// Error 返回字符串化错误的描述
func (e StringifyError) Error() string {
	switch e {
	case STRINGIFY_OK:
		return "字符串化成功"
	default:
		return "未知错误"
	}
}

// maxInt 返回两个整数中的较大值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
