// leptjson.go - JSON 库的基础数据结构和函数
package leptjson

// 导入所需的包
import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// 解析结果的常量
type ParseResult int

const (
	PARSE_OK ParseResult = iota
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
	PARSE_USER_STOPPED
)

// JSON 值的类型
type ValueType int

const (
	NULL ValueType = iota
	FALSE
	TRUE
	NUMBER
	STRING
	ARRAY
	OBJECT
)

// Member 表示对象中的一个键值对
type Member struct {
	K string // 键名
	V *Value // 值
}

// Value 表示一个 JSON 值
type Value struct {
	Type ValueType // 值的类型
	N    float64   // 数字值
	S    string    // 字符串值
	A    []*Value  // 数组值
	O    []Member  // 对象值
}

// 字符解析上下文
type parseContext struct {
	json  string // 要解析的 JSON 文本
	stack []byte // 解析临时缓冲区
}

// 字符串解析辅助函数
func parseHex4(s string) (rune, error) {
	if len(s) < 4 {
		return 0, fmt.Errorf("invalid unicode hex")
	}

	// 将16进制字符转换为数字
	r, err := strconv.ParseInt(s[:4], 16, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid unicode hex")
	}

	return rune(r), nil
}

// 编码 UTF-8
func encodeUTF8(u rune) []byte {
	if u <= 0x7F {
		return []byte{byte(u & 0xFF)}
	} else if u <= 0x7FF {
		return []byte{
			byte(0xC0 | ((u >> 6) & 0xFF)),
			byte(0x80 | (u & 0x3F)),
		}
	} else if u <= 0xFFFF {
		return []byte{
			byte(0xE0 | ((u >> 12) & 0xFF)),
			byte(0x80 | ((u >> 6) & 0x3F)),
			byte(0x80 | (u & 0x3F)),
		}
	} else {
		return []byte{
			byte(0xF0 | ((u >> 18) & 0xFF)),
			byte(0x80 | ((u >> 12) & 0x3F)),
			byte(0x80 | ((u >> 6) & 0x3F)),
			byte(0x80 | (u & 0x3F)),
		}
	}
}

// 解析4位十六进制数并编码为 UTF-8
func parseUTF8(pc *parseContext, s string, i *int) error {
	u, err := parseHex4(s[*i:])
	if err != nil {
		return err
	}
	*i += 4

	// 处理 UTF-16 代理对
	if u >= 0xD800 && u <= 0xDBFF {
		if *i+6 > len(s) {
			return fmt.Errorf("invalid unicode surrogate")
		}

		if s[*i] != '\\' || s[*i+1] != 'u' {
			return fmt.Errorf("invalid unicode surrogate")
		}

		*i += 2
		u2, err := parseHex4(s[*i:])
		if err != nil {
			return err
		}

		if u2 < 0xDC00 || u2 > 0xDFFF {
			return fmt.Errorf("invalid unicode surrogate")
		}

		*i += 4
		u = 0x10000 + (u-0xD800)*0x400 + (u2 - 0xDC00)
	}

	// 编码为 UTF-8 并添加到栈中
	pc.stack = append(pc.stack, encodeUTF8(u)...)
	return nil
}

// 解析字符串
func parseString(pc *parseContext, v *Value, start int) (ParseResult, int) {
	i := start + 1          // 跳过开始的引号
	pc.stack = pc.stack[:0] // 清空栈

	for i < len(pc.json) {
		ch := pc.json[i]
		if ch == '"' {
			v.Type = STRING
			v.S = string(pc.stack)
			return PARSE_OK, i + 1 // 跳过结束的引号
		}

		if ch == '\\' {
			i++
			if i >= len(pc.json) {
				return PARSE_INVALID_STRING_ESCAPE, 0
			}

			switch pc.json[i] {
			case '"':
				pc.stack = append(pc.stack, '"')
			case '\\':
				pc.stack = append(pc.stack, '\\')
			case '/':
				pc.stack = append(pc.stack, '/')
			case 'b':
				pc.stack = append(pc.stack, '\b')
			case 'f':
				pc.stack = append(pc.stack, '\f')
			case 'n':
				pc.stack = append(pc.stack, '\n')
			case 'r':
				pc.stack = append(pc.stack, '\r')
			case 't':
				pc.stack = append(pc.stack, '\t')
			case 'u':
				i++
				if err := parseUTF8(pc, pc.json, &i); err != nil {
					return PARSE_INVALID_UNICODE_SURROGATE, 0
				}
				continue // parseUTF8 已经更新了 i，所以我们跳过自增
			default:
				return PARSE_INVALID_STRING_ESCAPE, 0
			}
		} else if ch < 0x20 {
			// 不允许未转义的控制字符
			return PARSE_INVALID_STRING_CHAR, 0
		} else {
			// 普通字符
			pc.stack = append(pc.stack, ch)
		}
		i++
	}

	// 如果到达这里，说明字符串没有结束
	return PARSE_MISS_QUOTATION_MARK, 0
}

// 解析数字
func parseNumber(pc *parseContext, v *Value, start int) (ParseResult, int) {
	s := pc.json[start:]

	// 查找数字结束的位置
	end := 0
	for end < len(s) {
		if strings.IndexByte("0123456789.eE+-", s[end]) < 0 {
			break
		}
		end++
	}

	// 解析成浮点数
	num, err := strconv.ParseFloat(s[:end], 64)
	if err != nil {
		return PARSE_INVALID_VALUE, 0
	}

	// 检查是否太大
	if math.IsInf(num, 0) || math.IsNaN(num) {
		return PARSE_NUMBER_TOO_BIG, 0
	}

	v.Type = NUMBER
	v.N = num
	return PARSE_OK, start + end
}

// 解析数组
func parseArray(pc *parseContext, v *Value, start int) (ParseResult, int) {
	i := start + 1 // 跳过 [
	v.Type = ARRAY
	v.A = []*Value{}

	// 处理空数组
	for i < len(pc.json) {
		if pc.json[i] == ' ' || pc.json[i] == '\t' || pc.json[i] == '\n' || pc.json[i] == '\r' {
			i++
			continue
		}

		if pc.json[i] == ']' {
			return PARSE_OK, i + 1
		}

		// 解析数组元素
		element := &Value{}
		ret, pos := parseValue(pc, element, i)
		if ret != PARSE_OK {
			return ret, 0
		}

		v.A = append(v.A, element)
		i = pos

		// 处理逗号或结束括号
		for i < len(pc.json) {
			if pc.json[i] == ' ' || pc.json[i] == '\t' || pc.json[i] == '\n' || pc.json[i] == '\r' {
				i++
				continue
			}

			if pc.json[i] == ',' {
				i++
				break
			} else if pc.json[i] == ']' {
				return PARSE_OK, i + 1
			} else {
				return PARSE_MISS_COMMA_OR_SQUARE_BRACKET, 0
			}
		}
	}

	// 如果到达这里，说明数组没有正确结束
	return PARSE_MISS_COMMA_OR_SQUARE_BRACKET, 0
}

// 解析对象
func parseObject(pc *parseContext, v *Value, start int) (ParseResult, int) {
	i := start + 1 // 跳过 {
	v.Type = OBJECT
	v.O = []Member{}

	// 处理空对象
	for i < len(pc.json) {
		if pc.json[i] == ' ' || pc.json[i] == '\t' || pc.json[i] == '\n' || pc.json[i] == '\r' {
			i++
			continue
		}

		if pc.json[i] == '}' {
			return PARSE_OK, i + 1
		}

		// 解析键
		if pc.json[i] != '"' {
			return PARSE_MISS_KEY, 0
		}

		key := &Value{}
		ret, pos := parseString(pc, key, i)
		if ret != PARSE_OK {
			return ret, 0
		}
		i = pos

		// 处理冒号
		for i < len(pc.json) {
			if pc.json[i] == ' ' || pc.json[i] == '\t' || pc.json[i] == '\n' || pc.json[i] == '\r' {
				i++
				continue
			}

			if pc.json[i] == ':' {
				i++
				break
			} else {
				return PARSE_MISS_COLON, 0
			}
		}

		// 解析值
		value := &Value{}
		ret, pos = parseValue(pc, value, i)
		if ret != PARSE_OK {
			return ret, 0
		}

		// 添加键值对
		member := Member{K: key.S, V: value}
		v.O = append(v.O, member)

		i = pos

		// 处理逗号或结束大括号
		for i < len(pc.json) {
			if pc.json[i] == ' ' || pc.json[i] == '\t' || pc.json[i] == '\n' || pc.json[i] == '\r' {
				i++
				continue
			}

			if pc.json[i] == ',' {
				i++
				break
			} else if pc.json[i] == '}' {
				return PARSE_OK, i + 1
			} else {
				return PARSE_MISS_COMMA_OR_CURLY_BRACKET, 0
			}
		}
	}

	// 如果到达这里，说明对象没有正确结束
	return PARSE_MISS_COMMA_OR_CURLY_BRACKET, 0
}

// 跳过空白字符
func skipWhitespace(s string, start int) int {
	i := start
	for i < len(s) {
		if s[i] == ' ' || s[i] == '\t' || s[i] == '\n' || s[i] == '\r' {
			i++
		} else {
			break
		}
	}
	return i
}

// 解析 null, true, false 字面量
func parseLiteral(pc *parseContext, v *Value, literal string, vtype ValueType, start int) (ParseResult, int) {
	if len(pc.json) < start+len(literal) {
		return PARSE_INVALID_VALUE, 0
	}

	if pc.json[start:start+len(literal)] != literal {
		return PARSE_INVALID_VALUE, 0
	}

	// 确保字面量后没有其他非空白字符
	if start+len(literal) < len(pc.json) {
		ch := pc.json[start+len(literal)]
		if strings.IndexByte(" \t\n\r,]}\"", ch) < 0 {
			return PARSE_INVALID_VALUE, 0
		}
	}

	v.Type = vtype
	return PARSE_OK, start + len(literal)
}

// 解析 JSON 值
func parseValue(pc *parseContext, v *Value, start int) (ParseResult, int) {
	start = skipWhitespace(pc.json, start)
	if start >= len(pc.json) {
		return PARSE_EXPECT_VALUE, 0
	}

	switch pc.json[start] {
	case 'n':
		return parseLiteral(pc, v, "null", NULL, start)
	case 't':
		return parseLiteral(pc, v, "true", TRUE, start)
	case 'f':
		return parseLiteral(pc, v, "false", FALSE, start)
	case '"':
		return parseString(pc, v, start)
	case '[':
		return parseArray(pc, v, start)
	case '{':
		return parseObject(pc, v, start)
	case 0:
		return PARSE_EXPECT_VALUE, 0
	default:
		return parseNumber(pc, v, start)
	}
}

// Parse 解析 JSON 字符串，并将结果存储在 v 中
func Parse(v *Value, json string) ParseResult {
	pc := &parseContext{
		json:  json,
		stack: []byte{},
	}

	ret, pos := parseValue(pc, v, 0)
	if ret != PARSE_OK {
		return ret
	}

	// 检查根值后是否有非空白字符
	pos = skipWhitespace(pc.json, pos)
	if pos < len(pc.json) {
		return PARSE_ROOT_NOT_SINGULAR
	}

	return PARSE_OK
}

// Stringify 将 JSON 值转换为 JSON 字符串
func Stringify(v *Value) (string, error) {
	if v == nil {
		return "", fmt.Errorf("不能序列化 nil 值")
	}

	var sb strings.Builder

	// 递归生成 JSON 字符串
	if err := generateString(v, &sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}

// 辅助函数：转义 JSON 字符串
func escapeString(s string, sb *strings.Builder) {
	sb.WriteByte('"')
	for _, ch := range s {
		switch ch {
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
			if ch < 0x20 {
				sb.WriteString(fmt.Sprintf("\\u%04X", ch))
			} else {
				sb.WriteRune(ch)
			}
		}
	}
	sb.WriteByte('"')
}

// 生成JSON字符串
func generateString(v *Value, sb *strings.Builder) error {
	switch v.Type {
	case NULL:
		sb.WriteString("null")
	case FALSE:
		sb.WriteString("false")
	case TRUE:
		sb.WriteString("true")
	case NUMBER:
		// 使用 strconv.FormatFloat 格式化浮点数
		sb.WriteString(strconv.FormatFloat(v.N, 'f', -1, 64))
	case STRING:
		escapeString(v.S, sb)
	case ARRAY:
		sb.WriteByte('[')
		for i, elem := range v.A {
			if i > 0 {
				sb.WriteByte(',')
			}
			if err := generateString(elem, sb); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case OBJECT:
		sb.WriteByte('{')
		for i, member := range v.O {
			if i > 0 {
				sb.WriteByte(',')
			}
			escapeString(member.K, sb)
			sb.WriteByte(':')
			if err := generateString(member.V, sb); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		return fmt.Errorf("未知的值类型")
	}

	return nil
}

// 数据访问功能

// GetType 返回 JSON 值的类型
func GetType(v *Value) ValueType {
	if v == nil {
		return NULL
	}
	return v.Type
}

// SetNull 设置值为 null
func SetNull(v *Value) {
	Free(v)
	v.Type = NULL
}

// GetBoolean 获取布尔值
func GetBoolean(v *Value) bool {
	if v == nil || (v.Type != TRUE && v.Type != FALSE) {
		return false
	}
	return v.Type == TRUE
}

// SetBoolean 设置布尔值
func SetBoolean(v *Value, b bool) {
	Free(v)
	if b {
		v.Type = TRUE
	} else {
		v.Type = FALSE
	}
}

// GetNumber 获取数字值
func GetNumber(v *Value) float64 {
	if v == nil || v.Type != NUMBER {
		return 0.0
	}
	return v.N
}

// SetNumber 设置数字值
func SetNumber(v *Value, n float64) {
	Free(v)
	v.Type = NUMBER
	v.N = n
}

// GetString 获取字符串值
func GetString(v *Value) string {
	if v == nil || v.Type != STRING {
		return ""
	}
	return v.S
}

// SetString 设置字符串值
func SetString(v *Value, s string) {
	Free(v)
	v.Type = STRING
	v.S = s
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

// SetArray 设置值为空数组
func SetArray(v *Value, capacity int) {
	Free(v)
	v.Type = ARRAY
	v.A = make([]*Value, 0, capacity)
}

// PushBackArrayElement 添加数组元素
func PushBackArrayElement(v *Value) *Value {
	if v == nil || v.Type != ARRAY {
		return nil
	}
	element := &Value{}
	v.A = append(v.A, element)
	return element
}

// PopBackArrayElement 删除最后一个数组元素
func PopBackArrayElement(v *Value) bool {
	if v == nil || v.Type != ARRAY || len(v.A) == 0 {
		return false
	}
	v.A = v.A[:len(v.A)-1]
	return true
}

// InsertArrayElement 插入数组元素
func InsertArrayElement(v *Value, index int) *Value {
	if v == nil || v.Type != ARRAY || index < 0 || index > len(v.A) {
		return nil
	}

	element := &Value{}
	if index == len(v.A) {
		v.A = append(v.A, element)
	} else {
		v.A = append(v.A, nil) // 先扩容
		copy(v.A[index+1:], v.A[index:])
		v.A[index] = element
	}
	return element
}

// EraseArrayElement 删除指定数组元素
func EraseArrayElement(v *Value, index, count int) bool {
	if v == nil || v.Type != ARRAY || index < 0 || count <= 0 || index+count > len(v.A) {
		return false
	}

	v.A = append(v.A[:index], v.A[index+count:]...)
	return true
}

// ClearArray 清空数组
func ClearArray(v *Value) {
	if v == nil || v.Type != ARRAY {
		return
	}
	v.A = v.A[:0]
}

// GetObjectSize 获取对象成员数
func GetObjectSize(v *Value) int {
	if v == nil || v.Type != OBJECT {
		return 0
	}
	return len(v.O)
}

// GetObjectKey 获取对象的键
func GetObjectKey(v *Value, index int) string {
	if v == nil || v.Type != OBJECT || index < 0 || index >= len(v.O) {
		return ""
	}
	return v.O[index].K
}

// GetObjectValue 获取对象的值
func GetObjectValue(v *Value, index int) *Value {
	if v == nil || v.Type != OBJECT || index < 0 || index >= len(v.O) {
		return nil
	}
	return v.O[index].V
}

// GetObjectValueByKey 获取指定键的值
func GetObjectValueByKey(v *Value, key string) *Value {
	if v == nil || v.Type != OBJECT {
		return nil
	}
	for _, member := range v.O {
		if member.K == key {
			return member.V
		}
	}
	return nil
}

// FindObjectKey 查找对象中的键，返回找到的值和是否存在
func FindObjectKey(v *Value, key string) (*Value, bool) {
	if v == nil || v.Type != OBJECT {
		return nil, false
	}
	for _, member := range v.O {
		if member.K == key {
			return member.V, true
		}
	}
	return nil, false
}

// SetObject 设置值为空对象
func SetObject(v *Value) {
	Free(v)
	v.Type = OBJECT
	v.O = []Member{}
}

// SetObjectValue 设置对象键值
func SetObjectValue(v *Value, key string) *Value {
	// 先查找是否已存在该键
	if value := GetObjectValueByKey(v, key); value != nil {
		return value
	}

	// 添加新键值对
	value := &Value{}
	v.O = append(v.O, Member{K: key, V: value})
	return value
}

// RemoveObjectValue 移除对象成员
func RemoveObjectValue(v *Value, index int) bool {
	if v == nil || v.Type != OBJECT || index < 0 || index >= len(v.O) {
		return false
	}

	v.O = append(v.O[:index], v.O[index+1:]...)
	return true
}

// RemoveObjectValueByKey 根据键移除对象成员
func RemoveObjectValueByKey(v *Value, key string) bool {
	if v == nil || v.Type != OBJECT {
		return false
	}

	for i, member := range v.O {
		if member.K == key {
			return RemoveObjectValue(v, i)
		}
	}
	return false
}

// ClearObject 清空对象
func ClearObject(v *Value) {
	if v == nil || v.Type != OBJECT {
		return
	}
	v.O = v.O[:0]
}

// Free 释放值所占用的内存
func Free(v *Value) {
	if v == nil {
		return
	}

	switch v.Type {
	case ARRAY:
		for _, element := range v.A {
			Free(element)
		}
		v.A = nil
	case OBJECT:
		for _, member := range v.O {
			Free(member.V)
		}
		v.O = nil
	}

	v.Type = NULL
}

// Copy 复制值
func Copy(dst, src *Value) {
	if dst == nil || src == nil {
		return
	}

	Free(dst)

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
		for i, element := range src.A {
			dst.A[i] = &Value{}
			Copy(dst.A[i], element)
		}
	case OBJECT:
		dst.Type = OBJECT
		dst.O = make([]Member, len(src.O))
		for i, member := range src.O {
			dst.O[i].K = member.K
			dst.O[i].V = &Value{}
			Copy(dst.O[i].V, member.V)
		}
	}
}

// Move 移动值
func Move(dst, src *Value) {
	if dst == nil || src == nil || dst == src {
		return
	}

	Free(dst)

	// 简单地交换内容，避免深拷贝
	*dst = *src

	// 重置 src
	src.Type = NULL
	src.N = 0
	src.S = ""
	src.A = nil
	src.O = nil
}

// Swap 交换两个值
func Swap(lhs, rhs *Value) {
	if lhs == nil || rhs == nil || lhs == rhs {
		return
	}

	temp := *lhs
	*lhs = *rhs
	*rhs = temp
}

// Equal 比较两个 JSON 值是否相等
func Equal(lhs, rhs *Value) bool {
	if lhs == nil || rhs == nil {
		return lhs == rhs
	}

	if lhs.Type != rhs.Type {
		return false
	}

	switch lhs.Type {
	case NULL, TRUE, FALSE:
		return true
	case NUMBER:
		return lhs.N == rhs.N
	case STRING:
		return lhs.S == rhs.S
	case ARRAY:
		if len(lhs.A) != len(rhs.A) {
			return false
		}
		for i := range lhs.A {
			if !Equal(lhs.A[i], rhs.A[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(lhs.O) != len(rhs.O) {
			return false
		}
		// 检查 rhs 中是否有 lhs 的每个成员，并且值相等
		for _, lmember := range lhs.O {
			found := false
			for _, rmember := range rhs.O {
				if lmember.K == rmember.K {
					if !Equal(lmember.V, rmember.V) {
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
	}
	return false
}
