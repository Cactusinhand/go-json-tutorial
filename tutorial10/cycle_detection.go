// cycle_detection.go - 循环引用检测实现
package leptjson

import (
	"strconv"
)

// CycleError 表示循环引用错误
type CycleError int

// 循环引用错误常量
const (
	CYCLE_OK CycleError = iota
	CYCLE_DETECTED
)

// 实现 Error 接口
func (e CycleError) Error() string {
	switch e {
	case CYCLE_OK:
		return "未检测到循环引用"
	case CYCLE_DETECTED:
		return "检测到循环引用"
	default:
		return "未知的循环引用错误"
	}
}

// pathNode 表示JSON值路径中的一个节点
type pathNode struct {
	Value *Value    // JSON值指针
	Next  *pathNode // 链表下一个节点
}

// pathStack 表示一个简单的栈，用于存储遍历路径
type pathStack struct {
	Head *pathNode // 栈顶节点
}

// 入栈操作
func (s *pathStack) Push(v *Value) {
	s.Head = &pathNode{Value: v, Next: s.Head}
}

// 出栈操作
func (s *pathStack) Pop() {
	if s.Head != nil {
		s.Head = s.Head.Next
	}
}

// 检查值是否在栈中（循环检测的核心）
func (s *pathStack) Contains(v *Value) bool {
	for node := s.Head; node != nil; node = node.Next {
		// 通过内存地址比较指针
		if node.Value == v {
			return true
		}
	}
	return false
}

// DetectCycle 检测JSON值中是否存在循环引用
func DetectCycle(v *Value) CycleError {
	stack := &pathStack{}
	return detectCycleInternal(v, stack)
}

// 内部递归函数，用于检测循环引用
func detectCycleInternal(v *Value, stack *pathStack) CycleError {
	// 只有数组和对象可能有循环引用
	if v.Type != ARRAY && v.Type != OBJECT {
		return CYCLE_OK
	}

	// 检查当前值是否已经在路径中（通过内存地址比较）
	// 这是检测循环引用的关键
	for node := stack.Head; node != nil; node = node.Next {
		if node.Value == v {
			return CYCLE_DETECTED
		}
	}

	// 将当前值加入路径
	stack.Push(v)

	// 根据类型检查子元素
	switch v.Type {
	case ARRAY:
		for _, elem := range v.A {
			if err := detectCycleInternal(elem, stack); err != CYCLE_OK {
				return err
			}
		}

	case OBJECT:
		for _, member := range v.O {
			if err := detectCycleInternal(member.V, stack); err != CYCLE_OK {
				return err
			}
		}
	}

	// 回溯，将值从路径中移除
	stack.Pop()

	return CYCLE_OK
}

// SafeCopy 安全复制JSON值，检测并处理循环引用
func SafeCopy(dst, src *Value) CycleError {
	// 首先检测源值是否有循环引用
	if err := DetectCycle(src); err != CYCLE_OK {
		return err
	}

	// 无循环引用，进行常规复制
	Copy(dst, src)
	return CYCLE_OK
}

// CircularReplacer 定义了在发现循环引用时的替换函数类型
type CircularReplacer func(path []string) *Value

// SafeCopyWithReplacer 带替换器的安全复制，处理循环引用
func SafeCopyWithReplacer(dst, src *Value, replacer CircularReplacer) {
	// 创建访问路径记录
	path := []string{}

	// 使用替换器进行安全复制
	safeCopyWithPath(dst, src, &path, make(map[*Value]bool), replacer)
}

// 内部函数，带路径信息的安全复制
func safeCopyWithPath(dst, src *Value, path *[]string, visited map[*Value]bool, replacer CircularReplacer) {
	// 基本类型直接复制
	if src.Type != ARRAY && src.Type != OBJECT {
		Copy(dst, src)
		return
	}

	// 检查是否已访问过该节点（检测循环）
	if visited[src] {
		// 使用替换器生成替代值
		replacement := replacer(*path)
		Copy(dst, replacement)
		return
	}

	// 标记为已访问
	visited[src] = true

	// 根据类型进行复制
	switch src.Type {
	case ARRAY:
		// 设置目标为数组
		SetArray(dst, len(src.A))

		// 复制数组元素
		for i, elem := range src.A {
			// 更新路径
			indexStr := strconv.Itoa(i)
			*path = append(*path, indexStr)

			// 复制元素
			elemValue := &Value{}
			safeCopyWithPath(elemValue, elem, path, visited, replacer)
			dst.A[i] = elemValue

			// 回溯路径
			*path = (*path)[:len(*path)-1]
		}

	case OBJECT:
		// 设置目标为对象
		SetObject(dst)

		// 复制对象成员
		for _, member := range src.O {
			// 更新路径
			*path = append(*path, member.K)

			// 复制成员值
			valuePtr := SetObjectValue(dst, member.K)
			safeCopyWithPath(valuePtr, member.V, path, visited, replacer)

			// 回溯路径
			*path = (*path)[:len(*path)-1]
		}
	}

	// 移除访问标记（允许同一值在不同路径中出现）
	delete(visited, src)
}
