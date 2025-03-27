// leptjson.go - JSON 库的基础实现
package tutorial14

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// 类型常量
const (
	NULL   = iota // null
	FALSE         // false
	TRUE          // true
	NUMBER        // 数值
	STRING        // 字符串
	ARRAY         // 数组
	OBJECT        // 对象
)

// 解析结果常量
const (
	PARSE_OK = iota
	PARSE_EXPECT_VALUE
	PARSE_INVALID_VALUE
	PARSE_ROOT_NOT_SINGULAR
	PARSE_NUMBER_TOO_BIG
	PARSE_MISS_QUOTATION_MARK
	PARSE_INVALID_STRING_ESCAPE
	PARSE_INVALID_STRING_CHAR
	PARSE_INVALID_UNICODE_SURROGATE
	PARSE_INVALID_UNICODE_HEX
	PARSE_MISS_COMMA_OR_SQUARE_BRACKET
	PARSE_MISS_KEY
	PARSE_MISS_COLON
	PARSE_MISS_COMMA_OR_CURLY_BRACKET
)

// Member 表示对象中的一个键值对
type Member struct {
	K string // 键名
	V *Value // 值
}

// Value 表示一个 JSON 值
type Value struct {
	Type int       // 值类型
	N    float64   // 数值
	S    string    // 字符串
	A    []*Value  // 数组
	M    []*Member // 对象
}

// 字符分类函数
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isDigit1to9(ch byte) bool {
	return ch >= '1' && ch <= '9'
}

// 字符流解析器
type context struct {
	json  string // JSON 文本
	stack []byte // 字符缓冲区
	pos   int    // 当前位置
}

// 获取当前字符
func (c *context) currentChar() byte {
	if c.pos >= len(c.json) {
		return 0
	}
	return c.json[c.pos]
}

// 获取下一个字符
func (c *context) nextChar() byte {
	c.pos++
	return c.currentChar()
}

// 跳过空白字符
func (c *context) skipWhitespace() {
	for c.pos < len(c.json) {
		ch := c.json[c.pos]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			c.pos++
		} else {
			break
		}
	}
}

// 解析 JSON 字符串
func parseString(c *context, v *Value) int {
	// 跳过开始的引号
	p := c.pos + 1
	for p < len(c.json) {
		ch := c.json[p]
		if ch == '"' {
			// 找到结束引号
			v.S = c.json[c.pos+1 : p]
			v.Type = STRING
			c.pos = p + 1
			return PARSE_OK
		} else if ch == '\\' {
			// 处理转义字符
			p++
			if p >= len(c.json) {
				return PARSE_INVALID_STRING_ESCAPE
			}
			switch c.json[p] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				// 这些是有效的转义序列
			case 'u':
				// Unicode 转义序列
				if p+4 >= len(c.json) {
					return PARSE_INVALID_UNICODE_HEX
				}
				// 检查 4 位十六进制数字
				for i := 1; i <= 4; i++ {
					ch := c.json[p+i]
					if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
						return PARSE_INVALID_UNICODE_HEX
					}
				}
				p += 4
			default:
				return PARSE_INVALID_STRING_ESCAPE
			}
		} else if ch < 0x20 {
			// 控制字符不允许在字符串中
			return PARSE_INVALID_STRING_CHAR
		}
		p++
	}
	// 字符串没有结束引号
	return PARSE_MISS_QUOTATION_MARK
}

// 解析 null/true/false 字面值
func parseLiteral(c *context, v *Value, literal string, type_ int) int {
	// 检查字面值是否匹配
	if len(c.json) < c.pos+len(literal) {
		return PARSE_INVALID_VALUE
	}
	if c.json[c.pos:c.pos+len(literal)] != literal {
		return PARSE_INVALID_VALUE
	}
	c.pos += len(literal)
	v.Type = type_
	return PARSE_OK
}

// 解析数字
func parseNumber(c *context, v *Value) int {
	start := c.pos

	// 负号
	if c.currentChar() == '-' {
		c.nextChar()
	}

	// 整数部分
	if c.currentChar() == '0' {
		c.nextChar()
	} else if isDigit1to9(c.currentChar()) {
		c.nextChar()
		for isDigit(c.currentChar()) {
			c.nextChar()
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 小数部分
	if c.currentChar() == '.' {
		c.nextChar()
		if !isDigit(c.currentChar()) {
			return PARSE_INVALID_VALUE
		}
		c.nextChar()
		for isDigit(c.currentChar()) {
			c.nextChar()
		}
	}

	// 指数部分
	if c.currentChar() == 'e' || c.currentChar() == 'E' {
		c.nextChar()
		if c.currentChar() == '+' || c.currentChar() == '-' {
			c.nextChar()
		}
		if !isDigit(c.currentChar()) {
			return PARSE_INVALID_VALUE
		}
		c.nextChar()
		for isDigit(c.currentChar()) {
			c.nextChar()
		}
	}

	// 尝试转换字符串为浮点数
	numStr := c.json[start:c.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil || math.IsInf(num, 0) {
		return PARSE_NUMBER_TOO_BIG
	}
	v.N = num
	v.Type = NUMBER
	return PARSE_OK
}

// 解析 JSON 数组
func parseArray(c *context, v *Value) int {
	c.pos++ // 跳过 '['
	c.skipWhitespace()

	v.Type = ARRAY
	v.A = []*Value{}

	if c.currentChar() == ']' {
		c.pos++ // 跳过 ']'
		return PARSE_OK
	}

	for {
		// 解析数组元素
		elem := &Value{}
		ret := parseValue(c, elem)
		if ret != PARSE_OK {
			return ret
		}

		// 添加元素到数组
		v.A = append(v.A, elem)

		c.skipWhitespace()
		if c.currentChar() == ',' {
			c.pos++ // 跳过 ','
			c.skipWhitespace()
		} else if c.currentChar() == ']' {
			c.pos++ // 跳过 ']'
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}
	}
}

// 解析 JSON 对象
func parseObject(c *context, v *Value) int {
	c.pos++ // 跳过 '{'
	c.skipWhitespace()

	v.Type = OBJECT
	v.M = []*Member{}

	if c.currentChar() == '}' {
		c.pos++ // 跳过 '}'
		return PARSE_OK
	}

	for {
		// 解析键（必须是字符串）
		if c.currentChar() != '"' {
			return PARSE_MISS_KEY
		}

		key := &Value{}
		ret := parseString(c, key)
		if ret != PARSE_OK {
			return ret
		}

		// 检查冒号
		c.skipWhitespace()
		if c.currentChar() != ':' {
			return PARSE_MISS_COLON
		}
		c.pos++ // 跳过 ':'

		// 解析值
		val := &Value{}
		c.skipWhitespace()
		ret = parseValue(c, val)
		if ret != PARSE_OK {
			return ret
		}

		// 添加键值对到对象
		member := &Member{K: key.S, V: val}
		v.M = append(v.M, member)

		// 检查是否结束或有更多键值对
		c.skipWhitespace()
		if c.currentChar() == ',' {
			c.pos++ // 跳过 ','
			c.skipWhitespace()
		} else if c.currentChar() == '}' {
			c.pos++ // 跳过 '}'
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}
	}
}

// parseValue 解析任何 JSON 值
func parseValue(c *context, v *Value) int {
	switch c.currentChar() {
	case 'n':
		return parseLiteral(c, v, "null", NULL)
	case 't':
		return parseLiteral(c, v, "true", TRUE)
	case 'f':
		return parseLiteral(c, v, "false", FALSE)
	case '"':
		return parseString(c, v)
	case '[':
		return parseArray(c, v)
	case '{':
		return parseObject(c, v)
	case 0:
		return PARSE_EXPECT_VALUE
	default:
		return parseNumber(c, v)
	}
}

// 初始化错误消息映射
var errorMessages = map[int]string{}

func init() {
	errorMessages[PARSE_OK] = "解析成功"
	errorMessages[PARSE_EXPECT_VALUE] = "期望一个值"
	errorMessages[PARSE_INVALID_VALUE] = "无效的值"
	errorMessages[PARSE_ROOT_NOT_SINGULAR] = "根后有多余内容"
	errorMessages[PARSE_NUMBER_TOO_BIG] = "数字太大"
	errorMessages[PARSE_MISS_QUOTATION_MARK] = "缺少引号"
	errorMessages[PARSE_INVALID_STRING_ESCAPE] = "无效的转义字符"
	errorMessages[PARSE_INVALID_STRING_CHAR] = "无效的字符串字符"
	errorMessages[PARSE_INVALID_UNICODE_HEX] = "无效的 Unicode 十六进制数"
	errorMessages[PARSE_INVALID_UNICODE_SURROGATE] = "无效的 Unicode 代理对"
	errorMessages[PARSE_MISS_COMMA_OR_SQUARE_BRACKET] = "缺少逗号或方括号"
	errorMessages[PARSE_MISS_KEY] = "缺少键"
	errorMessages[PARSE_MISS_COLON] = "缺少冒号"
	errorMessages[PARSE_MISS_COMMA_OR_CURLY_BRACKET] = "缺少逗号或大括号"
}

// Parse 解析 JSON 字符串为 Value 结构
func Parse(v *Value, json string) int {
	c := context{json: json, stack: []byte{}, pos: 0}

	c.skipWhitespace()
	ret := parseValue(&c, v)
	if ret == PARSE_OK {
		c.skipWhitespace()
		if c.pos < len(json) {
			v.Type = NULL // 重置为默认值
			return PARSE_ROOT_NOT_SINGULAR
		}
	}
	return ret
}

// Marshal 将 Go 类型转换为 JSON 字符串
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将 JSON 字符串转换为 Go 类型
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Stringify 将 Value 结构转换为 JSON 字符串
func Stringify(v *Value) (string, error) {
	var result strings.Builder

	switch v.Type {
	case NULL:
		result.WriteString("null")
	case FALSE:
		result.WriteString("false")
	case TRUE:
		result.WriteString("true")
	case NUMBER:
		result.WriteString(strconv.FormatFloat(v.N, 'f', -1, 64))
	case STRING:
		// 处理字符串转义
		result.WriteString("\"")
		for _, ch := range v.S {
			switch ch {
			case '"':
				result.WriteString("\\\"")
			case '\\':
				result.WriteString("\\\\")
			case '\b':
				result.WriteString("\\b")
			case '\f':
				result.WriteString("\\f")
			case '\n':
				result.WriteString("\\n")
			case '\r':
				result.WriteString("\\r")
			case '\t':
				result.WriteString("\\t")
			default:
				if ch < 0x20 {
					// 控制字符需要使用 \u 转义
					result.WriteString(fmt.Sprintf("\\u%04X", ch))
				} else {
					result.WriteRune(ch)
				}
			}
		}
		result.WriteString("\"")
	case ARRAY:
		result.WriteString("[")
		for i, elem := range v.A {
			if i > 0 {
				result.WriteString(",")
			}
			elemStr, err := Stringify(elem)
			if err != nil {
				return "", err
			}
			result.WriteString(elemStr)
		}
		result.WriteString("]")
	case OBJECT:
		result.WriteString("{")
		for i, m := range v.M {
			if i > 0 {
				result.WriteString(",")
			}
			// 键需要用引号包裹并处理转义
			result.WriteString("\"")
			for _, ch := range m.K {
				switch ch {
				case '"':
					result.WriteString("\\\"")
				case '\\':
					result.WriteString("\\\\")
				case '\b':
					result.WriteString("\\b")
				case '\f':
					result.WriteString("\\f")
				case '\n':
					result.WriteString("\\n")
				case '\r':
					result.WriteString("\\r")
				case '\t':
					result.WriteString("\\t")
				default:
					if ch < 0x20 {
						result.WriteString(fmt.Sprintf("\\u%04X", ch))
					} else {
						result.WriteRune(ch)
					}
				}
			}
			result.WriteString("\":")

			// 值
			valStr, err := Stringify(m.V)
			if err != nil {
				return "", err
			}
			result.WriteString(valStr)
		}
		result.WriteString("}")
	default:
		return "", fmt.Errorf("未知的JSON类型")
	}

	return result.String(), nil
}

// GetTypeStr 获取类型的字符串表示
func GetTypeStr(type_ int) string {
	switch type_ {
	case NULL:
		return "NULL"
	case FALSE:
		return "FALSE"
	case TRUE:
		return "TRUE"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case ARRAY:
		return "ARRAY"
	case OBJECT:
		return "OBJECT"
	default:
		return "UNKNOWN"
	}
}

// GetErrorMessage 获取错误码对应的错误消息
func GetErrorMessage(err int) string {
	if msg, ok := errorMessages[err]; ok {
		return msg
	}
	return "未知错误"
}

// 辅助函数：获取和设置值

// GetType 获取值的类型
func GetType(v *Value) int {
	return v.Type
}

// GetNull 检查是否为 null
func GetNull(v *Value) bool {
	return v.Type == NULL
}

// GetBoolean 获取布尔值
func GetBoolean(v *Value) bool {
	if v.Type == TRUE {
		return true
	}
	return false
}

// GetNumber 获取数值
func GetNumber(v *Value) float64 {
	if v.Type == NUMBER {
		return v.N
	}
	return 0.0
}

// GetString 获取字符串
func GetString(v *Value) string {
	if v.Type == STRING {
		return v.S
	}
	return ""
}

// GetArray 获取数组
func GetArray(v *Value) []*Value {
	if v.Type == ARRAY {
		return v.A
	}
	return nil
}

// GetArraySize 获取数组大小
func GetArraySize(v *Value) int {
	if v.Type == ARRAY {
		return len(v.A)
	}
	return 0
}

// GetArrayElement 获取数组元素
func GetArrayElement(v *Value, index int) *Value {
	if v.Type == ARRAY && index >= 0 && index < len(v.A) {
		return v.A[index]
	}
	return nil
}

// GetObject 获取对象
func GetObject(v *Value) []*Member {
	if v.Type == OBJECT {
		return v.M
	}
	return nil
}

// GetObjectSize 获取对象大小
func GetObjectSize(v *Value) int {
	if v.Type == OBJECT {
		return len(v.M)
	}
	return 0
}

// FindObjectKey 在对象中查找键
func FindObjectKey(v *Value, key string) (*Value, bool) {
	if v.Type == OBJECT {
		for _, m := range v.M {
			if m.K == key {
				return m.V, true
			}
		}
	}
	return nil, false
}

// 辅助函数：设置值

// SetNull 设置为 null
func SetNull(v *Value) {
	v.Type = NULL
}

// SetBoolean 设置布尔值
func SetBoolean(v *Value, b bool) {
	if b {
		v.Type = TRUE
	} else {
		v.Type = FALSE
	}
}

// SetNumber 设置数值
func SetNumber(v *Value, n float64) {
	v.Type = NUMBER
	v.N = n
}

// SetString 设置字符串
func SetString(v *Value, s string) {
	v.Type = STRING
	v.S = s
}

// SetArray 设置为数组
func SetArray(v *Value, capacity int) {
	v.Type = ARRAY
	v.A = make([]*Value, 0, capacity)
}

// PushBackArrayElement 向数组末尾添加元素
func PushBackArrayElement(v *Value) *Value {
	if v.Type != ARRAY {
		return nil
	}
	elem := &Value{}
	v.A = append(v.A, elem)
	return elem
}

// PopBackArrayElement 移除数组末尾元素
func PopBackArrayElement(v *Value) bool {
	size := GetArraySize(v)
	if size == 0 {
		return false
	}
	v.A = v.A[:size-1]
	return true
}

// InsertArrayElement 在指定位置插入元素
func InsertArrayElement(v *Value, index int, value *Value) bool {
	if v.Type != ARRAY {
		return false
	}
	if index < 0 || index > len(v.A) {
		return false
	}
	if index == len(v.A) {
		v.A = append(v.A, value)
	} else {
		v.A = append(v.A[:index+1], v.A[index:]...)
		v.A[index] = value
	}
	return true
}

// EraseArrayElement 删除指定位置的元素
func EraseArrayElement(v *Value, index, count int) bool {
	size := GetArraySize(v)
	if index < 0 || count <= 0 || index+count > size {
		return false
	}
	v.A = append(v.A[:index], v.A[index+count:]...)
	return true
}

// ClearArray 清空数组
func ClearArray(v *Value) bool {
	if v.Type != ARRAY {
		return false
	}
	v.A = []*Value{}
	return true
}

// SetObject 设置为对象
func SetObject(v *Value) {
	v.Type = OBJECT
	v.M = []*Member{}
}

// SetObjectValue 设置对象的键值
func SetObjectValue(v *Value, key string) *Value {
	if v.Type != OBJECT {
		return nil
	}
	// 检查键是否已存在
	for _, m := range v.M {
		if m.K == key {
			return m.V // 键已存在，返回对应的值
		}
	}
	// 键不存在，添加新成员
	newValue := &Value{}
	newMember := &Member{K: key, V: newValue}
	v.M = append(v.M, newMember)
	return newValue
}

// RemoveObjectValue 删除对象中的键值对
func RemoveObjectValue(v *Value, key string) bool {
	if v.Type != OBJECT {
		return false
	}
	for i, m := range v.M {
		if m.K == key {
			v.M = append(v.M[:i], v.M[i+1:]...)
			return true
		}
	}
	return false // 键不存在
}

// ClearObject 清空对象
func ClearObject(v *Value) bool {
	if v.Type != OBJECT {
		return false
	}
	v.M = []*Member{}
	return true
}

// Copy 深拷贝 JSON 值
func Copy(dst, src *Value) {
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
		dst.Type = STRING
		dst.S = src.S
	case ARRAY:
		dst.Type = ARRAY
		dst.A = make([]*Value, len(src.A))
		for i, srcElem := range src.A {
			dst.A[i] = &Value{}
			Copy(dst.A[i], srcElem)
		}
	case OBJECT:
		dst.Type = OBJECT
		dst.M = make([]*Member, len(src.M))
		for i, srcMember := range src.M {
			dst.M[i] = &Member{K: srcMember.K, V: &Value{}}
			Copy(dst.M[i].V, srcMember.V)
		}
	}
}

// Equal 比较两个 JSON 值是否相等
func Equal(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case NULL, FALSE, TRUE:
		return true // 简单类型，类型相同即相等
	case NUMBER:
		return a.N == b.N
	case STRING:
		return a.S == b.S
	case ARRAY:
		if len(a.A) != len(b.A) {
			return false
		}
		for i := 0; i < len(a.A); i++ {
			if !Equal(a.A[i], b.A[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(a.M) != len(b.M) {
			return false
		}
		// 对象比较稍复杂，需要检查每个键是否都存在且值相等
		for _, aMember := range a.M {
			found := false
			for _, bMember := range b.M {
				if aMember.K == bMember.K {
					if !Equal(aMember.V, bMember.V) {
						return false
					}
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
		return true
	default:
		return false
	}
}
