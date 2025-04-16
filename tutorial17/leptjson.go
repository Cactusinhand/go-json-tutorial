// leptjson.go - Go语言版JSON库
package leptjson

import (
	// "math"

	"bytes"
	"fmt"
	"reflect" // 引入 reflect 包
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
	PARSE_MAX_DEPTH_EXCEEDED                             // 超过最大嵌套深度
	PARSE_COMMENT_NOT_CLOSED                             // 注释未闭合
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

// Parse 解析JSON文本（使用默认选项）
func Parse(v *Value, json string) ParseError {
	return ParseWithOptions(v, json, DefaultParseOptions())
}

// ParseWithOptions 使用自定义选项解析JSON文本
//
// 解析步骤：
// 1. 检查输入大小（安全检查）
// 2. 跳过前导空白字符
// 3. 解析JSON值
// 4. 跳过后续空白字符
// 5. 检查是否还有额外内容（这将导致PARSE_ROOT_NOT_SINGULAR错误）
func ParseWithOptions(v *Value, json string, options ParseOptions) ParseError {
	c := newContext(json, options)
	v.Type = NULL // 初始化为NULL类型

	// 检查输入总大小（安全检查）
	if ok, errInfo := c.checkTotalSize(); !ok {
		return errInfo.Code
	}

	// 跳过空白字符和注释
	c.parseWhitespace()

	// 解析JSON值
	err := parseValue(c, v)
	if err != PARSE_OK {
		// 返回解析错误
		return err
	}

	// 跳过尾部空白字符和注释
	c.parseWhitespace()

	// 检查是否有多余内容
	if c.index < len(c.json) {
		return PARSE_ROOT_NOT_SINGULAR
	}

	return PARSE_OK
}

// parseNull 解析null值
//
// 检查是否匹配"null"字符串，并设置值类型为NULL
func parseNull(c *parseContext, v *Value) ParseError {
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
func parseTrue(c *parseContext, v *Value) ParseError {
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
func parseFalse(c *parseContext, v *Value) ParseError {
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
func parseNumber(c *parseContext, v *Value) ParseError {
	startIndex := c.index

	// 处理负号
	if c.peekChar() == '-' {
		c.nextChar()
	}

	// 整数部分
	if c.peekChar() == '0' {
		c.nextChar()
		// 0后面不能直接跟数字，必须是小数点或指数符号
		if c.index < len(c.json) {
			ch := c.peekChar()
			// 检查0后面是否跟着x或X（十六进制表示法）或数字，都是非法的
			if (ch >= '0' && ch <= '9') || ch == 'x' || ch == 'X' {
				return PARSE_INVALID_VALUE
			}
		}
	} else if c.index < len(c.json) && c.peekChar() >= '1' && c.peekChar() <= '9' {
		c.nextChar()
		for c.index < len(c.json) && c.peekChar() >= '0' && c.peekChar() <= '9' {
			c.nextChar()
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 小数部分
	if c.index < len(c.json) && c.peekChar() == '.' {
		c.nextChar()
		if c.index >= len(c.json) || c.peekChar() < '0' || c.peekChar() > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.peekChar() >= '0' && c.peekChar() <= '9' {
			c.nextChar()
		}
	}

	// 指数部分
	if c.index < len(c.json) && (c.peekChar() == 'e' || c.peekChar() == 'E') {
		c.nextChar()
		if c.index < len(c.json) && (c.peekChar() == '+' || c.peekChar() == '-') {
			c.nextChar()
		}
		if c.index >= len(c.json) || c.peekChar() < '0' || c.peekChar() > '9' {
			return PARSE_INVALID_VALUE
		}
		for c.index < len(c.json) && c.peekChar() >= '0' && c.peekChar() <= '9' {
			c.nextChar()
		}
	}

	// 解析完成，将JSON文本中的数字子串转换为浮点数
	numStr := c.json[startIndex:c.index]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		// 可能是数字太大等原因导致的转换失败
		return PARSE_NUMBER_TOO_BIG
	}

	// 只有对实际的数值解析时才进行数值范围安全检查，不对嵌套的数据结构进行此检查
	if c.options.EnabledSecurity {
		// 直接进行检查而不通过checkNumberRange，避免产生PARSE_NUMBER_RANGE_EXCEEDED错误
		if num > c.options.MaxNumberValue || num < c.options.MinNumberValue {
			return PARSE_NUMBER_RANGE_EXCEEDED
		}
	}

	v.Type = NUMBER
	v.N = num
	return PARSE_OK
}

// parseString 解析字符串值
//
// 解析双引号包围的字符串并处理转义序列
func parseString(c *parseContext, v *Value) ParseError {
	var sb strings.Builder
	err := parseStringRaw(c, nil, &sb)
	if err != PARSE_OK {
		return err
	}

	// 获取解析后的字符串值
	result := sb.String()

	// 添加字符串长度安全检查
	if ok, errInfo := c.checkStringLength(len(result)); !ok {
		return errInfo.Code
	}

	v.Type = STRING
	v.S = result
	return PARSE_OK
}

// parseStringRaw 解析原始字符串
//
// 当s不为nil时，直接将解析结果存入s而不使用sb
// 当s为nil时，使用sb构建字符串
func parseStringRaw(c *parseContext, s *string, sb *strings.Builder) ParseError {
	// 确保是以引号开始的
	if c.peekChar() != '"' {
		return PARSE_MISS_QUOTATION_MARK
	}
	c.nextChar() // 跳过开始的双引号

	for c.index < len(c.json) {
		ch := c.nextChar() // 读取字符并前进
		switch ch {
		case '"': // 字符串结束
			if s != nil {
				*s = sb.String()
			}
			return PARSE_OK
		case '\\': // 转义序列
			if c.index >= len(c.json) {
				return PARSE_INVALID_STRING_ESCAPE
			}
			switch c.peekChar() {
			case '"', '\\':
				sb.WriteByte(c.nextChar())
			case 'b':
				c.nextChar()
				sb.WriteByte('\b')
			case 'f':
				c.nextChar()
				sb.WriteByte('\f')
			case 'n':
				c.nextChar()
				sb.WriteByte('\n')
			case 'r':
				c.nextChar()
				sb.WriteByte('\r')
			case 't':
				c.nextChar()
				sb.WriteByte('\t')
			case 'u': // Unicode
				c.nextChar() // 跳过'u'
				// 解析4位十六进制数字
				if c.index+3 >= len(c.json) {
					return PARSE_INVALID_UNICODE_HEX
				}

				// 解析高代理项（surrogate high）
				var codepoint uint32
				for i := 0; i < 4; i++ {
					ch := c.nextChar()
					codepoint <<= 4
					if ch >= '0' && ch <= '9' {
						codepoint |= uint32(ch - '0')
					} else if ch >= 'A' && ch <= 'F' {
						codepoint |= uint32(ch - 'A' + 10)
					} else if ch >= 'a' && ch <= 'f' {
						codepoint |= uint32(ch - 'a' + 10)
					} else {
						return PARSE_INVALID_UNICODE_HEX
					}
				}

				// 检查是否是代理对的高位部分
				if codepoint >= 0xD800 && codepoint <= 0xDBFF {
					// 确保后面跟着低代理项
					if c.index+5 >= len(c.json) {
						return PARSE_INVALID_UNICODE_SURROGATE
					}
					if c.nextChar() != '\\' || c.nextChar() != 'u' {
						return PARSE_INVALID_UNICODE_SURROGATE
					}

					// 解析低代理项（surrogate low）
					var lowSurrogate uint32
					for i := 0; i < 4; i++ {
						ch := c.nextChar()
						lowSurrogate <<= 4
						if ch >= '0' && ch <= '9' {
							lowSurrogate |= uint32(ch - '0')
						} else if ch >= 'A' && ch <= 'F' {
							lowSurrogate |= uint32(ch - 'A' + 10)
						} else if ch >= 'a' && ch <= 'f' {
							lowSurrogate |= uint32(ch - 'a' + 10)
						} else {
							return PARSE_INVALID_UNICODE_HEX
						}
					}

					// 验证是否是有效的低代理项
					if lowSurrogate < 0xDC00 || lowSurrogate > 0xDFFF {
						return PARSE_INVALID_UNICODE_SURROGATE
					}

					// 组合高低代理项计算实际的Unicode码点
					codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (lowSurrogate - 0xDC00)
				}

				// 将Unicode码点编码为UTF-8
				if codepoint <= 0x7F {
					sb.WriteByte(byte(codepoint & 0xFF))
				} else if codepoint <= 0x7FF {
					sb.WriteByte(byte(0xC0 | ((codepoint >> 6) & 0xFF)))
					sb.WriteByte(byte(0x80 | (codepoint & 0x3F)))
				} else if codepoint <= 0xFFFF {
					sb.WriteByte(byte(0xE0 | ((codepoint >> 12) & 0xFF)))
					sb.WriteByte(byte(0x80 | ((codepoint >> 6) & 0x3F)))
					sb.WriteByte(byte(0x80 | (codepoint & 0x3F)))
				} else {
					sb.WriteByte(byte(0xF0 | ((codepoint >> 18) & 0xFF)))
					sb.WriteByte(byte(0x80 | ((codepoint >> 12) & 0x3F)))
					sb.WriteByte(byte(0x80 | ((codepoint >> 6) & 0x3F)))
					sb.WriteByte(byte(0x80 | (codepoint & 0x3F)))
				}
			default:
				return PARSE_INVALID_STRING_ESCAPE
			}
		case 0: // 这里用数字0代替'\0'，避免非法字符错误
			return PARSE_MISS_QUOTATION_MARK
		default:
			// 控制字符（ASCII码小于0x20的字符）必须使用转义表示
			if ch < 0x20 {
				return PARSE_INVALID_STRING_CHAR
			}
			sb.WriteByte(ch)
		}
	}

	// 如果执行到这里，说明字符串没有正常结束（缺少结束引号）
	return PARSE_MISS_QUOTATION_MARK
}

// parseArray 解析数组值
//
// 解析形如 [value, value, ...] 的数组
func parseArray(c *parseContext, v *Value) ParseError {
	// 检查嵌套深度
	canNest, errInfo := c.enterNesting()
	if !canNest {
		return errInfo.Code
	}
	defer c.exitNesting()

	// 确保是以'['开始的
	if c.peekChar() != '[' {
		return PARSE_INVALID_VALUE
	}
	c.nextChar() // 跳过'['

	// 初始化数组元素计数和安全检查
	c.enterArray()
	defer c.exitArray()

	// 初始化为空数组
	v.Type = ARRAY
	v.A = make([]*Value, 0)

	// 跳过空白字符
	c.parseWhitespace()

	// 处理空数组情况
	if c.peekChar() == ']' {
		c.nextChar() // 跳过']'
		return PARSE_OK
	}

	// 循环解析数组元素
	for {
		// 创建一个新元素
		e := new(Value)

		// 解析元素值
		if err := parseValue(c, e); err != PARSE_OK {
			// 解析失败，释放已分配的内存
			v.Type = NULL
			v.A = nil
			return err
		}

		// 添加到数组中
		v.A = append(v.A, e)

		// 安全检查：添加数组元素
		if ok, errInfo := c.addArrayElement(); !ok {
			// 安全检查失败，释放已分配的内存
			v.Type = NULL
			v.A = nil
			return errInfo.Code
		}

		// 跳过空白字符
		c.parseWhitespace()

		// 检查是否到达数组结束或需要继续
		if c.peekChar() == ']' {
			c.nextChar() // 跳过']'
			return PARSE_OK
		} else if c.peekChar() == ',' {
			c.nextChar() // 跳过','
			c.parseWhitespace()
			// 允许尾随逗号的情况
			if c.options.AllowTrailing && c.peekChar() == ']' {
				c.nextChar() // 跳过']'
				return PARSE_OK
			}
		} else {
			// 既不是']'也不是','，语法错误
			v.Type = NULL
			v.A = nil
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}
	}
}

// parseObject 解析对象值
//
// 解析形如 {"key":value, "key":value, ...} 的对象
func parseObject(c *parseContext, v *Value) ParseError {
	// 检查嵌套深度
	canNest, errInfo := c.enterNesting()
	if !canNest {
		return errInfo.Code
	}
	defer c.exitNesting()

	// 确保是以'{'开始的
	if c.peekChar() != '{' {
		return PARSE_INVALID_VALUE
	}
	c.nextChar() // 跳过'{'

	// 初始化对象成员计数和安全检查
	c.enterObject()
	defer c.exitObject()

	// 初始化为空对象
	v.Type = OBJECT
	v.O = make([]Member, 0)

	// 跳过空白字符
	c.parseWhitespace()

	// 处理空对象情况
	if c.peekChar() == '}' {
		c.nextChar() // 跳过'}'
		return PARSE_OK
	}

	// 循环解析对象成员
	for {
		var m Member

		// 检查键是否以引号开始
		if c.peekChar() != '"' {
			v.Type = NULL
			v.O = nil
			return PARSE_MISS_KEY
		}

		// 解析成员的键
		var sb strings.Builder
		if err := parseStringRaw(c, &m.K, &sb); err != PARSE_OK {
			v.Type = NULL
			v.O = nil
			return err
		}

		// 跳过空白字符
		c.parseWhitespace()

		// 检查冒号
		if c.peekChar() != ':' {
			v.Type = NULL
			v.O = nil
			return PARSE_MISS_COLON
		}
		c.nextChar() // 跳过':'

		// 跳过冒号后的空白字符
		c.parseWhitespace()

		// 解析成员的值
		m.V = new(Value)
		if err := parseValue(c, m.V); err != PARSE_OK {
			v.Type = NULL
			v.O = nil
			return err
		}

		// 添加到对象中
		v.O = append(v.O, m)

		// 安全检查：添加对象成员
		if ok, errInfo := c.addObjectMember(); !ok {
			// 安全检查失败，释放已分配的内存
			v.Type = NULL
			v.O = nil
			return errInfo.Code
		}

		// 跳过空白字符
		c.parseWhitespace()

		// 检查是否结束或需要继续
		if c.peekChar() == '}' {
			c.nextChar() // 跳过'}'
			return PARSE_OK
		} else if c.peekChar() == ',' {
			c.nextChar() // 跳过','
			c.parseWhitespace()
			// 允许尾随逗号的情况
			if c.options.AllowTrailing && c.peekChar() == '}' {
				c.nextChar() // 跳过'}'
				return PARSE_OK
			}
		} else {
			// 既不是'}'也不是','，语法错误
			v.Type = NULL
			v.O = nil
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
func parseValue(c *parseContext, v *Value) ParseError {
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
	case PARSE_MAX_DEPTH_EXCEEDED:
		return "超过最大嵌套深度"
	case PARSE_COMMENT_NOT_CLOSED:
		return "注释未闭合"
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

// 导出的API函数 - JSON指针

// GetValueByPointer 使用JSON指针获取值
func GetValueByPointer(v *Value, pointerStr string) (*Value, error) {
	pointer, err := ParseJSONPointer(pointerStr)
	if err != POINTER_OK {
		return nil, err
	}
	return pointer.Get(v)
}

// SetValueByPointer 使用JSON指针设置值
func SetValueByPointer(v *Value, pointerStr string, value *Value) error {
	pointer, err := ParseJSONPointer(pointerStr)
	if err != POINTER_OK {
		return err
	}
	return pointer.Insert(v, value)
}

// RemoveValueByPointer 使用JSON指针删除值
func RemoveValueByPointer(v *Value, pointerStr string) error {
	pointer, err := ParseJSONPointer(pointerStr)
	if err != POINTER_OK {
		return err
	}
	return pointer.Remove(v)
}

// BuildJSONPointer 创建一个JSON指针字符串
func BuildJSONPointer(segments ...interface{}) (string, error) {
	pointer, err := GetJSONPointer(segments...)
	if err != nil {
		return "", err
	}
	return pointer.String(), nil
}

// 导出的API函数 - 循环引用检测

// HasCycle 检测JSON值中是否存在循环引用
func HasCycle(v *Value) bool {
	return DetectCycle(v) == CYCLE_DETECTED
}

// CopySafe 安全复制JSON值，避免循环引用
func CopySafe(dst, src *Value) error {
	return SafeCopy(dst, src)
}

// DefaultCircularReplacer 默认循环引用替换器
func DefaultCircularReplacer(path []string) *Value {
	// 创建一个表示循环引用的字符串值
	v := &Value{}
	SetString(v, fmt.Sprintf("循环引用 -> /%s", strings.Join(path, "/")))
	return v
}

// CopySafeWithReplacement 带替换的安全复制
func CopySafeWithReplacement(dst, src *Value) {
	SafeCopyWithReplacer(dst, src, DefaultCircularReplacer)
}

// CustomCopySafeWithReplacement 使用自定义替换器的安全复制
func CustomCopySafeWithReplacement(dst, src *Value, replacer CircularReplacer) {
	SafeCopyWithReplacer(dst, src, replacer)
}

// SetNull 将值设置为NULL类型
func SetNull(v *Value) {
	Free(v) // 释放可能存在的资源
	v.Type = NULL
}

// Marshal 将 Go 值序列化为 JSON 字符串，支持 struct tag。
func Marshal(v interface{}) (string, error) {
	value, err := marshalToValue(reflect.ValueOf(v))
	if err != nil {
		return "", err
	}
	// return Stringify(value) // 旧的错误返回方式
	s, stringifyErr := Stringify(value) // 接收两个返回值
	if stringifyErr != STRINGIFY_OK {   // 检查错误码
		return "", stringifyErr // 返回 StringifyError (它实现了 error 接口)
	}
	return s, nil // 成功时返回 nil error
}

// marshalToValue 是 Marshal 的核心递归函数
// 它将 reflect.Value 转换为 *leptjson.Value
func marshalToValue(rv reflect.Value) (*Value, error) {
	// 处理无效值
	if !rv.IsValid() {
		return &Value{Type: NULL}, nil
	}

	// 处理指针和接口
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return &Value{Type: NULL}, nil
		}
		rv = rv.Elem() // 解引用或获取接口的动态值
		// 再次检查解引用后的值是否有效
		if !rv.IsValid() {
			return &Value{Type: NULL}, nil
		}
	}

	// 根据类型处理
	switch rv.Kind() {
	case reflect.Invalid:
		return &Value{Type: NULL}, nil
	case reflect.Bool:
		val := &Value{}
		SetBoolean(val, rv.Bool())
		return val, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val := &Value{}
		SetNumber(val, float64(rv.Int())) // 转换为 float64
		return val, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		val := &Value{}
		SetNumber(val, float64(rv.Uint())) // 转换为 float64
		return val, nil
	case reflect.Float32, reflect.Float64:
		val := &Value{}
		SetNumber(val, rv.Float())
		return val, nil
	case reflect.String:
		val := &Value{}
		SetString(val, rv.String())
		return val, nil
	case reflect.Slice, reflect.Array:
		// 处理数组/切片 (稍后实现)
		return marshalArrayToValue(rv)
	case reflect.Map:
		// 处理映射 (稍后实现)
		return marshalMapToValue(rv)
	case reflect.Struct:
		// 处理结构体 (稍后实现)
		return marshalStructToValue(rv)
	default:
		return nil, fmt.Errorf("不支持的 Go 类型进行 Marshal: %s", rv.Kind())
	}
}

// marshalArrayToValue 将 Go 数组/切片转换为 *leptjson.Value (ARRAY)
func marshalArrayToValue(rv reflect.Value) (*Value, error) {
	arrVal := &Value{}
	SetArray(arrVal, rv.Len()) // 初始化数组并设置容量

	for i := 0; i < rv.Len(); i++ {
		elemRv := rv.Index(i)
		elemVal, err := marshalToValue(elemRv) // 递归转换元素
		if err != nil {
			// 清理已添加的元素？暂时不处理，直接返回错误
			return nil, fmt.Errorf("序列化数组元素 %d 失败: %w", i, err)
		}
		// 直接追加，因为 SetArray 初始化了空的 slice
		arrVal.A = append(arrVal.A, elemVal)
	}
	return arrVal, nil
}

// marshalMapToValue 将 Go map[string]interface{} 转换为 *leptjson.Value (OBJECT)
func marshalMapToValue(rv reflect.Value) (*Value, error) {
	// 检查 map key 类型是否为 string
	if rv.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("只支持 string 类型的 map key 进行 Marshal")
	}

	objVal := &Value{}
	SetObject(objVal) // 初始化对象

	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key().String() // 获取 string key
		valueRv := iter.Value()    // 获取 value (reflect.Value)

		mapValue, err := marshalToValue(valueRv) // 递归转换 map value
		if err != nil {
			// 清理？暂时不处理
			return nil, fmt.Errorf("序列化 map 值失败 (key: %s): %w", key, err)
		}
		// 添加到对象
		valuePtr := SetObjectValue(objVal, key)
		Copy(valuePtr, mapValue)
	}
	return objVal, nil
}

// marshalStructToValue 将 Go struct 转换为 *leptjson.Value (OBJECT)
func marshalStructToValue(rv reflect.Value) (*Value, error) {
	objVal := &Value{}
	SetObject(objVal)
	t := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := t.Field(i)       // 获取字段类型信息 (StructField)
		fieldValue := rv.Field(i) // 获取字段值 (Value)

		// 跳过非导出字段
		if field.PkgPath != "" {
			continue
		}

		// 解析 json tag
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue // 忽略带 "-" tag 的字段
		}

		// 解析 tag 选项 (omitempty)
		tagParts := strings.Split(tag, ",")
		jsonKey := tagParts[0]
		omitempty := false
		if len(tagParts) > 1 && tagParts[1] == "omitempty" {
			omitempty = true
		}

		// 如果 tag 没有指定 key name，使用字段名
		if jsonKey == "" {
			jsonKey = field.Name
		}

		// 处理 omitempty
		if omitempty && isEmptyValue(fieldValue) {
			continue
		}

		// 递归转换字段值
		structFieldValue, err := marshalToValue(fieldValue)
		if err != nil {
			return nil, fmt.Errorf("序列化 struct 字段 '%s' 失败: %w", field.Name, err)
		}

		// 添加到对象
		valuePtr := SetObjectValue(objVal, jsonKey)
		Copy(valuePtr, structFieldValue)
	}

	return objVal, nil
}

// isEmptyValue 检查 reflect.Value 是否为其类型的零值
// 这是 omitempty 的简化实现
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	// 对于 struct 等其他类型，omitempty 通常不适用或行为更复杂
	// 这里简化处理，认为非上述类型不为空
	return false
}
