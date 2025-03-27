// leptjson.go - JSON 库的基础实现
package leptjson

import (
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

	// 提取并解析数值
	numStr := c.json[start:c.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil || math.IsInf(num, 0) || math.IsNaN(num) {
		return PARSE_NUMBER_TOO_BIG
	}

	v.Type = NUMBER
	v.N = num
	return PARSE_OK
}

// 前向声明
var parseValue func(c *context, v *Value) int

// 解析数组
func parseArray(c *context, v *Value) int {
	c.pos++ // 跳过 '['
	v.Type = ARRAY
	v.A = make([]*Value, 0)

	c.skipWhitespace()
	if c.currentChar() == ']' {
		c.pos++ // 空数组
		return PARSE_OK
	}

	for {
		// 解析元素
		elem := &Value{}
		ret := parseValue(c, elem)
		if ret != PARSE_OK {
			return ret
		}
		v.A = append(v.A, elem)

		c.skipWhitespace()
		if c.currentChar() == ',' {
			c.pos++
			c.skipWhitespace()
		} else if c.currentChar() == ']' {
			c.pos++
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}
	}
}

// 解析对象
func parseObject(c *context, v *Value) int {
	c.pos++ // 跳过 '{'
	v.Type = OBJECT
	v.M = make([]*Member, 0)

	c.skipWhitespace()
	if c.currentChar() == '}' {
		c.pos++ // 空对象
		return PARSE_OK
	}

	for {
		// 解析键名
		key := &Value{}
		c.skipWhitespace()
		if c.currentChar() != '"' {
			return PARSE_MISS_KEY
		}
		if ret := parseString(c, key); ret != PARSE_OK {
			return ret
		}

		// 解析冒号
		c.skipWhitespace()
		if c.currentChar() != ':' {
			return PARSE_MISS_COLON
		}
		c.pos++

		// 解析值
		val := &Value{}
		if ret := parseValue(c, val); ret != PARSE_OK {
			return ret
		}

		// 添加键值对
		member := &Member{
			K: key.S,
			V: val,
		}
		v.M = append(v.M, member)

		// 查看下一个字符
		c.skipWhitespace()
		if c.currentChar() == ',' {
			c.pos++
			c.skipWhitespace()
		} else if c.currentChar() == '}' {
			c.pos++
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}
	}
}

// 实现前向声明的函数
func init() {
	parseValue = func(c *context, v *Value) int {
		c.skipWhitespace()
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
}

// Parse 解析 JSON 字符串
func Parse(v *Value, json string) int {
	c := &context{
		json:  json,
		stack: make([]byte, 0),
		pos:   0,
	}

	// 重置 Value
	*v = Value{}

	// 解析 JSON 值
	c.skipWhitespace()
	ret := parseValue(c, v)
	if ret != PARSE_OK {
		return ret
	}

	// 检查是否有多余内容
	c.skipWhitespace()
	if c.pos < len(c.json) {
		return PARSE_ROOT_NOT_SINGULAR
	}

	return PARSE_OK
}

// Stringify 将 JSON 值转换为字符串
func Stringify(v *Value) (string, error) {
	if v == nil {
		return "", fmt.Errorf("值不能为 nil")
	}

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
		result.WriteString(fmt.Sprintf("%q", v.S))
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
		for i, member := range v.M {
			if i > 0 {
				result.WriteString(",")
			}
			result.WriteString(fmt.Sprintf("%q:", member.K))
			valStr, err := Stringify(member.V)
			if err != nil {
				return "", err
			}
			result.WriteString(valStr)
		}
		result.WriteString("}")
	default:
		return "", fmt.Errorf("未知的值类型: %d", v.Type)
	}

	return result.String(), nil
}

// GetTypeStr 获取类型的字符串表示
func GetTypeStr(type_ int) string {
	switch type_ {
	case NULL:
		return "null"
	case TRUE:
		return "true"
	case FALSE:
		return "false"
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

// GetErrorMessage 获取解析错误的字符串表示
func GetErrorMessage(err int) string {
	switch err {
	case PARSE_OK:
		return "正常"
	case PARSE_EXPECT_VALUE:
		return "空 JSON 文档"
	case PARSE_INVALID_VALUE:
		return "无效的值"
	case PARSE_ROOT_NOT_SINGULAR:
		return "根之后还有其他内容"
	case PARSE_NUMBER_TOO_BIG:
		return "数字过大"
	case PARSE_MISS_QUOTATION_MARK:
		return "缺少引号"
	case PARSE_INVALID_STRING_ESCAPE:
		return "无效的转义序列"
	case PARSE_INVALID_STRING_CHAR:
		return "无效的字符串字符"
	case PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的 Unicode 代理项"
	case PARSE_INVALID_UNICODE_HEX:
		return "无效的 Unicode 十六进制数字"
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

// GetType 获取值类型
func GetType(v *Value) int {
	if v == nil {
		return NULL
	}
	return v.Type
}

// GetNull 获取 null 值
func GetNull(v *Value) bool {
	return v != nil && v.Type == NULL
}

// GetBoolean 获取布尔值
func GetBoolean(v *Value) bool {
	if v == nil || (v.Type != TRUE && v.Type != FALSE) {
		return false
	}
	return v.Type == TRUE
}

// GetNumber 获取数值
func GetNumber(v *Value) float64 {
	if v == nil || v.Type != NUMBER {
		return 0.0
	}
	return v.N
}

// GetString 获取字符串
func GetString(v *Value) string {
	if v == nil || v.Type != STRING {
		return ""
	}
	return v.S
}

// GetArray 获取数组
func GetArray(v *Value) []*Value {
	if v == nil || v.Type != ARRAY {
		return nil
	}
	return v.A
}

// GetArraySize 获取数组大小
func GetArraySize(v *Value) int {
	if v == nil || v.Type != ARRAY {
		return 0
	}
	return len(v.A)
}

// GetArrayElement 获取数组元素
func GetArrayElement(v *Value, index int) *Value {
	if v == nil || v.Type != ARRAY || index < 0 || index >= len(v.A) {
		return nil
	}
	return v.A[index]
}

// GetObject 获取对象
func GetObject(v *Value) []*Member {
	if v == nil || v.Type != OBJECT {
		return nil
	}
	return v.M
}

// GetObjectSize 获取对象大小
func GetObjectSize(v *Value) int {
	if v == nil || v.Type != OBJECT {
		return 0
	}
	return len(v.M)
}

// FindObjectKey 在对象中查找指定键
func FindObjectKey(v *Value, key string) (*Value, bool) {
	if v == nil || v.Type != OBJECT {
		return nil, false
	}
	for _, m := range v.M {
		if m.K == key {
			return m.V, true
		}
	}
	return nil, false
}

// SetNull 设置 null 值
func SetNull(v *Value) {
	*v = Value{Type: NULL}
}

// SetBoolean 设置布尔值
func SetBoolean(v *Value, b bool) {
	if b {
		*v = Value{Type: TRUE}
	} else {
		*v = Value{Type: FALSE}
	}
}

// SetNumber 设置数值
func SetNumber(v *Value, n float64) {
	*v = Value{Type: NUMBER, N: n}
}

// SetString 设置字符串
func SetString(v *Value, s string) {
	*v = Value{Type: STRING, S: s}
}

// SetArray 设置数组
func SetArray(v *Value, capacity int) {
	*v = Value{Type: ARRAY, A: make([]*Value, 0, capacity)}
}

// PushBackArrayElement 向数组末尾添加元素
func PushBackArrayElement(v *Value) *Value {
	if v.Type != ARRAY {
		return nil
	}
	element := &Value{}
	v.A = append(v.A, element)
	return element
}

// PopBackArrayElement 从数组末尾移除元素
func PopBackArrayElement(v *Value) bool {
	if v.Type != ARRAY || len(v.A) == 0 {
		return false
	}
	v.A = v.A[:len(v.A)-1]
	return true
}

// InsertArrayElement 在数组中插入元素
func InsertArrayElement(v *Value, index int, value *Value) bool {
	if v.Type != ARRAY || index < 0 || index > len(v.A) {
		return false
	}
	if index == len(v.A) {
		// 追加到末尾
		newValue := &Value{}
		Copy(newValue, value)
		v.A = append(v.A, newValue)
	} else {
		// 在中间插入
		newValue := &Value{}
		Copy(newValue, value)
		v.A = append(v.A[:index+1], v.A[index:]...)
		v.A[index] = newValue
	}
	return true
}

// EraseArrayElement 从数组中擦除元素
func EraseArrayElement(v *Value, index, count int) bool {
	if v.Type != ARRAY || index < 0 || count <= 0 || index+count > len(v.A) {
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
	v.A = v.A[:0]
	return true
}

// SetObject 设置对象
func SetObject(v *Value) {
	*v = Value{Type: OBJECT, M: make([]*Member, 0)}
}

// SetObjectValue 设置对象的键值对
func SetObjectValue(v *Value, key string) *Value {
	if v.Type != OBJECT {
		return nil
	}
	// 查找是否已存在该键
	for _, m := range v.M {
		if m.K == key {
			return m.V
		}
	}
	// 不存在则添加新成员
	member := &Member{
		K: key,
		V: &Value{},
	}
	v.M = append(v.M, member)
	return member.V
}

// RemoveObjectValue 移除对象的键值对
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
	return false
}

// ClearObject 清空对象
func ClearObject(v *Value) bool {
	if v.Type != OBJECT {
		return false
	}
	v.M = v.M[:0]
	return true
}

// Copy 复制 JSON 值
func Copy(dst, src *Value) {
	if dst == nil || src == nil {
		return
	}

	switch src.Type {
	case NULL:
		SetNull(dst)
	case FALSE:
		SetBoolean(dst, false)
	case TRUE:
		SetBoolean(dst, true)
	case NUMBER:
		SetNumber(dst, src.N)
	case STRING:
		SetString(dst, src.S)
	case ARRAY:
		SetArray(dst, len(src.A))
		for _, elem := range src.A {
			newElem := PushBackArrayElement(dst)
			Copy(newElem, elem)
		}
	case OBJECT:
		SetObject(dst)
		for _, m := range src.M {
			newVal := SetObjectValue(dst, m.K)
			Copy(newVal, m.V)
		}
	}
}

// Equal 比较两个 JSON 值是否相等
func Equal(a, b *Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case NULL, TRUE, FALSE:
		return true
	case NUMBER:
		return a.N == b.N
	case STRING:
		return a.S == b.S
	case ARRAY:
		if len(a.A) != len(b.A) {
			return false
		}
		for i := range a.A {
			if !Equal(a.A[i], b.A[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(a.M) != len(b.M) {
			return false
		}
		// 对象比较需要忽略成员顺序
		for _, mA := range a.M {
			found := false
			for _, mB := range b.M {
				if mA.K == mB.K {
					found = true
					if !Equal(mA.V, mB.V) {
						return false
					}
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
