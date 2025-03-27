// JSON Pointer 实现 (RFC 6901)
package tutorial14

import (
	"fmt"
	"strconv"
	"strings"
)

// JSONPointer 表示 JSON Pointer (RFC 6901)
type JSONPointer struct {
	Tokens []string
}

// NewJSONPointer 创建一个新的 JSON Pointer
func NewJSONPointer(path string) (*JSONPointer, error) {
	// 处理特殊情况：空字符串表示整个文档
	if path == "" {
		return &JSONPointer{Tokens: []string{}}, nil
	}

	// 处理特殊情况：根路径
	if path == "/" {
		return &JSONPointer{Tokens: []string{""}}, nil
	}

	// 必须以 / 开头
	if !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("JSON Pointer 错误: 路径必须以 '/' 开头")
	}

	// 分割路径
	if len(path) == 1 {
		// 仅有一个 '/'
		return &JSONPointer{Tokens: []string{""}}, nil
	}

	tokens := strings.Split(path[1:], "/")
	// 解码每个 token
	for i, token := range tokens {
		// 先替换 ~1，再替换 ~0
		token = strings.ReplaceAll(token, "~1", "/")
		token = strings.ReplaceAll(token, "~0", "~")
		tokens[i] = token
	}

	return &JSONPointer{Tokens: tokens}, nil
}

// String 返回 JSON Pointer 的字符串表示
func (p *JSONPointer) String() string {
	if len(p.Tokens) == 0 {
		return ""
	}

	var result strings.Builder
	for _, token := range p.Tokens {
		result.WriteString("/")
		// 编码 ~ 和 /
		encoded := strings.ReplaceAll(token, "~", "~0")
		encoded = strings.ReplaceAll(encoded, "/", "~1")
		result.WriteString(encoded)
	}
	return result.String()
}

// Resolve 解析 JSON Pointer 对应的值
func (p *JSONPointer) Resolve(document interface{}) (interface{}, error) {
	// 特殊情况：空路径返回整个文档
	if len(p.Tokens) == 0 || (len(p.Tokens) == 1 && p.Tokens[0] == "") {
		return document, nil
	}

	current := document
	for i, token := range p.Tokens {
		if current == nil {
			return nil, fmt.Errorf("JSON Pointer 错误: 空值无法继续解析路径")
		}

		// 处理对象
		if obj, isObj := current.(map[string]interface{}); isObj {
			val, exists := obj[token]
			if !exists {
				return nil, fmt.Errorf("JSON Pointer 错误: 键 '%s' 在对象中不存在", token)
			}
			current = val
			continue
		}

		// 处理数组
		if arr, isArr := current.([]interface{}); isArr {
			// 检查 token 是否是有效的数组索引
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 {
				return nil, fmt.Errorf("JSON Pointer 错误: 无效的数组索引 '%s'", token)
			}
			if index >= len(arr) {
				return nil, fmt.Errorf("JSON Pointer 错误: 数组索引越界 '%d'", index)
			}
			current = arr[index]
			continue
		}

		// 如果既不是对象也不是数组，则无法继续解析
		return nil, fmt.Errorf("JSON Pointer 错误: 无法在非对象/非数组值上继续解析路径，位于 token '%s' (索引 %d)", token, i)
	}

	return current, nil
}

// Add 在指定路径添加值
func (p *JSONPointer) Add(document interface{}, value interface{}) (interface{}, error) {
	// 空路径或根路径，直接替换整个文档
	if len(p.Tokens) == 0 || (len(p.Tokens) == 1 && p.Tokens[0] == "") {
		return value, nil
	}

	// 复制文档以避免修改原始文档
	result, err := deepCopy(document)
	if err != nil {
		return nil, err
	}

	// 获取父路径
	parentPath := &JSONPointer{Tokens: p.Tokens[:len(p.Tokens)-1]}
	lastToken := p.Tokens[len(p.Tokens)-1]

	// 获取父对象
	parent, err := parentPath.Resolve(result)
	if err != nil {
		return nil, err
	}

	// 在父对象上设置值
	if obj, isObj := parent.(map[string]interface{}); isObj {
		obj[lastToken] = value
		return result, nil
	}

	// 处理数组
	if arr, isArr := parent.([]interface{}); isArr {
		// 检查 lastToken 是否是有效的数组索引
		index, err := strconv.Atoi(lastToken)
		if err != nil || index < 0 {
			return nil, fmt.Errorf("JSON Pointer 错误: 无效的数组索引 '%s'", lastToken)
		}

		// 特殊情况："-" 表示追加到数组末尾
		if lastToken == "-" {
			newArr := append(arr, value)
			// 更新父路径上的数组
			return updateParentWithNewValue(result, parentPath, newArr)
		}

		// 检查索引是否越界
		if index > len(arr) {
			return nil, fmt.Errorf("JSON Pointer 错误: 数组索引越界 '%d'", index)
		}

		// 在指定位置插入
		if index == len(arr) {
			// 追加到末尾
			newArr := append(arr, value)
			return updateParentWithNewValue(result, parentPath, newArr)
		} else {
			// 在中间插入
			newArr := make([]interface{}, len(arr)+1)
			copy(newArr, arr[:index])
			newArr[index] = value
			copy(newArr[index+1:], arr[index:])
			return updateParentWithNewValue(result, parentPath, newArr)
		}
	}

	return nil, fmt.Errorf("JSON Pointer 错误: 无法在非对象/非数组值上添加属性")
}

// deepCopy 创建对象的深拷贝
func deepCopy(src interface{}) (interface{}, error) {
	if src == nil {
		return nil, nil
	}

	switch v := src.(type) {
	case map[string]interface{}:
		dst := make(map[string]interface{}, len(v))
		for key, value := range v {
			copy, err := deepCopy(value)
			if err != nil {
				return nil, err
			}
			dst[key] = copy
		}
		return dst, nil
	case []interface{}:
		dst := make([]interface{}, len(v))
		for i, value := range v {
			copy, err := deepCopy(value)
			if err != nil {
				return nil, err
			}
			dst[i] = copy
		}
		return dst, nil
	default:
		// 基本类型可以直接返回
		return v, nil
	}
}

// updateParentWithNewValue 更新父路径上的值
func updateParentWithNewValue(document interface{}, parentPath *JSONPointer, newValue interface{}) (interface{}, error) {
	// 如果是根路径，直接返回新值
	if len(parentPath.Tokens) == 0 {
		return newValue, nil
	}

	// 创建一个新的指针，指向父路径的父级
	grandParentPath := &JSONPointer{Tokens: parentPath.Tokens[:len(parentPath.Tokens)-1]}
	lastToken := parentPath.Tokens[len(parentPath.Tokens)-1]

	// 获取祖父对象
	grandParent, err := grandParentPath.Resolve(document)
	if err != nil {
		return nil, err
	}

	// 在祖父对象上更新父对象
	if obj, isObj := grandParent.(map[string]interface{}); isObj {
		obj[lastToken] = newValue
		return document, nil
	}

	// 处理数组
	if arr, isArr := grandParent.([]interface{}); isArr {
		// 检查 lastToken 是否是有效的数组索引
		index, err := strconv.Atoi(lastToken)
		if err != nil || index < 0 || index >= len(arr) {
			return nil, fmt.Errorf("JSON Pointer 错误: 无效的数组索引 '%s'", lastToken)
		}

		// 更新指定位置的元素
		arr[index] = newValue
		return document, nil
	}

	return nil, fmt.Errorf("JSON Pointer 错误: 无法在非对象/非数组值上更新属性")
}
