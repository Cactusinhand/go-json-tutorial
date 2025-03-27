// json_patch.go - JSON Patch 实现 (RFC 6902)
package leptjson

import (
	"fmt"
	"strconv"
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
	if patchDoc.Type != ARRAY {
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
		if opVal.Type != OBJECT {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   "JSON Patch 操作必须是一个对象",
			}
		}

		op := PatchOperation{}

		// 获取操作类型 (op)
		opTypeVal, found := FindObjectKey(opVal, "op")
		if !found || opTypeVal.Type != STRING {
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
		pathVal, found := FindObjectKey(opVal, "path")
		if !found || pathVal.Type != STRING {
			return nil, &PatchError{
				Operation: "parse",
				Path:      fmt.Sprintf("[%d]", i),
				Message:   "缺少有效的 'path' 字段",
			}
		}
		op.Path = GetString(pathVal)

		// 对于 move 和 copy 操作，获取 from 路径
		if op.Op == "move" || op.Op == "copy" {
			fromVal, found := FindObjectKey(opVal, "from")
			if !found || fromVal.Type != STRING {
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
			valueVal, found := FindObjectKey(opVal, "value")
			if !found {
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
	err := Parse(patchDoc, patchStr)
	if err != PARSE_OK {
		return nil, &PatchError{
			Operation: "parse",
			Path:      "",
			Message:   fmt.Sprintf("无法解析 JSON Patch 文档: %s", GetErrorMessage(err)),
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
	if op.Op == "replace" && op.Path == "/" {
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
	pointer, err := NewJSONPointer(op.Path)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %v", err),
		}
	}

	// 添加值
	return Add(doc, pointer, op.Value)
}

// applyRemoveOperation 实现 remove 操作
func applyRemoveOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, err := NewJSONPointer(op.Path)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %v", err),
		}
	}

	// 移除值
	return Remove(doc, pointer)
}

// applyReplaceOperation 实现 replace 操作
func applyReplaceOperation(doc *Value, op *PatchOperation) error {
	// replace 可以实现为 remove 然后 add
	if err := applyRemoveOperation(doc, op); err != nil {
		return err
	}
	return applyAddOperation(doc, op)
}

// applyMoveOperation 实现 move 操作
func applyMoveOperation(doc *Value, op *PatchOperation) error {
	// 解析源路径
	fromPointer, err := NewJSONPointer(op.From)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无效的源 JSON Pointer: %v", err),
		}
	}

	// 解析目标路径
	toPointer, err := NewJSONPointer(op.Path)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的目标 JSON Pointer: %v", err),
		}
	}

	// 检查源路径是否是目标路径的前缀 (不能移动到自己的子节点)
	if isPrefix(op.From, op.Path) {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   "不能移动节点到其子节点",
		}
	}

	// 获取要移动的值
	value, err := Resolve(doc, fromPointer)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无法解析源路径: %v", err),
		}
	}

	// 复制值
	tempValue := &Value{}
	Copy(tempValue, value)

	// 删除源位置的值
	if err := Remove(doc, fromPointer); err != nil {
		return err
	}

	// 添加到目标位置
	return Add(doc, toPointer, tempValue)
}

// applyCopyOperation 实现 copy 操作
func applyCopyOperation(doc *Value, op *PatchOperation) error {
	// 解析源路径
	fromPointer, err := NewJSONPointer(op.From)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无效的源 JSON Pointer: %v", err),
		}
	}

	// 解析目标路径
	toPointer, err := NewJSONPointer(op.Path)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的目标 JSON Pointer: %v", err),
		}
	}

	// 获取要复制的值
	value, err := Resolve(doc, fromPointer)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.From,
			Message:   fmt.Sprintf("无法解析源路径: %v", err),
		}
	}

	// 复制值
	tempValue := &Value{}
	Copy(tempValue, value)

	// 添加到目标位置
	return Add(doc, toPointer, tempValue)
}

// applyTestOperation 实现 test 操作
func applyTestOperation(doc *Value, op *PatchOperation) error {
	// 解析 JSON Pointer
	pointer, err := NewJSONPointer(op.Path)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无效的 JSON Pointer: %v", err),
		}
	}

	// 获取要测试的值
	value, err := Resolve(doc, pointer)
	if err != nil {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   fmt.Sprintf("无法解析路径: %v", err),
		}
	}

	// 检查值是否相等
	if !Equal(value, op.Value) {
		return &PatchError{
			Operation: op.Op,
			Path:      op.Path,
			Message:   "测试失败: 值不相等",
		}
	}

	return nil
}

// isValidOperation 检查操作类型是否有效
func isValidOperation(op string) bool {
	validOps := map[string]bool{
		"add":     true,
		"remove":  true,
		"replace": true,
		"move":    true,
		"copy":    true,
		"test":    true,
	}
	return validOps[op]
}

// isPrefix 检查 path1 是否是 path2 的前缀
func isPrefix(path1, path2 string) bool {
	// 格式化路径，确保它们都以 '/' 开头
	if !strings.HasPrefix(path1, "/") {
		path1 = "/" + path1
	}
	if !strings.HasPrefix(path2, "/") {
		path2 = "/" + path2
	}

	// 检查 path2 是否以 path1 开头
	return path2 != path1 && strings.HasPrefix(path2, path1)
}

// CreatePatch 生成从 source 到 target 的 JSON Patch
func CreatePatch(source, target *Value) *JSONPatch {
	patch := &JSONPatch{
		Operations: []PatchOperation{},
	}

	// 使用 JSON 指针 "/" 表示文档根
	rootPointer, _ := NewJSONPointer("")

	// 递归比较两个文档
	compareValues(source, target, rootPointer, patch)

	return patch
}

// compareValues 递归比较两个 JSON 值，并生成对应的补丁操作
func compareValues(source, target *Value, pointer *JSONPointer, patch *JSONPatch) {
	// 如果类型不同，直接替换
	if source.Type != target.Type {
		addReplacePatchOperation(source, target, pointer, patch)
		return
	}

	// 根据类型进行比较
	switch source.Type {
	case NULL, TRUE, FALSE:
		// 对于基本类型，只需检查是否相等
		if source.Type != target.Type {
			addReplacePatchOperation(source, target, pointer, patch)
		}
	case NUMBER:
		// 比较数值
		if source.N != target.N {
			addReplacePatchOperation(source, target, pointer, patch)
		}
	case STRING:
		// 比较字符串
		if source.S != target.S {
			addReplacePatchOperation(source, target, pointer, patch)
		}
	case ARRAY:
		compareArrays(source, target, pointer, patch)
	case OBJECT:
		compareObjects(source, target, pointer, patch)
	}
}

// compareArrays 比较两个数组，并生成对应的补丁操作
func compareArrays(source, target *Value, pointer *JSONPointer, patch *JSONPatch) {
	sourceLen := len(source.A)
	targetLen := len(target.A)

	// 使用最长公共子序列算法会更高效，但这里我们使用简单的方法
	// 1. 先比较数组中的共同元素
	minLen := sourceLen
	if targetLen < sourceLen {
		minLen = targetLen
	}

	// 比较共同部分
	for i := 0; i < minLen; i++ {
		childPointer := extendPointer(pointer, strconv.Itoa(i))
		compareValues(source.A[i], target.A[i], childPointer, patch)
	}

	// 2. 处理数组长度变化
	if sourceLen > targetLen {
		// 需要删除多余的元素
		// 注意：从后向前删除，避免索引变化问题
		for i := sourceLen - 1; i >= targetLen; i-- {
			childPointer := extendPointer(pointer, strconv.Itoa(i))
			addRemovePatchOperation(childPointer, patch)
		}
	} else if targetLen > sourceLen {
		// 需要添加新元素
		for i := sourceLen; i < targetLen; i++ {
			childPointer := extendPointer(pointer, strconv.Itoa(i))
			addAddPatchOperation(target.A[i], childPointer, patch)
		}
	}
}

// compareObjects 比较两个对象，并生成对应的补丁操作
func compareObjects(source, target *Value, pointer *JSONPointer, patch *JSONPatch) {
	// 创建源对象的键映射，用于快速查找
	sourceKeys := make(map[string]struct{})
	sourceValues := make(map[string]*Value)

	for _, member := range source.M {
		sourceKeys[member.K] = struct{}{}
		sourceValues[member.K] = member.V
	}

	// 创建目标对象的键映射
	targetKeys := make(map[string]struct{})

	// 1. 检查目标对象的每个键
	for _, member := range target.M {
		targetKeys[member.K] = struct{}{}

		childPointer := extendPointer(pointer, member.K)

		// 键在源对象中存在
		if sourceValue, ok := sourceValues[member.K]; ok {
			// 比较值
			compareValues(sourceValue, member.V, childPointer, patch)
		} else {
			// 键在源对象中不存在，需要添加
			addAddPatchOperation(member.V, childPointer, patch)
		}
	}

	// 2. 检查源对象中存在但目标对象中不存在的键（需要删除）
	for _, member := range source.M {
		if _, ok := targetKeys[member.K]; !ok {
			// 键在目标对象中不存在，需要删除
			childPointer := extendPointer(pointer, member.K)
			addRemovePatchOperation(childPointer, patch)
		}
	}
}

// 辅助函数：扩展 JSON 指针
func extendPointer(pointer *JSONPointer, token string) *JSONPointer {
	tokens := append([]string{}, pointer.Tokens...)
	tokens = append(tokens, token)
	return &JSONPointer{Tokens: tokens}
}

// 辅助函数：添加 replace 操作
func addReplacePatchOperation(source, target *Value, pointer *JSONPointer, patch *JSONPatch) {
	pointerStr := pointer.String()

	// 如果是空路径，我们不能使用 replace，需要完全替换文档
	if pointerStr == "" {
		// 对于根路径，使用 /
		pointerStr = "/"
	}

	patch.Operations = append(patch.Operations, PatchOperation{
		Op:    "replace",
		Path:  pointerStr,
		Value: target,
	})
}

// 辅助函数：添加 add 操作
func addAddPatchOperation(value *Value, pointer *JSONPointer, patch *JSONPatch) {
	patch.Operations = append(patch.Operations, PatchOperation{
		Op:    "add",
		Path:  pointer.String(),
		Value: value,
	})
}

// 辅助函数：添加 remove 操作
func addRemovePatchOperation(pointer *JSONPointer, patch *JSONPatch) {
	patch.Operations = append(patch.Operations, PatchOperation{
		Op:   "remove",
		Path: pointer.String(),
	})
}

// String 返回 JSON Patch 的字符串表示
func (p *JSONPatch) String() (string, error) {
	// 创建一个 JSON 数组
	patchDoc := &Value{}
	SetArray(patchDoc, len(p.Operations))

	// 添加每个操作
	for _, op := range p.Operations {
		// 创建操作对象
		opObj := PushBackArrayElement(patchDoc)
		SetObject(opObj)

		// 添加 op 字段
		opField := SetObjectValue(opObj, "op")
		SetString(opField, op.Op)

		// 添加 path 字段
		pathField := SetObjectValue(opObj, "path")
		SetString(pathField, op.Path)

		// 对于 move 和 copy 操作，添加 from 字段
		if op.Op == "move" || op.Op == "copy" {
			fromField := SetObjectValue(opObj, "from")
			SetString(fromField, op.From)
		}

		// 对于 add, replace 和 test 操作，添加 value 字段
		if op.Op == "add" || op.Op == "replace" || op.Op == "test" {
			valueField := SetObjectValue(opObj, "value")
			Copy(valueField, op.Value)
		}
	}

	// 将 JSON 数组转换为字符串
	return Stringify(patchDoc)
}
