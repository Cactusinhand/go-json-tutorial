// json_path.go - JSON Path 实现
package leptjson

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// JSONPathError 表示解析或执行 JSON Path 时的错误
type JSONPathError struct {
	Path    string // JSON Path 表达式
	Message string // 错误消息
	Index   int    // 错误发生的位置
}

// Error 实现 error 接口
func (e JSONPathError) Error() string {
	return fmt.Sprintf("JSON Path 错误 '%s' 在位置 %d: %s", e.Path, e.Index, e.Message)
}

// TokenType 表示 JSON Path 令牌的类型
type TokenType int

const (
	ROOT              TokenType = iota // $ - 根节点
	CURRENT                            // @ - 当前节点
	DOT                                // . - 子属性访问
	RECURSIVE_DESCENT                  // .. - 递归下降
	WILDCARD                           // * - 通配符
	BRACKET_START                      // [ - 下标访问开始
	BRACKET_END                        // ] - 下标访问结束
	INDEX                              // 数字索引
	PROPERTY                           // 属性名
	SLICE                              // 切片 [start:end:step]
	UNION                              // 并集 [expr,expr]
	FILTER                             // ?() - 过滤器
)

// Token 表示 JSON Path 中的一个令牌
type Token struct {
	Type  TokenType // 令牌类型
	Value string    // 令牌值
}

// SliceInfo 存储数组切片信息
type SliceInfo struct {
	Start int
	End   int
	Step  int
}

// JSONPath 表示一个解析后的 JSON Path 表达式
type JSONPath struct {
	Path   string  // 原始路径表达式
	Tokens []Token // 令牌列表
}

// NewJSONPath 解析 JSON Path 表达式并创建一个 JSONPath 对象
func NewJSONPath(path string) (*JSONPath, error) {
	jp := &JSONPath{
		Path: path,
	}

	if err := jp.parse(); err != nil {
		return nil, err
	}

	return jp, nil
}

// parse 将 JSON Path 表达式解析为令牌列表
func (jp *JSONPath) parse() error {
	// 基本验证
	if len(jp.Path) == 0 {
		return &JSONPathError{Path: jp.Path, Message: "JSON Path 表达式不能为空"}
	}

	// 确保路径以 $ 开头
	if !strings.HasPrefix(jp.Path, "$") {
		return &JSONPathError{Path: jp.Path, Message: "JSON Path 必须以 $ 开始"}
	}

	// 将字符串解析为令牌
	jp.Tokens = []Token{{Type: ROOT, Value: "$"}}

	i := 1 // 跳过 $
	for i < len(jp.Path) {
		switch jp.Path[i] {
		case '.':
			if i+1 < len(jp.Path) && jp.Path[i+1] == '.' {
				// 递归下降 ..
				jp.Tokens = append(jp.Tokens, Token{Type: RECURSIVE_DESCENT, Value: ".."})
				i += 2
			} else {
				// 子属性访问 .
				jp.Tokens = append(jp.Tokens, Token{Type: DOT, Value: "."})
				i++
			}

			// 处理属性名或通配符
			if i < len(jp.Path) {
				if jp.Path[i] == '*' {
					jp.Tokens = append(jp.Tokens, Token{Type: WILDCARD, Value: "*"})
					i++
				} else if isValidPropertyNameStart(jp.Path[i]) {
					propName, newPos := parsePropertyName(jp.Path, i)
					jp.Tokens = append(jp.Tokens, Token{Type: PROPERTY, Value: propName})
					i = newPos
				}
			}

		case '[':
			jp.Tokens = append(jp.Tokens, Token{Type: BRACKET_START, Value: "["})
			i++

			// 跳过空白
			for i < len(jp.Path) && unicode.IsSpace(rune(jp.Path[i])) {
				i++
			}

			// 处理数组索引、切片或属性名
			if i < len(jp.Path) {
				if jp.Path[i] == '*' {
					// 通配符 [*]
					jp.Tokens = append(jp.Tokens, Token{Type: WILDCARD, Value: "*"})
					i++
				} else if jp.Path[i] == '\'' || jp.Path[i] == '"' {
					// 带引号的属性名 ["name"] 或 ['name']
					quote := jp.Path[i]
					i++ // 跳过引号
					start := i

					for i < len(jp.Path) && jp.Path[i] != quote {
						i++
					}

					if i >= len(jp.Path) {
						return &JSONPathError{
							Path:    jp.Path,
							Message: "未闭合的引号",
							Index:   start - 1,
						}
					}

					jp.Tokens = append(jp.Tokens, Token{
						Type:  PROPERTY,
						Value: jp.Path[start:i],
					})
					i++ // 跳过结束引号
				} else if unicode.IsDigit(rune(jp.Path[i])) || jp.Path[i] == '-' {
					// 数字索引或切片
					if isSliceNotation(jp.Path[i:]) {
						// 切片 [start:end:step]
						slice, newPos, err := parseSlice(jp.Path, i)
						if err != nil {
							return err
						}
						jp.Tokens = append(jp.Tokens, Token{
							Type:  SLICE,
							Value: slice,
						})
						i = newPos
					} else {
						// 单个索引
						index, newPos, err := parseIndex(jp.Path, i)
						if err != nil {
							return err
						}
						jp.Tokens = append(jp.Tokens, Token{
							Type:  INDEX,
							Value: fmt.Sprintf("%d", index),
						})
						i = newPos
					}
				}
			}

			// 跳过空白
			for i < len(jp.Path) && unicode.IsSpace(rune(jp.Path[i])) {
				i++
			}

			// 确保有 ]
			if i >= len(jp.Path) || jp.Path[i] != ']' {
				return &JSONPathError{
					Path:    jp.Path,
					Message: "未闭合的方括号",
					Index:   i,
				}
			}

			jp.Tokens = append(jp.Tokens, Token{Type: BRACKET_END, Value: "]"})
			i++

		default:
			i++ // 跳过不识别的字符
		}
	}

	return nil
}

// 辅助函数：判断是否是有效的属性名起始字符
func isValidPropertyNameStart(c byte) bool {
	return unicode.IsLetter(rune(c)) || c == '_' || c == '$'
}

// 辅助函数：判断是否是有效的属性名字符
func isValidPropertyNameChar(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c)) || c == '_' || c == '$'
}

// 辅助函数：解析属性名
func parsePropertyName(path string, start int) (string, int) {
	i := start
	for i < len(path) && isValidPropertyNameChar(path[i]) {
		i++
	}
	return path[start:i], i
}

// 辅助函数：判断是否是切片表示法 [start:end:step]
func isSliceNotation(s string) bool {
	// 如果长度不足，不可能是切片表示法
	if len(s) < 1 {
		return false
	}

	// 检查是否包含冒号
	return strings.Contains(s, ":")
}

// 辅助函数：解析切片表示法
func parseSlice(path string, start int) (string, int, error) {
	// 找到结束括号
	end := start
	bracketCount := 0

	for end < len(path) {
		if path[end] == '[' {
			bracketCount++
		} else if path[end] == ']' {
			if bracketCount == 0 {
				break
			}
			bracketCount--
		}
		end++
	}

	if end >= len(path) || path[end] != ']' {
		return "", start, &JSONPathError{
			Path:    path,
			Message: "切片表示法未闭合",
			Index:   start,
		}
	}

	sliceStr := path[start:end]
	return sliceStr, end, nil
}

// 辅助函数：解析数组索引
func parseIndex(path string, start int) (int, int, error) {
	i := start

	// 处理负数索引
	negative := false
	if path[i] == '-' {
		negative = true
		i++
	}

	// 解析数字
	numStart := i
	for i < len(path) && unicode.IsDigit(rune(path[i])) {
		i++
	}

	if numStart == i {
		return 0, start, &JSONPathError{
			Path:    path,
			Message: "无效的数组索引",
			Index:   start,
		}
	}

	num, err := strconv.Atoi(path[numStart:i])
	if err != nil {
		return 0, start, &JSONPathError{
			Path:    path,
			Message: "无效的数组索引",
			Index:   start,
		}
	}

	if negative {
		num = -num
	}

	return num, i, nil
}

// 解析切片参数 (开始:结束:步长)
func parseSliceParams(sliceStr string) (*SliceInfo, error) {
	// 默认值
	result := &SliceInfo{
		Start: 0,
		End:   -1, // -1 表示到数组末尾
		Step:  1,
	}

	parts := strings.Split(sliceStr, ":")
	if len(parts) > 3 {
		return nil, fmt.Errorf("无效的切片表示法: %s", sliceStr)
	}

	// 解析开始
	if len(parts[0]) > 0 {
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("无效的切片开始索引: %s", parts[0])
		}
		result.Start = start
	}

	// 解析结束
	if len(parts) > 1 && len(parts[1]) > 0 {
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("无效的切片结束索引: %s", parts[1])
		}
		result.End = end
	}

	// 解析步长
	if len(parts) > 2 && len(parts[2]) > 0 {
		step, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("无效的切片步长: %s", parts[2])
		}
		if step == 0 {
			return nil, fmt.Errorf("切片步长不能为零")
		}
		result.Step = step
	}

	return result, nil
}

// Query 使用 JSON Path 查询 JSON 值并返回匹配的值列表
func (jp *JSONPath) Query(doc *Value) ([]*Value, error) {
	if doc == nil {
		return nil, fmt.Errorf("JSON 文档不能为空")
	}

	// 从根节点开始查询
	matches, err := jp.evaluate(doc, 0)
	if err != nil {
		return nil, err
	}
	return matches, nil
}

// evaluate 从指定令牌索引开始评估路径
func (jp *JSONPath) evaluate(current *Value, tokenIndex int) ([]*Value, error) {
	// 基本情况：已处理所有令牌
	if tokenIndex >= len(jp.Tokens) {
		return []*Value{current}, nil
	}

	// 获取当前令牌
	token := jp.Tokens[tokenIndex]

	switch token.Type {
	case ROOT:
		// $ 表示文档根节点，继续处理下一个令牌
		return jp.evaluate(current, tokenIndex+1)

	case DOT:
		// 确保下一个令牌是属性名或通配符
		if tokenIndex+1 >= len(jp.Tokens) {
			return nil, &JSONPathError{
				Path:    jp.Path,
				Message: ". 后面必须跟属性名或通配符",
			}
		}
		nextToken := jp.Tokens[tokenIndex+1]

		if nextToken.Type == PROPERTY {
			// 处理对象属性访问
			if current.Type != OBJECT {
				return []*Value{}, nil // 不是对象，返回空结果
			}

			var value *Value
			for i := 0; i < len(current.O); i++ {
				if current.O[i].K == nextToken.Value {
					value = current.O[i].V
					break
				}
			}

			if value == nil {
				return []*Value{}, nil // 未找到属性
			}

			return jp.evaluate(value, tokenIndex+2)

		} else if nextToken.Type == WILDCARD {
			// 处理通配符（返回所有属性）
			if current.Type != OBJECT {
				return []*Value{}, nil // 不是对象，返回空结果
			}

			var results []*Value
			for i := 0; i < len(current.O); i++ {
				subResults, err := jp.evaluate(current.O[i].V, tokenIndex+2)
				if err != nil {
					return nil, err
				}
				results = append(results, subResults...)
			}
			return results, nil
		}

		return nil, &JSONPathError{
			Path:    jp.Path,
			Message: fmt.Sprintf("不支持的 . 后的令牌类型: %v", nextToken.Type),
		}

	case RECURSIVE_DESCENT:
		// 递归下降 (..) 操作
		// 确保下一个令牌是属性名或通配符
		if tokenIndex+1 >= len(jp.Tokens) {
			return nil, &JSONPathError{
				Path:    jp.Path,
				Message: ".. 后面必须跟属性名或通配符",
			}
		}

		// 递归处理当前节点及其所有子节点
		return jp.findRecursive(current, tokenIndex+1)

	case BRACKET_START:
		// 方括号表达式 [...]
		// 确保有下一个令牌
		if tokenIndex+1 >= len(jp.Tokens) {
			return nil, &JSONPathError{
				Path:    jp.Path,
				Message: "[ 后必须有表达式",
			}
		}

		bracketToken := jp.Tokens[tokenIndex+1]

		// 确保后面跟着 ]
		if tokenIndex+2 >= len(jp.Tokens) || jp.Tokens[tokenIndex+2].Type != BRACKET_END {
			return nil, &JSONPathError{
				Path:    jp.Path,
				Message: "未闭合的 [",
			}
		}

		switch bracketToken.Type {
		case INDEX:
			// 数组索引 [0]
			if current.Type != ARRAY {
				return []*Value{}, nil // 不是数组，返回空结果
			}

			index, err := strconv.Atoi(bracketToken.Value)
			if err != nil {
				return nil, &JSONPathError{
					Path:    jp.Path,
					Message: fmt.Sprintf("无效的数组索引: %s", bracketToken.Value),
				}
			}

			// 处理负索引 (从末尾计数)
			if index < 0 {
				index = len(current.A) + index
			}

			// 索引范围检查
			if index < 0 || index >= len(current.A) {
				return []*Value{}, nil // 索引越界，返回空结果
			}

			return jp.evaluate(current.A[index], tokenIndex+3)

		case WILDCARD:
			// 通配符 [*]
			if current.Type != ARRAY && current.Type != OBJECT {
				return []*Value{}, nil // 既不是数组也不是对象，返回空结果
			}

			var results []*Value

			if current.Type == ARRAY {
				for i := 0; i < len(current.A); i++ {
					subResults, err := jp.evaluate(current.A[i], tokenIndex+3)
					if err != nil {
						return nil, err
					}
					results = append(results, subResults...)
				}
			} else { // OBJECT
				for i := 0; i < len(current.O); i++ {
					subResults, err := jp.evaluate(current.O[i].V, tokenIndex+3)
					if err != nil {
						return nil, err
					}
					results = append(results, subResults...)
				}
			}

			return results, nil

		case PROPERTY:
			// 属性访问 ["name"]
			if current.Type != OBJECT {
				return []*Value{}, nil // 不是对象，返回空结果
			}

			var value *Value
			for i := 0; i < len(current.O); i++ {
				if current.O[i].K == bracketToken.Value {
					value = current.O[i].V
					break
				}
			}

			if value == nil {
				return []*Value{}, nil // 未找到属性
			}

			return jp.evaluate(value, tokenIndex+3)

		case SLICE:
			// 切片 [start:end:step]
			if current.Type != ARRAY {
				return []*Value{}, nil // 不是数组，返回空结果
			}

			sliceInfo, err := parseSliceParams(bracketToken.Value)
			if err != nil {
				return nil, &JSONPathError{
					Path:    jp.Path,
					Message: err.Error(),
				}
			}

			var results []*Value
			arrayLen := len(current.A)

			// 调整负索引和默认结束索引
			start := sliceInfo.Start
			if start < 0 {
				start = arrayLen + start
			}
			if start < 0 {
				start = 0
			}

			end := sliceInfo.End
			if end < 0 {
				end = arrayLen
			}
			if end > arrayLen {
				end = arrayLen
			}

			// 应用切片
			step := sliceInfo.Step

			// 正向遍历
			if step > 0 {
				for i := start; i < end; i += step {
					if i >= 0 && i < arrayLen {
						subResults, err := jp.evaluate(current.A[i], tokenIndex+3)
						if err != nil {
							return nil, err
						}
						results = append(results, subResults...)
					}
				}
			} else {
				// 反向遍历
				for i := start; i > end; i += step {
					if i >= 0 && i < arrayLen {
						subResults, err := jp.evaluate(current.A[i], tokenIndex+3)
						if err != nil {
							return nil, err
						}
						results = append(results, subResults...)
					}
				}
			}

			return results, nil
		}

		return nil, &JSONPathError{
			Path:    jp.Path,
			Message: fmt.Sprintf("不支持的方括号表达式类型: %v", bracketToken.Type),
		}

	default:
		return nil, &JSONPathError{
			Path:    jp.Path,
			Message: fmt.Sprintf("不支持的令牌类型: %v", token.Type),
		}
	}
}

// findRecursive 递归查找匹配目标属性的所有节点
func (jp *JSONPath) findRecursive(current *Value, tokenIndex int) ([]*Value, error) {
	if current == nil {
		return []*Value{}, nil
	}

	// 创建结果集合
	var results []*Value

	// 尝试从当前节点匹配
	matches, err := jp.matchProperty(current, tokenIndex)
	if err == nil && len(matches) > 0 {
		results = append(results, matches...)
	}

	// 递归处理子节点
	if current.Type == OBJECT {
		for i := 0; i < len(current.O); i++ {
			childResults, err := jp.findRecursive(current.O[i].V, tokenIndex)
			if err == nil {
				results = append(results, childResults...)
			}
		}
	} else if current.Type == ARRAY {
		for i := 0; i < len(current.A); i++ {
			childResults, err := jp.findRecursive(current.A[i], tokenIndex)
			if err == nil {
				results = append(results, childResults...)
			}
		}
	}

	return results, nil
}

// matchProperty 尝试匹配当前节点的属性
func (jp *JSONPath) matchProperty(current *Value, tokenIndex int) ([]*Value, error) {
	if tokenIndex >= len(jp.Tokens) {
		return []*Value{}, nil
	}

	token := jp.Tokens[tokenIndex]

	if token.Type == PROPERTY {
		// 属性匹配
		if current.Type != OBJECT {
			return []*Value{}, nil
		}

		var value *Value
		for i := 0; i < len(current.O); i++ {
			if current.O[i].K == token.Value {
				value = current.O[i].V
				break
			}
		}

		if value == nil {
			return []*Value{}, nil
		}

		return jp.evaluate(value, tokenIndex+1)
	} else if token.Type == WILDCARD {
		// 通配符匹配
		if current.Type == OBJECT {
			var results []*Value
			for i := 0; i < len(current.O); i++ {
				childResults, err := jp.evaluate(current.O[i].V, tokenIndex+1)
				if err == nil {
					results = append(results, childResults...)
				}
			}
			return results, nil
		} else if current.Type == ARRAY {
			var results []*Value
			for i := 0; i < len(current.A); i++ {
				childResults, err := jp.evaluate(current.A[i], tokenIndex+1)
				if err == nil {
					results = append(results, childResults...)
				}
			}
			return results, nil
		}

		return []*Value{}, nil
	}

	return []*Value{}, nil
}

// QueryOne 返回第一个匹配的值，如果没有匹配则返回 nil
func (jp *JSONPath) QueryOne(doc *Value) (*Value, error) {
	results, err := jp.Query(doc)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
}

// QueryString 是一个便捷函数，直接使用路径表达式查询 JSON 值
func QueryString(doc *Value, path string) ([]*Value, error) {
	jp, err := NewJSONPath(path)
	if err != nil {
		return nil, err
	}
	return jp.Query(doc)
}

// QueryOneString 是一个便捷函数，返回匹配路径的第一个值
func QueryOneString(doc *Value, path string) (*Value, error) {
	jp, err := NewJSONPath(path)
	if err != nil {
		return nil, err
	}
	return jp.QueryOne(doc)
}
