// json_pointer.go - JSON指针实现 (RFC6901)
package leptjson

import (
	"fmt"
	"strconv"
	"strings"
)

// JSONPointerError 表示JSON指针相关错误
type JSONPointerError int

// JSON指针错误常量
const (
	POINTER_OK JSONPointerError = iota
	POINTER_INVALID_FORMAT
	POINTER_INDEX_OUT_OF_RANGE
	POINTER_KEY_NOT_FOUND
	POINTER_INVALID_TARGET
)

// JSONPointer 表示一个JSON指针（RFC6901）
type JSONPointer struct {
	tokens []string // 路径令牌
}

// 实现 Error 接口
func (e JSONPointerError) Error() string {
	switch e {
	case POINTER_OK:
		return "JSON指针操作成功"
	case POINTER_INVALID_FORMAT:
		return "无效的JSON指针格式"
	case POINTER_INDEX_OUT_OF_RANGE:
		return "数组索引超出范围"
	case POINTER_KEY_NOT_FOUND:
		return "对象中未找到指定的键"
	case POINTER_INVALID_TARGET:
		return "无效的目标类型"
	default:
		return "未知的JSON指针错误"
	}
}

// ParseJSONPointer 解析JSON指针字符串
// 例如: "/foo/0/bar" => ["foo", "0", "bar"]
func ParseJSONPointer(pointer string) (*JSONPointer, JSONPointerError) {
	// 空字符串表示整个文档
	if pointer == "" {
		return &JSONPointer{tokens: []string{}}, POINTER_OK
	}

	// 必须以 '/' 开头
	if !strings.HasPrefix(pointer, "/") {
		return nil, POINTER_INVALID_FORMAT
	}

	// 分割路径
	parts := strings.Split(pointer[1:], "/")
	tokens := make([]string, len(parts))

	// 处理转义字符
	for i, part := range parts {
		// 转义处理: ~1 => /, ~0 => ~
		unescaped := strings.ReplaceAll(part, "~1", "/")
		unescaped = strings.ReplaceAll(unescaped, "~0", "~")
		tokens[i] = unescaped
	}

	return &JSONPointer{tokens: tokens}, POINTER_OK
}

// Get 根据JSON指针获取值
func (p *JSONPointer) Get(root *Value) (*Value, JSONPointerError) {
	// 空指针直接返回根节点
	if len(p.tokens) == 0 {
		return root, POINTER_OK
	}

	current := root
	for _, token := range p.tokens {
		switch current.Type {
		case ARRAY:
			// 对于数组，token 必须是有效索引
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(current.A) {
				return nil, POINTER_INDEX_OUT_OF_RANGE
			}
			current = current.A[index]

		case OBJECT:
			// 对于对象，查找匹配的键
			found := false
			for _, member := range current.O {
				if member.K == token {
					current = member.V
					found = true
					break
				}
			}
			if !found {
				return nil, POINTER_KEY_NOT_FOUND
			}

		default:
			// 其他类型无法继续遍历
			return nil, POINTER_INVALID_TARGET
		}
	}

	return current, POINTER_OK
}

// Set 根据JSON指针设置值
func (p *JSONPointer) Set(root *Value, value *Value) JSONPointerError {
	// 特殊情况：空指针，替换整个文档
	if len(p.tokens) == 0 {
		Copy(root, value)
		return POINTER_OK
	}

	// 指针至少有一个token
	parent, err := p.getParent(root)
	if err != POINTER_OK {
		return err
	}

	lastToken := p.tokens[len(p.tokens)-1]

	switch parent.Type {
	case ARRAY:
		index, err := strconv.Atoi(lastToken)
		if err != nil || index < 0 || index >= len(parent.A) {
			return POINTER_INDEX_OUT_OF_RANGE
		}
		// 复制值
		Copy(parent.A[index], value)

	case OBJECT:
		// 查找对象成员
		for i, member := range parent.O {
			if member.K == lastToken {
				// 复制值
				Copy(parent.O[i].V, value)
				return POINTER_OK
			}
		}

		// 没找到键，创建新成员
		newValue := &Value{}
		Copy(newValue, value)
		parent.O = append(parent.O, Member{K: lastToken, V: newValue})

	default:
		return POINTER_INVALID_TARGET
	}

	return POINTER_OK
}

// Remove 根据JSON指针删除值
func (p *JSONPointer) Remove(root *Value) JSONPointerError {
	// 不能删除根节点
	if len(p.tokens) == 0 {
		return POINTER_INVALID_TARGET
	}

	parent, err := p.getParent(root)
	if err != POINTER_OK {
		return err
	}

	lastToken := p.tokens[len(p.tokens)-1]

	switch parent.Type {
	case ARRAY:
		index, err := strconv.Atoi(lastToken)
		if err != nil || index < 0 || index >= len(parent.A) {
			return POINTER_INDEX_OUT_OF_RANGE
		}
		// 删除数组元素
		EraseArrayElement(parent, index, 1)

	case OBJECT:
		// 查找对象成员索引
		for i, member := range parent.O {
			if member.K == lastToken {
				// 删除成员
				RemoveObjectValue(parent, i)
				return POINTER_OK
			}
		}
		return POINTER_KEY_NOT_FOUND

	default:
		return POINTER_INVALID_TARGET
	}

	return POINTER_OK
}

// 获取指针路径上的倒数第二个节点（父节点）
func (p *JSONPointer) getParent(root *Value) (*Value, JSONPointerError) {
	if len(p.tokens) <= 0 {
		return nil, POINTER_INVALID_FORMAT
	}

	// 如果只有一个token，父节点就是根节点
	if len(p.tokens) == 1 {
		return root, POINTER_OK
	}

	// 创建一个不包含最后一个token的指针
	parentPointer := &JSONPointer{
		tokens: p.tokens[:len(p.tokens)-1],
	}

	return parentPointer.Get(root)
}

// 创建一个JSON指针字符串表示
func (p *JSONPointer) String() string {
	if len(p.tokens) == 0 {
		return ""
	}

	var parts []string
	for _, token := range p.tokens {
		// 转义处理: ~ => ~0, / => ~1
		escaped := strings.ReplaceAll(token, "~", "~0")
		escaped = strings.ReplaceAll(escaped, "/", "~1")
		parts = append(parts, escaped)
	}

	return "/" + strings.Join(parts, "/")
}

// GetJSONPointer 创建一个指向指定路径的JSONPointer
// 例如: NewJSONPointer("foo", 0, "bar") => "/foo/0/bar"
func GetJSONPointer(segments ...interface{}) (*JSONPointer, error) {
	tokens := make([]string, len(segments))

	for i, segment := range segments {
		switch v := segment.(type) {
		case string:
			tokens[i] = v
		case int:
			tokens[i] = strconv.Itoa(v)
		default:
			return nil, fmt.Errorf("不支持的路径段类型: %T", segment)
		}
	}

	return &JSONPointer{tokens: tokens}, nil
}
