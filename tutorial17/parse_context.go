package leptjson

import (
	"fmt"
)

// parseContext 解析时的上下文结构
type parseContext struct {
	json     string       // JSON 文本
	index    int          // 当前索引位置
	options  ParseOptions // 解析选项
	line     int          // 当前行号
	column   int          // 当前列号
	linePos  []int        // 每行开始的索引位置
	depth    int          // 当前解析深度
	recovery bool         // 是否正在进行错误恢复

	// 安全统计
	stringLengths int // 已解析的字符串长度总和
	arrayElements int // 已解析的数组元素总数
	objectMembers int // 已解析的对象成员总数

	// 当前处理的数组/对象统计
	currentArraySize  int // 当前数组的元素数量
	currentObjectSize int // 当前对象的成员数量
}

// 初始化解析上下文
func newContext(json string, options ParseOptions) *parseContext {
	c := &parseContext{
		json:              json,
		index:             0,
		options:           options,
		line:              1,
		column:            1,
		linePos:           []int{}, // 首行的位置是0，我们不存储
		depth:             0,
		stringLengths:     0,
		arrayElements:     0,
		objectMembers:     0,
		currentArraySize:  0,
		currentObjectSize: 0,
	}

	// 预处理计算所有行的起始位置
	for i := 0; i < len(json); i++ {
		if json[i] == '\n' {
			c.linePos = append(c.linePos, i+1)
		}
	}

	return c
}

// nextChar 读取下一个字符并更新位置信息
func (c *parseContext) nextChar() byte {
	if c.index >= len(c.json) {
		return 0
	}

	ch := c.json[c.index]
	c.index++

	// 更新行列信息
	if ch == '\n' {
		c.line++
		c.column = 1
	} else {
		c.column++
	}

	return ch
}

// peekChar 查看当前字符但不移动位置
func (c *parseContext) peekChar() byte {
	if c.index >= len(c.json) {
		return 0
	}
	return c.json[c.index]
}

// createError 创建增强的错误信息
func (c *parseContext) createError(code ParseError, message string) *EnhancedError {
	isRecoverable := false

	// 确定哪些错误类型可恢复
	switch code {
	case PARSE_MISS_COMMA_OR_SQUARE_BRACKET,
		PARSE_MISS_COMMA_OR_CURLY_BRACKET,
		PARSE_INVALID_VALUE:
		isRecoverable = true
	}

	// 创建增强的错误信息
	var finalMessage string
	if message == "" {
		finalMessage = GetErrorMessage(code)
	} else {
		finalMessage = message
	}

	return createEnhancedError(
		code,
		c.json,
		c.index,
		c.linePos,
		finalMessage,
		isRecoverable && c.options.RecoverFromErrors,
	)
}

// 解析空白字符
func (c *parseContext) parseWhitespace() {
	for c.index < len(c.json) {
		if c.json[c.index] == ' ' || c.json[c.index] == '\t' ||
			c.json[c.index] == '\n' || c.json[c.index] == '\r' {
			c.nextChar()
		} else if c.options.AllowComments && c.json[c.index] == '/' {
			// 如果允许注释，尝试解析
			savedIndex := c.index
			savedLine := c.line
			savedColumn := c.column

			if err := c.parseComment(); err != PARSE_OK {
				// 恢复位置，这不是有效的注释
				c.index = savedIndex
				c.line = savedLine
				c.column = savedColumn
				break
			}
		} else {
			break
		}
	}
}

// parseComment 解析注释，支持 // 和 /* */
func (c *parseContext) parseComment() ParseError {
	if c.index+1 >= len(c.json) {
		return PARSE_INVALID_VALUE
	}

	next := c.json[c.index+1]
	if next == '/' { // 单行注释 //
		c.index += 2
		c.column += 2

		// 一直读到行尾
		for c.index < len(c.json) && c.json[c.index] != '\n' {
			c.index++
			c.column++
		}
		return PARSE_OK
	} else if next == '*' { // 多行注释 /* */
		c.index += 2
		c.column += 2

		for c.index+1 < len(c.json) {
			if c.json[c.index] == '*' && c.json[c.index+1] == '/' {
				c.index += 2
				c.column += 2
				return PARSE_OK
			}

			if c.json[c.index] == '\n' {
				c.line++
				c.column = 1
				c.linePos = append(c.linePos, c.index+1)
			} else {
				c.column++
			}
			c.index++
		}

		// 如果到了文件尾部还没找到注释结束
		return PARSE_COMMENT_NOT_CLOSED
	}

	return PARSE_INVALID_VALUE
}

// 检查嵌套深度并增加深度计数
func (c *parseContext) enterNesting() (bool, *EnhancedError) {
	c.depth++
	// 仅当安全检查开启时才检查最大嵌套深度
	if c.options.EnabledSecurity && c.depth > c.options.MaxDepth {
		return false, c.createError(PARSE_MAX_DEPTH_EXCEEDED,
			fmt.Sprintf("嵌套深度(%d)超过允许的最大值(%d)", c.depth, c.options.MaxDepth))
	}
	return true, nil
}

// 减少嵌套深度计数
func (c *parseContext) exitNesting() {
	c.depth--
}

// 新增安全检查方法

// 检查输入总大小
func (c *parseContext) checkTotalSize() (bool, *EnhancedError) {
	if !c.options.EnabledSecurity {
		return true, nil
	}

	if len(c.json) > c.options.MaxTotalSize {
		return false, c.createError(PARSE_MAX_TOTAL_SIZE_EXCEEDED,
			fmt.Sprintf("输入大小(%d)超过允许的最大值(%d)", len(c.json), c.options.MaxTotalSize))
	}
	return true, nil
}

// 检查字符串长度
func (c *parseContext) checkStringLength(length int) (bool, *EnhancedError) {
	if !c.options.EnabledSecurity {
		return true, nil
	}

	if length > c.options.MaxStringLength {
		return false, c.createError(PARSE_MAX_STRING_LENGTH_EXCEEDED,
			fmt.Sprintf("字符串长度(%d)超过允许的最大值(%d)", length, c.options.MaxStringLength))
	}

	c.stringLengths += length
	return true, nil
}

// 开始解析数组，检测并初始化数组大小统计
func (c *parseContext) enterArray() {
	c.currentArraySize = 0
}

// 添加数组元素时检查大小
func (c *parseContext) addArrayElement() (bool, *EnhancedError) {
	if !c.options.EnabledSecurity {
		return true, nil
	}

	c.currentArraySize++
	c.arrayElements++

	if c.currentArraySize > c.options.MaxArraySize {
		return false, c.createError(PARSE_MAX_ARRAY_SIZE_EXCEEDED,
			fmt.Sprintf("数组元素数量(%d)超过允许的最大值(%d)", c.currentArraySize, c.options.MaxArraySize))
	}
	return true, nil
}

// 退出数组解析
func (c *parseContext) exitArray() {
	// 重置当前数组大小统计
	c.currentArraySize = 0
}

// 开始解析对象，检测并初始化对象大小统计
func (c *parseContext) enterObject() {
	c.currentObjectSize = 0
}

// 添加对象成员时检查大小
func (c *parseContext) addObjectMember() (bool, *EnhancedError) {
	if !c.options.EnabledSecurity {
		return true, nil
	}

	c.currentObjectSize++
	c.objectMembers++

	if c.currentObjectSize > c.options.MaxObjectSize {
		return false, c.createError(PARSE_MAX_OBJECT_SIZE_EXCEEDED,
			fmt.Sprintf("对象成员数量(%d)超过允许的最大值(%d)", c.currentObjectSize, c.options.MaxObjectSize))
	}
	return true, nil
}

// 退出对象解析
func (c *parseContext) exitObject() {
	// 重置当前对象大小统计
	c.currentObjectSize = 0
}

// 检查数字值范围
func (c *parseContext) checkNumberRange(num float64) (bool, *EnhancedError) {
	if !c.options.EnabledSecurity {
		return true, nil
	}

	if num > c.options.MaxNumberValue || num < c.options.MinNumberValue {
		return false, c.createError(PARSE_NUMBER_RANGE_EXCEEDED,
			fmt.Sprintf("数值(%g)超出允许范围[%g, %g]", num, c.options.MinNumberValue, c.options.MaxNumberValue))
	}
	return true, nil
}
