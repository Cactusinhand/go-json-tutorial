// JSON Pointer 实现 (RFC 6901)
package leptjson

import (
	"fmt"
	"strconv"
	"strings"
)

// JSONPointer 表示一个 JSON Pointer (RFC 6901)
type JSONPointer struct {
	Tokens []string // 路径中的令牌，例如 "/foo/bar" 会被分解为 ["foo", "bar"]
}

// PointerError 表示处理 JSON Pointer 时出现的错误
type PointerError struct {
	Path    string // 出错的路径
	Message string // 错误消息
}

// Error 实现 error 接口
func (e *PointerError) Error() string {
	return fmt.Sprintf("JSON Pointer 错误 (%s): %s", e.Path, e.Message)
}

// NewJSONPointer 从字符串创建 JSON Pointer
func NewJSONPointer(path string) (*JSONPointer, error) {
	// 检查是否为空路径，表示整个文档
	if path == "" {
		return &JSONPointer{Tokens: []string{}}, nil
	}

	// 检查是否以 '/' 开头
	if !strings.HasPrefix(path, "/") {
		return nil, &PointerError{
			Path:    path,
			Message: "JSON Pointer 必须以 '/' 开头",
		}
	}

	// 如果路径只是 "/"，则表示根文档
	if path == "/" {
		// 特殊处理根路径
		return &JSONPointer{Tokens: []string{""}}, nil
	}

	// 分割路径
	parts := strings.Split(path, "/")
	// 移除第一个空元素 (由于 /foo/bar 分割后第一个元素为空)
	parts = parts[1:]

	// 反向转义每个部分
	tokens := make([]string, len(parts))
	for i, part := range parts {
		tokens[i] = unescapeReferenceToken(part)
	}

	return &JSONPointer{Tokens: tokens}, nil
}

// Resolve 解析 JSON Pointer 指向的值
func Resolve(doc *Value, pointer *JSONPointer) (*Value, error) {
	// 空路径表示整个文档
	if len(pointer.Tokens) == 0 {
		return doc, nil
	}

	// 如果只有一个令牌且为空字符串，表示根路径 "/"
	if len(pointer.Tokens) == 1 && pointer.Tokens[0] == "" {
		return doc, nil
	}

	current := doc
	for i, token := range pointer.Tokens {
		// 跳过表示根路径的空字符串
		if i == 0 && token == "" {
			continue
		}

		// 构建当前路径用于错误消息
		currentPath := "/" + strings.Join(pointer.Tokens[:i+1], "/")

		// 处理对象
		if current.Type == OBJECT {
			found := false
			for _, m := range current.M {
				if m.K == token {
					current = m.V
					found = true
					break
				}
			}
			if !found {
				return nil, &PointerError{
					Path:    currentPath,
					Message: fmt.Sprintf("对象中不存在键 '%s'", token),
				}
			}
		} else if current.Type == ARRAY {
			// 处理数组
			index, err := parseArrayIndex(token)
			if err != nil {
				return nil, &PointerError{
					Path:    currentPath,
					Message: fmt.Sprintf("无效的数组索引: %v", err),
				}
			}

			// 检查索引是否在范围内
			if index < 0 || index >= len(current.A) {
				return nil, &PointerError{
					Path:    currentPath,
					Message: fmt.Sprintf("数组索引 %d 超出范围 (0-%d)", index, len(current.A)-1),
				}
			}

			current = current.A[index]
		} else {
			// 对于不是对象或数组的值，无法继续解析
			return nil, &PointerError{
				Path:    currentPath,
				Message: fmt.Sprintf("无法对类型 %s 使用 JSON Pointer", GetTypeStr(current.Type)),
			}
		}
	}

	return current, nil
}

// Add 添加或替换值到指定路径
func Add(doc *Value, pointer *JSONPointer, value *Value) error {
	// 空路径或根路径 "/" 表示替换整个文档
	if len(pointer.Tokens) == 0 || (len(pointer.Tokens) == 1 && pointer.Tokens[0] == "") {
		// 复制值，替换整个文档
		Copy(doc, value)
		return nil
	}

	// 获取父节点
	parentTokens := make([]string, 0)
	lastTokenIndex := len(pointer.Tokens) - 1

	// 处理父路径可能包含空字符串的情况
	for i, token := range pointer.Tokens[:lastTokenIndex] {
		if i == 0 && token == "" {
			// 跳过根路径标记
			continue
		}
		parentTokens = append(parentTokens, token)
	}

	parentPointer := &JSONPointer{Tokens: parentTokens}

	parent, err := Resolve(doc, parentPointer)
	if err != nil {
		return err
	}

	// 获取最后一个令牌，用于添加或替换
	lastToken := pointer.Tokens[lastTokenIndex]

	// 处理父节点是对象的情况
	if parent.Type == OBJECT {
		// 查找是否已存在该键
		for _, m := range parent.M {
			if m.K == lastToken {
				// 替换现有值
				Copy(m.V, value)
				return nil
			}
		}

		// 添加新键值对
		newMember := &Member{
			K: lastToken,
			V: &Value{},
		}
		Copy(newMember.V, value)
		parent.M = append(parent.M, newMember)
	} else if parent.Type == ARRAY {
		// 处理父节点是数组的情况
		// 特殊情况: "-" 表示添加到数组末尾
		if lastToken == "-" {
			newElem := &Value{}
			Copy(newElem, value)
			parent.A = append(parent.A, newElem)
			return nil
		}

		// 解析数组索引
		index, err := parseArrayIndex(lastToken)
		if err != nil {
			return &PointerError{
				Path:    "/" + strings.Join(pointer.Tokens, "/"),
				Message: fmt.Sprintf("无效的数组索引: %v", err),
			}
		}

		// 检查索引是否在范围内 (允许等于数组长度，表示追加)
		if index < 0 || index > len(parent.A) {
			return &PointerError{
				Path:    "/" + strings.Join(pointer.Tokens, "/"),
				Message: fmt.Sprintf("数组索引 %d 超出范围 (0-%d)", index, len(parent.A)),
			}
		}

		// 在索引处插入元素
		if index == len(parent.A) {
			// 追加到末尾
			newElem := &Value{}
			Copy(newElem, value)
			parent.A = append(parent.A, newElem)
		} else {
			// 在中间插入
			newElem := &Value{}
			Copy(newElem, value)
			parent.A = append(parent.A[:index+1], parent.A[index:]...)
			parent.A[index] = newElem
		}
	} else {
		// 父节点既不是对象也不是数组，无法添加子元素
		return &PointerError{
			Path:    "/" + strings.Join(parentTokens, "/"),
			Message: fmt.Sprintf("无法向类型 %s 添加元素", GetTypeStr(parent.Type)),
		}
	}

	return nil
}

// Remove 删除指定路径的值
func Remove(doc *Value, pointer *JSONPointer) error {
	// 空路径无法删除 (不能删除整个文档)
	if len(pointer.Tokens) == 0 {
		return &PointerError{
			Path:    "",
			Message: "无法删除整个文档",
		}
	}

	// 获取父节点
	parentTokens := pointer.Tokens[:len(pointer.Tokens)-1]
	parentPointer := &JSONPointer{Tokens: parentTokens}

	parent, err := Resolve(doc, parentPointer)
	if err != nil {
		return err
	}

	// 获取最后一个令牌，用于删除
	lastToken := pointer.Tokens[len(pointer.Tokens)-1]

	// 处理父节点是对象的情况
	if parent.Type == OBJECT {
		// 查找要删除的键
		for i, m := range parent.M {
			if m.K == lastToken {
				// 删除该成员
				parent.M = append(parent.M[:i], parent.M[i+1:]...)
				return nil
			}
		}

		// 找不到要删除的键
		return &PointerError{
			Path:    "/" + strings.Join(pointer.Tokens, "/"),
			Message: fmt.Sprintf("对象中不存在键 '%s'", lastToken),
		}
	} else if parent.Type == ARRAY {
		// 处理父节点是数组的情况
		// 特殊情况: "-" 不能用于删除
		if lastToken == "-" {
			return &PointerError{
				Path:    "/" + strings.Join(pointer.Tokens, "/"),
				Message: "不能使用 '-' 来删除数组元素",
			}
		}

		// 解析数组索引
		index, err := parseArrayIndex(lastToken)
		if err != nil {
			return &PointerError{
				Path:    "/" + strings.Join(pointer.Tokens, "/"),
				Message: fmt.Sprintf("无效的数组索引: %v", err),
			}
		}

		// 检查索引是否在范围内
		if index < 0 || index >= len(parent.A) {
			return &PointerError{
				Path:    "/" + strings.Join(pointer.Tokens, "/"),
				Message: fmt.Sprintf("数组索引 %d 超出范围 (0-%d)", index, len(parent.A)-1),
			}
		}

		// 删除数组元素
		parent.A = append(parent.A[:index], parent.A[index+1:]...)
	} else {
		// 父节点既不是对象也不是数组，无法删除子元素
		return &PointerError{
			Path:    "/" + strings.Join(parentTokens, "/"),
			Message: fmt.Sprintf("无法从类型 %s 删除元素", GetTypeStr(parent.Type)),
		}
	}

	return nil
}

// parseArrayIndex 解析数组索引
func parseArrayIndex(token string) (int, error) {
	// 检查是否有前导零
	if len(token) > 1 && token[0] == '0' {
		return 0, fmt.Errorf("数组索引不能有前导零")
	}

	// 尝试解析为整数
	index, err := strconv.Atoi(token)
	if err != nil {
		return 0, fmt.Errorf("无效的数组索引: %v", err)
	}

	// 检查是否为非负数
	if index < 0 {
		return 0, fmt.Errorf("数组索引不能为负数")
	}

	return index, nil
}

// unescapeReferenceToken 反转义 JSON Pointer 中的引用令牌
func unescapeReferenceToken(token string) string {
	// 替换转义序列
	token = strings.ReplaceAll(token, "~1", "/")
	token = strings.ReplaceAll(token, "~0", "~")
	return token
}

// escapeReferenceToken 转义 JSON Pointer 中的引用令牌
func escapeReferenceToken(token string) string {
	// 替换转义序列 (注意顺序很重要，先替换 ~，避免二次转义)
	token = strings.ReplaceAll(token, "~", "~0")
	token = strings.ReplaceAll(token, "/", "~1")
	return token
}

// String 返回 JSON Pointer 的字符串表示
func (p *JSONPointer) String() string {
	if len(p.Tokens) == 0 {
		return ""
	}

	// 转义并连接所有令牌
	escapedTokens := make([]string, len(p.Tokens))
	for i, token := range p.Tokens {
		escapedTokens[i] = escapeReferenceToken(token)
	}

	return "/" + strings.Join(escapedTokens, "/")
}
