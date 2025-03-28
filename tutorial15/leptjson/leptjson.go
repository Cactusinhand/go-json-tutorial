// leptjson.go - 实现基于tutorial15的最新版leptjson库
package leptjson

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// 解析状态枚举
const (
	PARSE_OK                           = 0
	PARSE_EXPECT_VALUE                 = 1
	PARSE_INVALID_VALUE                = 2
	PARSE_ROOT_NOT_SINGULAR            = 3
	PARSE_NUMBER_TOO_BIG               = 4
	PARSE_MISS_QUOTATION_MARK          = 5
	PARSE_INVALID_STRING_ESCAPE        = 6
	PARSE_INVALID_STRING_CHAR          = 7
	PARSE_INVALID_UNICODE_HEX          = 8
	PARSE_INVALID_UNICODE_SURROGATE    = 9
	PARSE_MISS_COMMA_OR_SQUARE_BRACKET = 10
	PARSE_MISS_KEY                     = 11
	PARSE_MISS_COLON                   = 12
	PARSE_MISS_COMMA_OR_CURLY_BRACKET  = 13
)

// JSON 值类型枚举
const (
	JSON_NULL   = 0
	JSON_FALSE  = 1
	JSON_TRUE   = 2
	JSON_NUMBER = 3
	JSON_STRING = 4
	JSON_ARRAY  = 5
	JSON_OBJECT = 6
)

// ConcurrentParseOptions 定义并发解析的选项
type ConcurrentParseOptions struct {
	// 块大小（字节）
	ChunkSize int
	// 并发工作线程数量
	WorkerCount int
	// 是否自动检测最佳工作线程数量
	AutoWorkers bool
}

// Value 表示一个JSON值
type Value struct {
	Type int
	N    float64     // number
	S    string      // string
	A    []*Value    // array
	M    [][2]string // object, 使用二维数组存储键值对, 键是第一个元素, 值是第二个元素的JSON表示
}

// 解析器结构体
type parser struct {
	json string
	pos  int
}

// 设置JSON值为null
func SetNull(v *Value) {
	v.Type = JSON_NULL
}

// 获取JSON值的类型
func GetType(v *Value) int {
	return v.Type
}

// SetBoolean 设置JSON值为布尔值
func SetBoolean(v *Value, b bool) {
	if b {
		v.Type = JSON_TRUE
	} else {
		v.Type = JSON_FALSE
	}
}

// SetNumber 设置JSON值为数值
func SetNumber(v *Value, n float64) {
	v.Type = JSON_NUMBER
	v.N = n
}

// SetString 设置JSON值为字符串
func SetString(v *Value, s string) {
	v.Type = JSON_STRING
	v.S = s
}

// SetArray 设置JSON值为数组
func SetArray(v *Value, capacity int) {
	v.Type = JSON_ARRAY
	v.A = make([]*Value, 0, capacity)
}

// GetArraySize 获取数组大小
func GetArraySize(v *Value) int {
	if v.Type != JSON_ARRAY {
		return 0
	}
	return len(v.A)
}

// GetArrayElement 获取数组元素
func GetArrayElement(v *Value, index int) *Value {
	if v.Type != JSON_ARRAY || index < 0 || index >= len(v.A) {
		return nil
	}
	return v.A[index]
}

// InsertArrayElement 插入数组元素
func InsertArrayElement(v *Value, element *Value) *Value {
	if v.Type != JSON_ARRAY {
		return nil
	}
	newElement := &Value{}
	v.A = append(v.A, newElement)
	// 可以在这里拷贝element的值到newElement, 为简单起见这里省略了
	return newElement
}

// SetObject 设置JSON值为对象
func SetObject(v *Value) {
	v.Type = JSON_OBJECT
	v.M = make([][2]string, 0)
}

// SetObjectValue 设置对象属性
func SetObjectValue(v *Value, key string) *Value {
	if v.Type != JSON_OBJECT {
		return nil
	}

	// 尝试查找现有的键
	for i := 0; i < len(v.M); i++ {
		if v.M[i][0] == key {
			// 如果键已存在, 返回与之对应的值
			value := &Value{}
			if Parse(value, v.M[i][1]) != PARSE_OK {
				return nil
			}
			return value
		}
	}

	// 添加新键值对, 先添加键
	pair := [2]string{key, ""}
	v.M = append(v.M, pair)

	// 返回一个新值, 外部需要调用Stringify后设置回v.M[len(v.M)-1][1]
	return &Value{}
}

// Parse 解析JSON字符串
func Parse(v *Value, json string) int {
	p := &parser{json: json, pos: 0}

	// 初始化Value
	v.Type = JSON_NULL

	p.parseWhitespace()
	code := p.parseValue(v)

	if code == PARSE_OK {
		p.parseWhitespace()
		if p.pos < len(p.json) {
			return PARSE_ROOT_NOT_SINGULAR
		}
	}

	return code
}

// 解析空白字符
func (p *parser) parseWhitespace() {
	for p.pos < len(p.json) {
		c := p.json[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
		} else {
			break
		}
	}
}

// 解析值
func (p *parser) parseValue(v *Value) int {
	if p.pos >= len(p.json) {
		return PARSE_EXPECT_VALUE
	}

	switch p.json[p.pos] {
	case 'n':
		return p.parseLiteral(v, "null", JSON_NULL)
	case 't':
		return p.parseLiteral(v, "true", JSON_TRUE)
	case 'f':
		return p.parseLiteral(v, "false", JSON_FALSE)
	case '"':
		return p.parseString(v)
	case '[':
		return p.parseArray(v)
	case '{':
		return p.parseObject(v)
	case 0:
		return PARSE_EXPECT_VALUE
	default:
		return p.parseNumber(v)
	}
}

// 解析字面值 (null, true, false)
func (p *parser) parseLiteral(v *Value, literal string, type_ int) int {
	if len(p.json) < p.pos+len(literal) {
		return PARSE_INVALID_VALUE
	}

	for i := 0; i < len(literal); i++ {
		if p.json[p.pos+i] != literal[i] {
			return PARSE_INVALID_VALUE
		}
	}

	p.pos += len(literal)
	v.Type = type_

	return PARSE_OK
}

// 解析数字
func (p *parser) parseNumber(v *Value) int {
	startPos := p.pos

	// 检查负号
	if p.pos < len(p.json) && p.json[p.pos] == '-' {
		p.pos++
	}

	// 检查整数部分
	if p.pos < len(p.json) && p.json[p.pos] == '0' {
		p.pos++
	} else if p.pos < len(p.json) && p.json[p.pos] >= '1' && p.json[p.pos] <= '9' {
		p.pos++
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	} else {
		return PARSE_INVALID_VALUE
	}

	// 检查小数部分
	if p.pos < len(p.json) && p.json[p.pos] == '.' {
		p.pos++
		if p.pos >= len(p.json) || p.json[p.pos] < '0' || p.json[p.pos] > '9' {
			return PARSE_INVALID_VALUE
		}
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	}

	// 检查指数部分
	if p.pos < len(p.json) && (p.json[p.pos] == 'e' || p.json[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.json) && (p.json[p.pos] == '+' || p.json[p.pos] == '-') {
			p.pos++
		}
		if p.pos >= len(p.json) || p.json[p.pos] < '0' || p.json[p.pos] > '9' {
			return PARSE_INVALID_VALUE
		}
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	}

	// 将字符串转换为浮点数
	numStr := p.json[startPos:p.pos]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return PARSE_NUMBER_TOO_BIG
	}

	v.Type = JSON_NUMBER
	v.N = num

	return PARSE_OK
}

// 解析字符串中的转义字符
func (p *parser) parseStringRaw() (string, int) {
	p.pos++ // 跳过开头的双引号

	var result strings.Builder

	for p.pos < len(p.json) {
		ch := p.json[p.pos]
		if ch == '"' {
			p.pos++
			return result.String(), PARSE_OK
		} else if ch == '\\' {
			p.pos++
			if p.pos >= len(p.json) {
				return "", PARSE_INVALID_STRING_ESCAPE
			}

			switch p.json[p.pos] {
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
				// Unicode 转义处理
				if p.pos+4 >= len(p.json) {
					return "", PARSE_INVALID_UNICODE_HEX
				}

				var codepoint int
				for i := 1; i <= 4; i++ {
					p.pos++
					ch := p.json[p.pos]
					codepoint = codepoint << 4

					if ch >= '0' && ch <= '9' {
						codepoint |= int(ch - '0')
					} else if ch >= 'A' && ch <= 'F' {
						codepoint |= int(ch - 'A' + 10)
					} else if ch >= 'a' && ch <= 'f' {
						codepoint |= int(ch - 'a' + 10)
					} else {
						return "", PARSE_INVALID_UNICODE_HEX
					}
				}

				// 处理代理对
				if codepoint >= 0xD800 && codepoint <= 0xDBFF {
					// 高代理项, 需要后面跟一个低代理项
					if p.pos+6 >= len(p.json) || p.json[p.pos+1] != '\\' || p.json[p.pos+2] != 'u' {
						return "", PARSE_INVALID_UNICODE_SURROGATE
					}

					p.pos += 2 // 跳过 \u
					var codepoint2 int

					for i := 1; i <= 4; i++ {
						p.pos++
						ch := p.json[p.pos]
						codepoint2 = codepoint2 << 4

						if ch >= '0' && ch <= '9' {
							codepoint2 |= int(ch - '0')
						} else if ch >= 'A' && ch <= 'F' {
							codepoint2 |= int(ch - 'A' + 10)
						} else if ch >= 'a' && ch <= 'f' {
							codepoint2 |= int(ch - 'a' + 10)
						} else {
							return "", PARSE_INVALID_UNICODE_HEX
						}
					}

					if codepoint2 < 0xDC00 || codepoint2 > 0xDFFF {
						return "", PARSE_INVALID_UNICODE_SURROGATE
					}

					// 计算Unicode码点
					codepoint = 0x10000 + ((codepoint - 0xD800) << 10) + (codepoint2 - 0xDC00)
				}

				// 转换Unicode码点为UTF-8
				var buf [4]byte
				n := utf8.EncodeRune(buf[:], rune(codepoint))
				result.Write(buf[:n])
				p.pos++ // 跳过最后一个十六进制数字
				continue
			default:
				return "", PARSE_INVALID_STRING_ESCAPE
			}
		} else if ch < 0x20 {
			// 控制字符不允许出现在字符串中
			return "", PARSE_INVALID_STRING_CHAR
		} else {
			// 普通字符直接添加
			result.WriteByte(ch)
		}

		p.pos++
	}

	return "", PARSE_MISS_QUOTATION_MARK
}

// 解析字符串
func (p *parser) parseString(v *Value) int {
	s, code := p.parseStringRaw()
	if code != PARSE_OK {
		return code
	}

	v.Type = JSON_STRING
	v.S = s

	return PARSE_OK
}

// 解析数组
func (p *parser) parseArray(v *Value) int {
	p.pos++ // 跳过开头的'['
	p.parseWhitespace()

	if p.pos < len(p.json) && p.json[p.pos] == ']' {
		p.pos++
		SetArray(v, 0)
		return PARSE_OK
	}

	var elements []*Value

	for {
		elem := &Value{}
		code := p.parseValue(elem)

		if code != PARSE_OK {
			return code
		}

		elements = append(elements, elem)

		p.parseWhitespace()
		if p.pos >= len(p.json) {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}

		if p.json[p.pos] == ',' {
			p.pos++
			p.parseWhitespace()
		} else if p.json[p.pos] == ']' {
			p.pos++
			v.Type = JSON_ARRAY
			v.A = elements
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_SQUARE_BRACKET
		}
	}
}

// 解析对象
func (p *parser) parseObject(v *Value) int {
	p.pos++ // 跳过开头的'{'
	p.parseWhitespace()

	if p.pos < len(p.json) && p.json[p.pos] == '}' {
		p.pos++
		SetObject(v)
		return PARSE_OK
	}

	var members [][2]string

	for {
		// 解析键
		if p.pos >= len(p.json) || p.json[p.pos] != '"' {
			return PARSE_MISS_KEY
		}

		key, code := p.parseStringRaw()
		if code != PARSE_OK {
			return code
		}

		// 解析冒号
		p.parseWhitespace()
		if p.pos >= len(p.json) || p.json[p.pos] != ':' {
			return PARSE_MISS_COLON
		}
		p.pos++
		p.parseWhitespace()

		// 解析值
		valueObj := &Value{}
		code = p.parseValue(valueObj)
		if code != PARSE_OK {
			return code
		}

		// 将值转换为字符串形式并存储
		valueStr, err := stringify(valueObj)
		if err != nil {
			return PARSE_INVALID_VALUE
		}

		// 添加键值对
		pair := [2]string{key, valueStr}
		members = append(members, pair)

		// 解析逗号或右花括号
		p.parseWhitespace()
		if p.pos >= len(p.json) {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}

		if p.json[p.pos] == ',' {
			p.pos++
			p.parseWhitespace()
		} else if p.json[p.pos] == '}' {
			p.pos++
			v.Type = JSON_OBJECT
			v.M = members
			return PARSE_OK
		} else {
			return PARSE_MISS_COMMA_OR_CURLY_BRACKET
		}
	}
}

// 内部序列化函数
func stringify(v *Value) (string, error) {
	var sb strings.Builder
	if err := stringifyValue(v, &sb); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// stringifyValue 将JSON值序列化为字符串（内部实现）
func stringifyValue(v *Value, sb *strings.Builder) error {
	switch v.Type {
	case JSON_NULL:
		sb.WriteString("null")
	case JSON_FALSE:
		sb.WriteString("false")
	case JSON_TRUE:
		sb.WriteString("true")
	case JSON_NUMBER:
		// 处理特殊数字
		if math.IsNaN(v.N) || math.IsInf(v.N, 0) {
			return fmt.Errorf("无效的数字 (NaN 或 Infinity)")
		}
		sb.WriteString(strconv.FormatFloat(v.N, 'f', -1, 64))
	case JSON_STRING:
		writeJSONString(v.S, sb)
	case JSON_ARRAY:
		sb.WriteByte('[')
		for i, elem := range v.A {
			if i > 0 {
				sb.WriteByte(',')
			}
			if err := stringifyValue(elem, sb); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case JSON_OBJECT:
		sb.WriteByte('{')
		for i, pair := range v.M {
			if i > 0 {
				sb.WriteByte(',')
			}
			writeJSONString(pair[0], sb)
			sb.WriteByte(':')
			sb.WriteString(pair[1])
		}
		sb.WriteByte('}')
	default:
		return fmt.Errorf("未知的JSON值类型")
	}
	return nil
}

// 将字符串编码为JSON字符串，处理转义字符
func writeJSONString(s string, sb *strings.Builder) {
	sb.WriteByte('"')
	for i := 0; i < len(s); i++ {
		ch := s[i]
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
				sb.WriteByte(ch)
			}
		}
	}
	sb.WriteByte('"')
}

// Equal 比较两个JSON值是否相等
func Equal(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case JSON_NULL:
		return true
	case JSON_FALSE, JSON_TRUE:
		return a.Type == b.Type
	case JSON_NUMBER:
		return a.N == b.N
	case JSON_STRING:
		return a.S == b.S
	case JSON_ARRAY:
		if len(a.A) != len(b.A) {
			return false
		}
		for i := 0; i < len(a.A); i++ {
			if !Equal(a.A[i], b.A[i]) {
				return false
			}
		}
		return true
	case JSON_OBJECT:
		if len(a.M) != len(b.M) {
			return false
		}
		// 对象比较需要考虑键的顺序可能不同
		for _, pairA := range a.M {
			keyA := pairA[0]
			found := false
			for _, pairB := range b.M {
				if pairB[0] == keyA {
					// 解析值并比较
					valueA := &Value{}
					valueB := &Value{}
					if Parse(valueA, pairA[1]) != PARSE_OK || Parse(valueB, pairB[1]) != PARSE_OK {
						return false
					}
					if !Equal(valueA, valueB) {
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
