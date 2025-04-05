package leptjson

// JSONMergePatch 表示一个 JSON Merge Patch 文档
type JSONMergePatch struct {
	Document *Value
}

// NewJSONMergePatch 从 *Value 创建 JSON Merge Patch 对象
// 注意：Merge Patch 本身必须是有效的 JSON，由调用方保证
func NewJSONMergePatch(patchDoc *Value) (*JSONMergePatch, error) {
	// Merge Patch可以是任何JSON类型，所以这里不需要额外验证
	// 但如果需要可以添加检查，例如确保非空的补丁至少是 Object 或 Null？
	// RFC 7396: The patch is itself a JSON document.
	return &JSONMergePatch{Document: patchDoc}, nil
}

// Apply 将该 Merge Patch 应用到目标文档 (*Value)
// 返回修改后的新文档 (*Value)，不修改原始文档或 patch 本身。
func (p *JSONMergePatch) Apply(target *Value) (*Value, error) {
	// 创建目标文档的深拷贝，以避免修改原始文档
	result := &Value{}
	Copy(result, target)

	// 应用合并补丁到副本
	err := applyMergePatchRecursive(result, p.Document)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// String 返回 JSON Merge Patch 的字符串表示
func (p *JSONMergePatch) String() (string, error) {
	s, errCode := Stringify(p.Document)
	if errCode != STRINGIFY_OK {
		return "", errCode
	}
	return s, nil
}

// CreateMergePatch 创建从源文档到目标文档的 JSON Merge Patch (*Value)
// 返回一个表示变更的 JSONMergePatch 对象
func CreateMergePatch(source, target *Value) (*JSONMergePatch, error) {
	patchDoc := generateMergePatchRecursive(source, target)

	return &JSONMergePatch{Document: patchDoc}, nil
}

// applyMergePatchRecursive 递归地应用 JSON Merge Patch 到目标文档 (*Value)
// 直接修改目标 target Value。
func applyMergePatchRecursive(target *Value, patch *Value) error {
	if GetType(patch) != OBJECT {
		Copy(target, patch)
		return nil
	}

	if GetType(target) != OBJECT {
		Copy(target, patch)
		return nil
	}

	patchSize := GetObjectSize(patch)
	for i := 0; i < patchSize; i++ {
		key := GetObjectKey(patch, i)
		patchValue := GetObjectValue(patch, i)

		if GetType(patchValue) == NULL {
			RemoveObjectValueByKey(target, key)
		} else {
			targetValue := GetObjectValueByKey(target, key)

			if targetValue == nil {
				valueCopy := &Value{}
				Copy(valueCopy, patchValue)
				targetValuePtr := SetObjectValue(target, key)
				Copy(targetValuePtr, valueCopy)
			} else {
				if GetType(targetValue) == OBJECT && GetType(patchValue) == OBJECT {
					err := applyMergePatchRecursive(targetValue, patchValue)
					if err != nil {
						return err
					}
				} else {
					valueCopy := &Value{}
					Copy(valueCopy, patchValue)
					targetValuePtr := SetObjectValue(target, key)
					Copy(targetValuePtr, valueCopy)
				}
			}
		}
	}
	return nil
}

// generateMergePatchRecursive 递归地创建从源文档到目标文档的 JSON Merge Patch (*Value)
func generateMergePatchRecursive(source, target *Value) *Value {
	if GetType(target) == NULL {
		return &Value{Type: NULL}
	}

	if Equal(source, target) {
		patch := &Value{}
		SetObject(patch)
		return patch
	}

	if GetType(source) != OBJECT || GetType(target) != OBJECT {
		patch := &Value{}
		Copy(patch, target)
		return patch
	}

	patch := &Value{}
	SetObject(patch)

	targetSize := GetObjectSize(target)
	for i := 0; i < targetSize; i++ {
		key := GetObjectKey(target, i)
		targetValue := GetObjectValue(target, i)
		sourceValue := GetObjectValueByKey(source, key)

		if sourceValue == nil {
			valueCopy := &Value{}
			Copy(valueCopy, targetValue)
			patchValuePtr := SetObjectValue(patch, key)
			Copy(patchValuePtr, valueCopy)
		} else {
			if GetType(sourceValue) == OBJECT && GetType(targetValue) == OBJECT {
				nestedPatch := generateMergePatchRecursive(sourceValue, targetValue)
				if GetObjectSize(nestedPatch) > 0 {
					patchValuePtr := SetObjectValue(patch, key)
					Copy(patchValuePtr, nestedPatch)
				}
			} else if !Equal(sourceValue, targetValue) {
				valueCopy := &Value{}
				Copy(valueCopy, targetValue)
				patchValuePtr := SetObjectValue(patch, key)
				Copy(patchValuePtr, valueCopy)
			}
		}
	}

	sourceSize := GetObjectSize(source)
	for i := 0; i < sourceSize; i++ {
		key := GetObjectKey(source, i)
		if GetObjectValueByKey(target, key) == nil {
			targetPtr := SetObjectValue(patch, key)
			SetNull(targetPtr)
		}
	}

	return patch
}

// RemoveObjectValueByKey 是一个辅助函数，需要添加到 leptjson.go 或在此处实现
// 它根据键来查找并删除对象成员
func RemoveObjectValueByKey(v *Value, key string) bool {
	if GetType(v) != OBJECT {
		return false
	}
	index := FindObjectIndex(v, key)
	if index != -1 {
		RemoveObjectValue(v, index)
		return true
	}
	return false
}
