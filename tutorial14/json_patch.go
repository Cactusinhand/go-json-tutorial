// json_patch.go - JSON Patch 实现 (RFC 6902)
package leptjson

import (
	"fmt"
	"strings"
)

// PatchOperation 表示 JSON Patch 中的单个操作
type PatchOperation struct {
	Op    string // 操作类型: add, remove, replace, move, copy, test
	Path  string // 操作的目标路径 (JSON Pointer)
	From  string // 源路径 (用于 move 和 copy 操作)
	Value *Value // 值 (用于 add, replace 和 test 操作)
}

// PatchError 表示 JSON Patch 操作中的错误
type PatchError struct {
	Operation string // 发生错误的操作类型
	Path      string // 发生错误的路径
	Message   string // 错误消息
}

// Error 实现 error 接口
func (e PatchError) Error() string {
	return fmt.Sprintf("JSON Patch 错误 (%s %s): %s", e.Operation, e.Path, e.Message)
}

// JSONPatch 表示一个 JSON Patch 文档，包含多个操作
type JSONPatch struct {
	Operations []PatchOperation // 操作列表
}

// NewJSONPatch 从 JSON 值中创建 JSON Patch 对象
func NewJSONPatch(patchDoc *Value) (*JSONPatch, error) {
	// 检查 patchDoc 是否是一个数组
	if GetType(patchDoc) != ARRAY {
		return nil, &PatchError{
			Operation: "parse",
			Path:      "",
			Message:   "JSON Patch 必须是一个数组",
		}
	}

	patch := &JSONPatch{
		Operations: make([]PatchOperation, 0, len(patchDoc.A)),
	}

	// 解析每个操作
	for i, opVal := range patchDoc.A {
		if GetType(opVal) != OBJECT {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   "JSON Patch 操作必须是一个对象",
			}
		}

		op := PatchOperation{}

		// 获取操作类型 (op)
		opTypeVal := GetObjectValueByKey(opVal, "op")
		if opTypeVal == nil || GetType(opTypeVal) != STRING {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   "缺少有效的 'op' 字段",
			}
		}
		op.Op = GetString(opTypeVal)

		// 检查操作类型是否有效
		if !isValidOperation(op.Op) {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   fmt.Sprintf("无效的操作类型: %s", op.Op),
			}
		}

		// 获取路径 (path)
		pathVal := GetObjectValueByKey(opVal, "path")
		if pathVal == nil || GetType(pathVal) != STRING {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   "缺少有效的 'path' 字段",
			}
		}
		op.Path = GetString(pathVal)

		// 对于 move 和 copy 操作，获取 from 路径
		if op.Op == "move" || op.Op == "copy" {
			fromVal := GetObjectValueByKey(opVal, "from")
			if fromVal == nil || GetType(fromVal) != STRING {
				return nil, &PatchError{
					Operation: op.Op,
					Path:      op.Path,
					Message:   "缺少有效的 'from' 字段",
				}
			}
			op.From = GetString(fromVal)
		}

		// 对于 add, replace 和 test 操作，获取 value
		if op.Op == "add" || op.Op == "replace" || op.Op == "test" {
			valueVal := GetObjectValueByKey(opVal, "value")
			if valueVal == nil {
				return nil, &PatchError{
					Operation: op.Op,
					Path:      op.Path,
					Message:   "缺少 'value' 字段",
				}
			}
			// 复制值，避免引用原始文档
			op.Value = &Value{}
			Copy(op.Value, valueVal)
		}

		patch.Operations = append(patch.Operations, op)
	}

	return patch, nil
}

// NewJSONPatchFromString 从 JSON 字符串创建 JSON Patch 对象
func NewJSONPatchFromString(patchStr string) (*JSONPatch, error) {
	// 解析 JSON 字符串
	patchDoc := &Value{}
	errCode := Parse(patchDoc, patchStr)
	if errCode != PARSE_OK {
		// 假设 GetErrorMessage 仍然存在且可用
		// 如果 leptjson.go 中的错误处理机制变化，这里需要调整
		return nil, &PatchError{
			Operation: "parse",
			Path:      "",
			Message:   fmt.Sprintf("无法解析 JSON Patch 文档: %s", GetErrorMessage(errCode)),
		}
	}

	return NewJSONPatch(patchDoc)
}

// Apply 将 JSON Patch 应用到文档
func (p *JSONPatch) Apply(doc *Value) error {
	// 应用每个操作
	for i, op := range p.Operations {
		if err := applyOperation(doc, &op); err != nil {
			// 将错误包装成 PatchError，并添加操作索引
			if patchErr, ok := err.(*PatchError); ok {
				patchErr.Message = fmt.Sprintf("操作 %d: %s", i, patchErr.Message)
			}
			return err
		}
	}
	return nil
}

// applyOperation 应用单个 Patch 操作到文档
func applyOperation(doc *Value, op *PatchOperation) error {
	// 特殊处理替换整个文档的情况
	if op.Op == "replace" && (op.Path == "" || op.Path == "/") {
		// 使用 Copy 函数来替换整个文档
		Copy(doc, op.Value)
		return nil
	}

	switch op.Op {
	case "add":
		return applyAddOperation(doc, op)
	case "remove":
		return applyRemoveOperation(doc, op)
	case "replace":
		return applyReplaceOperation(doc, op)
	case "move":
		return applyMoveOperation(doc, op)
	case "copy":
		return applyCopyOperation(doc, op)
	case "test":
		return applyTestOperation(doc, op)
	default:
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   "不支持的操作类型",
		}
	}
}

// applyAddOperation 实现 add 操作
func applyAddOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, errCode := ParseJSONPointer(op.Path)
	if errCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %s", errCode.Error()),
		}
	}

	// 使用 Insert 方法
	if errCode = pointer.Insert(doc, op.Value); errCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("添加操作失败: %s", errCode.Error()),
		}
	}
	return nil
}

// applyRemoveOperation 实现 remove 操作
func applyRemoveOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, errCode := ParseJSONPointer(op.Path)
	if errCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %s", errCode.Error()),
		}
	}

	// 移除值
	if errCode = pointer.Remove(doc); errCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("移除操作失败: %s", errCode.Error()),
		}
	}
	return nil
}

// applyReplaceOperation 实现 replace 操作
func applyReplaceOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, errCode := ParseJSONPointer(op.Path)
	if errCode != POINTER_OK {
		return &PatchError{Operation: op.Op, Path: op.Path, Message: fmt.Sprintf("无效的 JSON Pointer: %s", errCode.Error())}
	}

	// 尝试获取当前值，确保路径有效 (RFC 6902 要求替换目标必须存在)
	_, getErrCode := pointer.Get(doc)
	if getErrCode != POINTER_OK {
		return &PatchError{Operation: op.Op, Path: op.Path, Message: fmt.Sprintf("替换目标不存在: %s", getErrCode.Error())}
	}

	// 使用 Replace 方法
	setErrCode := pointer.Replace(doc, op.Value)
	if setErrCode != POINTER_OK {
		return &PatchError{Operation: op.Op, Path: op.Path, Message: fmt.Sprintf("替换操作失败: %s", setErrCode.Error())}
	}
	return nil
}

// applyMoveOperation 实现 move 操作
func applyMoveOperation(doc *Value, op *PatchOperation) error {
	// 解析源路径
	fromPointer, fromErrCode := ParseJSONPointer(op.From)
	if fromErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无效的源 JSON Pointer: %s", fromErrCode.Error()),
		}
	}

	// 获取要移动的值
	valueToMove, getErrCode := fromPointer.Get(doc)
	if getErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("移动源不存在: %s", getErrCode.Error()),
		}
	}

	// 创建值的副本以防在移除和添加过程中出现问题
	copiedValue := &Value{}
	Copy(copiedValue, valueToMove)

	// 移除源位置的值
	if removeErrCode := fromPointer.Remove(doc); removeErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("移除移动源失败: %s", removeErrCode.Error()),
		}
	}

	// 解析目标路径
	toPointer, toErrCode := ParseJSONPointer(op.Path)
	if toErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的目标 JSON Pointer: %s", toErrCode.Error()),
		}
	}

	// 将复制的值添加到目标位置 (根据 RFC 语义，使用 Insert)
	if setErrCode := toPointer.Insert(doc, copiedValue); setErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("添加移动值到目标失败: %s", setErrCode.Error()),
		}
	}

	return nil
}

// applyCopyOperation 实现 copy 操作
func applyCopyOperation(doc *Value, op *PatchOperation) error {
	// 解析源路径
	fromPointer, fromErrCode := ParseJSONPointer(op.From)
	if fromErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无效的源 JSON Pointer: %s", fromErrCode.Error()),
		}
	}

	// 获取要复制的值
	valueToCopy, getErrCode := fromPointer.Get(doc)
	if getErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("复制源不存在: %s", getErrCode.Error()),
		}
	}

	// 创建值的副本
	copiedValue := &Value{}
	Copy(copiedValue, valueToCopy)

	// 解析目标路径
	toPointer, toErrCode := ParseJSONPointer(op.Path)
	if toErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的目标 JSON Pointer: %s", toErrCode.Error()),
		}
	}

	// 将复制的值添加到目标位置 (使用 Insert 语义)
	if setErrCode := toPointer.Insert(doc, copiedValue); setErrCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("添加到目标位置失败: %s", setErrCode.Error()),
		}
	}

	return nil
}

// applyTestOperation 实现 test 操作
func applyTestOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, errCode := ParseJSONPointer(op.Path)
	if errCode != POINTER_OK {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %s", errCode.Error()),
		}
	}

	// 获取目标路径的值
	targetValue, getErrCode := pointer.Get(doc)
	if getErrCode != POINTER_OK {
		// 如果 test 操作的值是 null，并且目标路径不存在，则测试通过
		// （RFC 6902 中对此情况的处理有歧义，这里采用一种常见解释：目标不存在不等于null）
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("测试目标不存在: %s", getErrCode.Error()),
		}
	}

	// 比较目标值和操作中的值是否深度相等
	if !Equal(targetValue, op.Value) {
		targetStr, _ := Stringify(targetValue)
		valueStr, _ := Stringify(op.Value)
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("测试失败：目标值 %s 不等于预期值 %s", targetStr, valueStr),
		}
	}

	return nil
}

// isValidOperation 检查操作类型是否有效
func isValidOperation(op string) bool {
	switch op {
	case "add", "remove", "replace", "move", "copy", "test":
		return true
	default:
		return false
	}
}

// CreatePatch 生成从 source 到 target 的 JSON Patch
func CreatePatch(source, target *Value) (*JSONPatch, error) {
	patch := &JSONPatch{Operations: make([]PatchOperation, 0)}
	diff(source, target, "", patch)
	return patch, nil
}

// diff 递归比较两个值并生成 patch 操作
func diff(source, target *Value, path string, patch *JSONPatch) {
	// 如果源和目标完全相同，则无需操作
	if Equal(source, target) {
		return
	}

	// 如果源不存在 (或者类型不同，视为替换)
	if source == nil || source.Type != target.Type {
		// 添加 replace 操作
		copiedTarget := &Value{}
		Copy(copiedTarget, target)
		patch.Operations = append(patch.Operations, PatchOperation{
			Op:    "replace",
			Path:  path,
			Value: copiedTarget,
		})
		return
	}

	switch target.Type {
	case ARRAY:
		diffArray(source, target, path, patch)
	case OBJECT:
		diffObject(source, target, path, patch)
	default:
		// 对于基本类型 (null, bool, number, string)，如果类型相同但不相等，则替换
		copiedTarget := &Value{}
		Copy(copiedTarget, target)
		patch.Operations = append(patch.Operations, PatchOperation{
			Op:    "replace",
			Path:  path,
			Value: copiedTarget,
		})
	}
}

// diffArray 比较两个数组并生成 patch 操作
func diffArray(source, target *Value, path string, patch *JSONPatch) {
	// 比较每个元素
	maxLen := len(source.A)
	if len(target.A) > maxLen {
		maxLen = len(target.A)
	}

	for i := 0; i < maxLen; i++ {
		itemPath := fmt.Sprintf("%s/%d", path, i)
		var srcItem, tgtItem *Value
		if i < len(source.A) {
			srcItem = source.A[i]
		}
		if i < len(target.A) {
			tgtItem = target.A[i]
		}

		if srcItem == nil && tgtItem != nil {
			// target 中有，source 中没有 -> add
			copiedItem := &Value{}
			Copy(copiedItem, tgtItem)
			patch.Operations = append(patch.Operations, PatchOperation{
				Op:    "add",
				Path:  itemPath,
				Value: copiedItem,
			})
		} else if srcItem != nil && tgtItem == nil {
			// source 中有，target 中没有 -> remove (需要从后向前移除避免索引变化)
			// 这个逻辑在 diff 函数外处理，先标记为需要移除
			// (优化：这里可以直接添加 remove 操作，但在循环中移除会改变索引)
		} else if srcItem != nil && tgtItem != nil {
			// 两边都有，递归比较
			diff(srcItem, tgtItem, itemPath, patch)
		}
	}

	// 处理 source 中多余的元素 (需要从后向前移除)
	if len(source.A) > len(target.A) {
		for i := len(source.A) - 1; i >= len(target.A); i-- {
			itemPath := fmt.Sprintf("%s/%d", path, i)
			patch.Operations = append(patch.Operations, PatchOperation{
				Op:   "remove",
				Path: itemPath,
			})
		}
	}
}

// diffObject 比较两个对象并生成 patch 操作
func diffObject(source, target *Value, path string, patch *JSONPatch) {
	sourceKeys := make(map[string]*Value)
	for _, m := range source.O {
		sourceKeys[m.K] = m.V
	}

	targetKeys := make(map[string]*Value)
	for _, m := range target.O {
		targetKeys[m.K] = m.V
	}

	// 检查 target 中的键
	for key, tgtVal := range targetKeys {
		keyPath := fmt.Sprintf("%s/%s", path, escapeJSONPointerToken(key))
		if srcVal, exists := sourceKeys[key]; exists {
			// 如果键在 source 和 target 中都存在，递归比较值
			diff(srcVal, tgtVal, keyPath, patch)
		} else {
			// 如果键只在 target 中存在 -> add
			copiedVal := &Value{}
			Copy(copiedVal, tgtVal)
			patch.Operations = append(patch.Operations, PatchOperation{
				Op:    "add",
				Path:  keyPath,
				Value: copiedVal,
			})
		}
	}

	// 检查 source 中的键，如果 target 中不存在 -> remove
	for key := range sourceKeys {
		if _, exists := targetKeys[key]; !exists {
			keyPath := fmt.Sprintf("%s/%s", path, escapeJSONPointerToken(key))
			patch.Operations = append(patch.Operations, PatchOperation{
				Op:   "remove",
				Path: keyPath,
			})
		}
	}
}

// String 返回 JSON Patch 的字符串表示
func (p *JSONPatch) String() (string, error) {
	patchDoc := &Value{Type: ARRAY, A: make([]*Value, 0, len(p.Operations))}

	for _, op := range p.Operations {
		opVal := &Value{}
		SetObject(opVal)

		// 设置 op
		SetString(SetObjectValue(opVal, "op"), op.Op)
		// 设置 path
		SetString(SetObjectValue(opVal, "path"), op.Path)

		if op.Op == "move" || op.Op == "copy" {
			SetString(SetObjectValue(opVal, "from"), op.From)
		}

		if op.Op == "add" || op.Op == "replace" || op.Op == "test" {
			valueCopy := &Value{}
			Copy(valueCopy, op.Value) // 复制一份值
			// SetObjectValue 返回的是指向新（或现有）Value 的指针，需要用 Copy 赋值
			Copy(SetObjectValue(opVal, "value"), valueCopy)
		}

		patchDoc.A = append(patchDoc.A, opVal)
	}

	// return Stringify(patchDoc)
	s, errCode := Stringify(patchDoc)
	if errCode != STRINGIFY_OK {
		// StringifyError 实现了 error 接口，可以直接返回
		return "", errCode
	}
	return s, nil // 成功时返回 nil error
}

// escapeJSONPointerToken 转义 JSON Pointer token 中的特殊字符
func escapeJSONPointerToken(token string) string {
	escaped := strings.ReplaceAll(token, "~", "~0")
	escaped = strings.ReplaceAll(escaped, "/", "~1")
	return escaped
}
