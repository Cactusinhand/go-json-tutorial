package tutorial14

import (
	"fmt"
	"reflect"
)

// JSONMergePatch 表示一个 JSON Merge Patch 文档
type JSONMergePatch struct {
	Document interface{}
}

// NewJSONMergePatch 从任何有效的 JSON 文档创建 JSON Merge Patch
// 返回 JSON Merge Patch 对象和可能出现的错误
func NewJSONMergePatch(document interface{}) (*JSONMergePatch, error) {
	// 确保传入的文档是有效的 JSON 类型
	if !isValidJSON(document) {
		return nil, fmt.Errorf("无效的 JSON 文档")
	}

	return &JSONMergePatch{Document: document}, nil
}

// Apply 将该 Merge Patch 应用到目标文档，返回修改后的新文档
// 不会修改原始文档或 patch 本身
func (p *JSONMergePatch) Apply(target interface{}) (interface{}, error) {
	// 如果 patch 不是对象，则直接返回 patch 值作为结果
	if p.Document == nil {
		return nil, nil
	}

	// 如果 target 是 nil，将其视为空对象
	if target == nil {
		return p.Document, nil
	}

	// 应用合并补丁
	result, err := applyMergePatch(target, p.Document)
	if err != nil {
		return nil, fmt.Errorf("应用 Merge Patch 失败: %v", err)
	}

	return result, nil
}

// String 返回 JSON Merge Patch 的字符串表示
func (p *JSONMergePatch) String() (string, error) {
	if p.Document == nil {
		return "null", nil
	}

	bytes, err := Marshal(p.Document)
	if err != nil {
		return "", fmt.Errorf("序列化 JSON Merge Patch 失败: %v", err)
	}

	return string(bytes), nil
}

// CreateMergePatch 创建从源文档到目标文档的 JSON Merge Patch
// 返回一个表示变更的 JSONMergePatch 对象
func CreateMergePatch(source, target interface{}) (*JSONMergePatch, error) {
	// 验证源和目标文档是否是有效的 JSON
	if !isValidJSON(source) {
		return nil, fmt.Errorf("源文档不是有效的 JSON")
	}
	if !isValidJSON(target) {
		return nil, fmt.Errorf("目标文档不是有效的 JSON")
	}

	// 计算差异并创建 Merge Patch
	patch, err := generateMergePatch(source, target)
	if err != nil {
		return nil, fmt.Errorf("生成 Merge Patch 失败: %v", err)
	}

	return &JSONMergePatch{Document: patch}, nil
}

// applyMergePatch 应用 JSON Merge Patch 到目标文档
// 递归地处理对象类型
func applyMergePatch(target, patch interface{}) (interface{}, error) {
	// 如果 patch 为 nil，返回 null
	if patch == nil {
		return nil, nil
	}

	// 如果 patch 不是对象，直接返回 patch 值
	patchObj, isPatchObj := patch.(map[string]interface{})
	if !isPatchObj {
		return patch, nil
	}

	// 如果 target 是 nil 或不是对象，将其视为空对象
	targetObj, isTargetObj := target.(map[string]interface{})
	if !isTargetObj || target == nil {
		targetObj = make(map[string]interface{})
	}

	// 创建结果对象，基于 target 对象
	result := make(map[string]interface{})
	for k, v := range targetObj {
		result[k] = v
	}

	// 递归应用每个 patch 属性
	for k, v := range patchObj {
		if v == nil {
			// 如果 patch 值为 null，删除该属性
			delete(result, k)
		} else if patchValue, isPatchObjValue := v.(map[string]interface{}); isPatchObjValue {
			// 如果 patch 值是对象，递归处理
			targetValue, exists := targetObj[k]
			if !exists {
				// 如果目标中没有该属性，直接添加整个对象
				result[k] = patchValue
			} else {
				// 如果目标中已存在该属性，递归应用 patch
				newValue, err := applyMergePatch(targetValue, patchValue)
				if err != nil {
					return nil, err
				}
				if newValue == nil {
					delete(result, k)
				} else {
					result[k] = newValue
				}
			}
		} else {
			// 其他类型（包括数组）直接替换
			result[k] = v
		}
	}

	return result, nil
}

// generateMergePatch 创建从源文档到目标文档的 JSON Merge Patch
func generateMergePatch(source, target interface{}) (interface{}, error) {
	// 如果 target 为 nil，返回 null
	if target == nil {
		return nil, nil
	}

	// 如果源和目标相同，返回空对象
	if reflect.DeepEqual(source, target) {
		return make(map[string]interface{}), nil
	}

	// 如果源为 nil，或源和目标类型不同，或它们都不是对象类型，则直接返回目标值
	sourceObj, isSourceObj := source.(map[string]interface{})
	targetObj, isTargetObj := target.(map[string]interface{})
	if source == nil || !isSourceObj || !isTargetObj {
		return target, nil
	}

	// 创建 patch 对象
	patch := make(map[string]interface{})

	// 处理目标中存在但源中不存在或已更改的属性
	for k, targetValue := range targetObj {
		sourceValue, exists := sourceObj[k]
		if !exists {
			// 如果源中不存在该属性，添加到 patch 中
			patch[k] = targetValue
		} else if !reflect.DeepEqual(sourceValue, targetValue) {
			// 如果属性值不同
			sourceValueObj, isSourceValueObj := sourceValue.(map[string]interface{})
			targetValueObj, isTargetValueObj := targetValue.(map[string]interface{})
			if isSourceValueObj && isTargetValueObj {
				// 如果都是对象类型，递归计算差异
				nestedPatch, err := generateMergePatch(sourceValueObj, targetValueObj)
				if err != nil {
					return nil, err
				}
				if nestedPatch != nil && len(nestedPatch.(map[string]interface{})) > 0 {
					patch[k] = nestedPatch
				}
			} else {
				// 其他类型，直接添加目标值
				patch[k] = targetValue
			}
		}
	}

	// 处理源中存在但目标中不存在的属性（设置为 null 表示删除）
	for k := range sourceObj {
		if _, exists := targetObj[k]; !exists {
			patch[k] = nil
		}
	}

	return patch, nil
}

// isValidJSON 检查值是否是有效的 JSON 类型
func isValidJSON(v interface{}) bool {
	if v == nil {
		return true
	}

	switch v.(type) {
	case bool, float64, string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		return true
	case []interface{}:
		arr := v.([]interface{})
		for _, item := range arr {
			if !isValidJSON(item) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		obj := v.(map[string]interface{})
		for _, value := range obj {
			if !isValidJSON(value) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
