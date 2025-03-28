package tutorial15

import (
	"errors"
	"fmt"
	"strconv"
)

// ValueType 是JSON值的类型枚举
type ValueType int

const (
	NULL ValueType = iota
	BOOLEAN
	NUMBER
	STRING
	ARRAY
	OBJECT
)

// 错误定义
var (
	ErrExpectValue              = errors.New("期待值")
	ErrInvalidValue             = errors.New("无效的值")
	ErrInvalidStringEscape      = errors.New("无效的字符串转义")
	ErrInvalidStringChar        = errors.New("无效的字符串字符")
	ErrInvalidUnicodeHex        = errors.New("无效的Unicode十六进制")
	ErrInvalidUnicodeSurrogate  = errors.New("无效的Unicode代理对")
	ErrMissQuotationMark        = errors.New("缺少引号")
	ErrMissCommaOrSquareBracket = errors.New("缺少逗号或方括号")
	ErrMissCommaOrCurlyBracket  = errors.New("缺少逗号或花括号")
	ErrMissColon                = errors.New("缺少冒号")
	ErrMissKey                  = errors.New("缺少键")
	ErrInvalidKeyType           = errors.New("无效的键类型")
	ErrInvalidNumberFormat      = errors.New("无效的数字格式")
)

// Member 表示JSON对象的一个成员
type Member struct {
	Key   string
	Value *Value
}

// Value 表示一个JSON值
type Value struct {
	Type     ValueType
	B        bool     // 布尔值
	N        float64  // 数值
	S        string   // 字符串
	Elements []*Value // 数组元素
	Members  []Member // 对象成员
}

// SetNull 设置为null值
func SetNull(v *Value) {
	v.Type = NULL
}

// SetBool 设置为布尔值
func SetBool(v *Value, b bool) {
	v.Type = BOOLEAN
	v.B = b
}

// SetNumber 设置为数值
func SetNumber(v *Value, n float64) {
	v.Type = NUMBER
	v.N = n
}

// SetString 设置为字符串值
func SetString(v *Value, s string) {
	v.Type = STRING
	v.S = s
}

// SetArray 设置为数组值
func SetArray(v *Value) {
	v.Type = ARRAY
	v.Elements = make([]*Value, 0)
}

// SetObject 设置为对象值
func SetObject(v *Value) {
	v.Type = OBJECT
	v.Members = make([]Member, 0)
}

// GetType 获取值类型
func GetType(v *Value) ValueType {
	return v.Type
}

// GetBool 获取布尔值
func GetBool(v *Value) bool {
	return v.B
}

// GetNumber 获取数值
func GetNumber(v *Value) float64 {
	return v.N
}

// GetString 获取字符串值
func GetString(v *Value) string {
	return v.S
}

// GetArraySize 获取数组大小
func GetArraySize(v *Value) int {
	if v.Type != ARRAY {
		return 0
	}
	return len(v.Elements)
}

// GetArrayElement 获取数组元素
func GetArrayElement(v *Value, index int) *Value {
	if v.Type != ARRAY || index < 0 || index >= len(v.Elements) {
		return nil
	}
	return v.Elements[index]
}

// PushBackArrayElement 向数组尾部添加元素
func PushBackArrayElement(v *Value, element *Value) {
	if v.Type != ARRAY {
		return
	}
	v.Elements = append(v.Elements, element)
}

// GetObjectSize 获取对象成员数量
func GetObjectSize(v *Value) int {
	if v.Type != OBJECT {
		return 0
	}
	return len(v.Members)
}

// GetObjectKey 获取对象的键
func GetObjectKey(v *Value, index int) string {
	if v.Type != OBJECT || index < 0 || index >= len(v.Members) {
		return ""
	}
	return v.Members[index].Key
}

// GetObjectValue 获取对象的值
func GetObjectValue(v *Value, index int) *Value {
	if v.Type != OBJECT || index < 0 || index >= len(v.Members) {
		return nil
	}
	return v.Members[index].Value
}

// SetObjectValue 设置对象属性
func SetObjectValue(v *Value, key string) *Value {
	if v.Type != OBJECT {
		return nil
	}

	// 检查键是否已存在
	for i := 0; i < len(v.Members); i++ {
		if v.Members[i].Key == key {
			return v.Members[i].Value
		}
	}

	// 添加新成员
	value := &Value{}
	v.Members = append(v.Members, Member{Key: key, Value: value})
	return value
}

// FindObjectValue 根据键查找对象值
func FindObjectValue(v *Value, key string) *Value {
	if v.Type != OBJECT {
		return nil
	}

	for i := 0; i < len(v.Members); i++ {
		if v.Members[i].Key == key {
			return v.Members[i].Value
		}
	}
	return nil
}

// FindObjectKey 根据键查找对象的成员，返回成员索引和是否找到
func FindObjectKey(v *Value, key string) (int, bool) {
	if v == nil || v.Type != OBJECT {
		return -1, false
	}

	for i, member := range v.Members {
		if member.Key == key {
			return i, true
		}
	}
	return -1, false
}

// GetTypeStr 获取JSON值类型的字符串表示
func GetTypeStr(v *Value) string {
	if v == nil {
		return "NULL"
	}

	switch v.Type {
	case NULL:
		return "NULL"
	case BOOLEAN:
		return "BOOLEAN"
	case NUMBER:
		return "NUMBER"
	case STRING:
		return "STRING"
	case ARRAY:
		return "ARRAY"
	case OBJECT:
		return "OBJECT"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", v.Type)
	}
}

// Equal 比较两个 JSON 值是否相等
func Equal(a, b *Value) bool {
	if a.Type != b.Type {
		return false
	}

	switch a.Type {
	case NULL:
		return true
	case BOOLEAN:
		return a.B == b.B
	case NUMBER:
		return a.N == b.N
	case STRING:
		return a.S == b.S
	case ARRAY:
		if len(a.Elements) != len(b.Elements) {
			return false
		}
		for i := 0; i < len(a.Elements); i++ {
			if !Equal(a.Elements[i], b.Elements[i]) {
				return false
			}
		}
		return true
	case OBJECT:
		if len(a.Members) != len(b.Members) {
			return false
		}
		// 简单实现，不考虑成员顺序
		for i := 0; i < len(a.Members); i++ {
			found := false
			for j := 0; j < len(b.Members); j++ {
				if a.Members[i].Key == b.Members[j].Key {
					found = true
					if !Equal(a.Members[i].Value, b.Members[j].Value) {
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

// Parse 解析JSON字符串
func Parse(v *Value, json string) error {
	parser := &Parser{json: json, pos: 0}
	parser.parseWhitespace()

	err := parser.parseValue(v)
	if err != nil {
		return err
	}

	parser.parseWhitespace()
	if parser.pos < len(parser.json) {
		return ErrInvalidValue
	}

	return nil
}

// Parser JSON解析器
type Parser struct {
	json string
	pos  int
}

// 解析空白字符
func (p *Parser) parseWhitespace() {
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
func (p *Parser) parseValue(v *Value) error {
	if p.pos >= len(p.json) {
		return ErrExpectValue
	}

	switch p.json[p.pos] {
	case 'n':
		return p.parseLiteral(v, "null", NULL)
	case 't':
		return p.parseLiteral(v, "true", BOOLEAN)
	case 'f':
		return p.parseLiteral(v, "false", BOOLEAN)
	case '"':
		return p.parseString(v)
	case '[':
		return p.parseArray(v)
	case '{':
		return p.parseObject(v)
	default:
		return p.parseNumber(v)
	}
}

// 解析字面量
func (p *Parser) parseLiteral(v *Value, literal string, vtype ValueType) error {
	if len(p.json)-p.pos < len(literal) {
		return ErrInvalidValue
	}

	for i := 0; i < len(literal); i++ {
		if p.json[p.pos+i] != literal[i] {
			return ErrInvalidValue
		}
	}

	p.pos += len(literal)

	switch vtype {
	case NULL:
		SetNull(v)
	case BOOLEAN:
		SetBool(v, literal == "true")
	}

	return nil
}

// 解析字符串
func (p *Parser) parseString(v *Value) error {
	var s string

	// 跳过开始的引号
	p.pos++

	start := p.pos
	for p.pos < len(p.json) {
		ch := p.json[p.pos]

		if ch == '"' {
			// 找到了结束的引号
			if v != nil {
				SetString(v, s+p.json[start:p.pos])
			}
			p.pos++
			return nil
		} else if ch == '\\' {
			// 转义字符
			s += p.json[start:p.pos]
			p.pos++
			if p.pos >= len(p.json) {
				return ErrInvalidStringEscape
			}

			switch p.json[p.pos] {
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
				// 简单转义
				switch p.json[p.pos] {
				case 'b':
					s += "\b"
				case 'f':
					s += "\f"
				case 'n':
					s += "\n"
				case 'r':
					s += "\r"
				case 't':
					s += "\t"
				default:
					s += string(p.json[p.pos])
				}
				p.pos++

			case 'u':
				// Unicode转义
				p.pos++ // 跳过 'u'
				if p.pos+4 > len(p.json) {
					return ErrInvalidUnicodeHex
				}

				var codepoint int
				for i := 0; i < 4; i++ {
					ch := p.json[p.pos+i]
					codepoint <<= 4
					if ch >= '0' && ch <= '9' {
						codepoint |= int(ch - '0')
					} else if ch >= 'A' && ch <= 'F' {
						codepoint |= int(ch - 'A' + 10)
					} else if ch >= 'a' && ch <= 'f' {
						codepoint |= int(ch - 'a' + 10)
					} else {
						return ErrInvalidUnicodeHex
					}
				}
				p.pos += 4

				// 处理UTF-16代理对
				if codepoint >= 0xD800 && codepoint <= 0xDBFF {
					// 高代理项，需要后续的低代理项
					if p.pos+6 > len(p.json) || p.json[p.pos] != '\\' || p.json[p.pos+1] != 'u' {
						return ErrInvalidUnicodeSurrogate
					}

					p.pos += 2 // 跳过 '\u'

					var codepoint2 int
					for i := 0; i < 4; i++ {
						ch := p.json[p.pos+i]
						codepoint2 <<= 4
						if ch >= '0' && ch <= '9' {
							codepoint2 |= int(ch - '0')
						} else if ch >= 'A' && ch <= 'F' {
							codepoint2 |= int(ch - 'A' + 10)
						} else if ch >= 'a' && ch <= 'f' {
							codepoint2 |= int(ch - 'a' + 10)
						} else {
							return ErrInvalidUnicodeHex
						}
					}
					p.pos += 4

					if codepoint2 < 0xDC00 || codepoint2 > 0xDFFF {
						return ErrInvalidUnicodeSurrogate
					}

					// 计算完整的代码点
					codepoint = 0x10000 + ((codepoint - 0xD800) << 10) | (codepoint2 - 0xDC00)
				}

				// 将码点转换为UTF-8
				if codepoint <= 0x7F {
					s += string(rune(codepoint))
				} else if codepoint <= 0x7FF {
					s += string([]byte{
						byte(0xC0 | ((codepoint >> 6) & 0xFF)),
						byte(0x80 | (codepoint & 0x3F)),
					})
				} else if codepoint <= 0xFFFF {
					s += string([]byte{
						byte(0xE0 | ((codepoint >> 12) & 0xFF)),
						byte(0x80 | ((codepoint >> 6) & 0x3F)),
						byte(0x80 | (codepoint & 0x3F)),
					})
				} else {
					s += string([]byte{
						byte(0xF0 | ((codepoint >> 18) & 0xFF)),
						byte(0x80 | ((codepoint >> 12) & 0x3F)),
						byte(0x80 | ((codepoint >> 6) & 0x3F)),
						byte(0x80 | (codepoint & 0x3F)),
					})
				}

			default:
				return ErrInvalidStringEscape
			}

			start = p.pos
		} else if ch < 0x20 {
			// 控制字符不允许直接出现在字符串中
			return ErrInvalidStringChar
		} else {
			p.pos++
		}
	}

	return ErrMissQuotationMark
}

// 解析数字
func (p *Parser) parseNumber(v *Value) error {
	start := p.pos

	// 负号
	if p.pos < len(p.json) && p.json[p.pos] == '-' {
		p.pos++
	}

	// 整数部分
	if p.pos < len(p.json) && p.json[p.pos] == '0' {
		p.pos++
	} else if p.pos < len(p.json) && p.json[p.pos] >= '1' && p.json[p.pos] <= '9' {
		p.pos++
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	} else {
		return ErrInvalidNumberFormat
	}

	// 小数部分
	if p.pos < len(p.json) && p.json[p.pos] == '.' {
		p.pos++
		if p.pos >= len(p.json) || p.json[p.pos] < '0' || p.json[p.pos] > '9' {
			return ErrInvalidNumberFormat
		}
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	}

	// 指数部分
	if p.pos < len(p.json) && (p.json[p.pos] == 'e' || p.json[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.json) && (p.json[p.pos] == '+' || p.json[p.pos] == '-') {
			p.pos++
		}
		if p.pos >= len(p.json) || p.json[p.pos] < '0' || p.json[p.pos] > '9' {
			return ErrInvalidNumberFormat
		}
		for p.pos < len(p.json) && p.json[p.pos] >= '0' && p.json[p.pos] <= '9' {
			p.pos++
		}
	}

	// 转换为数字
	if v != nil {
		n, err := strconv.ParseFloat(p.json[start:p.pos], 64)
		if err != nil {
			return ErrInvalidNumberFormat
		}
		SetNumber(v, n)
	}

	return nil
}

// 解析数组
func (p *Parser) parseArray(v *Value) error {
	// 跳过开始的'['
	p.pos++

	// 设置为数组
	SetArray(v)

	p.parseWhitespace()
	if p.pos < len(p.json) && p.json[p.pos] == ']' {
		p.pos++
		return nil
	}

	for {
		element := &Value{}
		err := p.parseValue(element)
		if err != nil {
			return err
		}

		PushBackArrayElement(v, element)

		p.parseWhitespace()
		if p.pos < len(p.json) && p.json[p.pos] == ',' {
			p.pos++
			p.parseWhitespace()
		} else if p.pos < len(p.json) && p.json[p.pos] == ']' {
			p.pos++
			return nil
		} else {
			return ErrMissCommaOrSquareBracket
		}
	}
}

// 解析对象
func (p *Parser) parseObject(v *Value) error {
	// 跳过开始的'{'
	p.pos++

	// 设置为对象
	SetObject(v)

	p.parseWhitespace()
	if p.pos < len(p.json) && p.json[p.pos] == '}' {
		p.pos++
		return nil
	}

	for {
		// 解析键
		p.parseWhitespace()
		if p.pos >= len(p.json) || p.json[p.pos] != '"' {
			return ErrMissKey
		}

		// 解析键名
		var key string
		keyStart := p.pos + 1
		for p.pos++; p.pos < len(p.json); p.pos++ {
			if p.json[p.pos] == '"' {
				key = p.json[keyStart:p.pos]
				p.pos++
				break
			}
		}
		if key == "" {
			return ErrMissKey
		}

		// 解析冒号
		p.parseWhitespace()
		if p.pos >= len(p.json) || p.json[p.pos] != ':' {
			return ErrMissColon
		}
		p.pos++

		// 解析值
		p.parseWhitespace()
		value := &Value{}
		if err := p.parseValue(value); err != nil {
			return err
		}

		// 添加成员
		v.Members = append(v.Members, Member{Key: key, Value: value})

		// 解析逗号或结束
		p.parseWhitespace()
		if p.pos < len(p.json) && p.json[p.pos] == ',' {
			p.pos++
			p.parseWhitespace()
		} else if p.pos < len(p.json) && p.json[p.pos] == '}' {
			p.pos++
			return nil
		} else {
			return ErrMissCommaOrCurlyBracket
		}
	}
}
