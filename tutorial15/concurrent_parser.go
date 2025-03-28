// concurrent_parser.go - 实现并发JSON解析
package tutorial15

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/Cactusinhand/go-json-tutorial/tutorial15/leptjson"
)

// ParseError 定义解析错误类型
type ParseError int

// ChunkSize 定义默认的块大小
const ChunkSize = 64 * 1024 // 64KB

// 额外的解析错误常量
const (
	PARSE_TIMEOUT  ParseError = 100 // 解析超时
	PARSE_IO_ERROR ParseError = 101 // IO错误
)

// GetErrorString 获取错误描述信息
func GetErrorString(e int) string {
	switch e {
	case leptjson.PARSE_OK:
		return "解析成功"
	case leptjson.PARSE_EXPECT_VALUE:
		return "期望一个值"
	case leptjson.PARSE_INVALID_VALUE:
		return "无效的值"
	case leptjson.PARSE_ROOT_NOT_SINGULAR:
		return "根节点不唯一"
	case leptjson.PARSE_NUMBER_TOO_BIG:
		return "数字太大"
	case leptjson.PARSE_MISS_QUOTATION_MARK:
		return "缺少引号"
	case leptjson.PARSE_INVALID_STRING_ESCAPE:
		return "无效的转义序列"
	case leptjson.PARSE_INVALID_STRING_CHAR:
		return "无效的字符"
	case leptjson.PARSE_INVALID_UNICODE_HEX:
		return "无效的Unicode十六进制"
	case leptjson.PARSE_INVALID_UNICODE_SURROGATE:
		return "无效的Unicode代理对"
	case leptjson.PARSE_MISS_COMMA_OR_SQUARE_BRACKET:
		return "缺少逗号或方括号"
	case leptjson.PARSE_MISS_KEY:
		return "缺少键"
	case leptjson.PARSE_MISS_COLON:
		return "缺少冒号"
	case leptjson.PARSE_MISS_COMMA_OR_CURLY_BRACKET:
		return "缺少逗号或花括号"
	case int(PARSE_TIMEOUT):
		return "解析超时"
	case int(PARSE_IO_ERROR):
		return "IO错误"
	default:
		return fmt.Sprintf("未知错误(%d)", e)
	}
}

// ConcurrentParseOptions 定义并发解析的选项
type ConcurrentParseOptions struct {
	// 块大小（字节）
	ChunkSize int
	// 并发工作线程数量，默认为可用CPU核心数
	WorkerCount int
	// 是否自动检测最佳工作线程数量
	AutoWorkers bool
	// 超时时间
	Timeout time.Duration
}

// DefaultConcurrentParseOptions 返回默认的并发解析选项
func DefaultConcurrentParseOptions() ConcurrentParseOptions {
	return ConcurrentParseOptions{
		ChunkSize:   ChunkSize,
		WorkerCount: 4,                 // 默认4个工作线程
		AutoWorkers: true,              // 自动调整为最优
		Timeout:     120 * time.Second, // 默认120秒超时，原来是30秒
	}
}

// ParseResult 表示解析结果
type ParseResult struct {
	Value *Value
	Error ParseError
}

// 解析任务
type parseTask struct {
	id       int    // 任务ID，用于确保顺序
	chunk    []byte // 块数据
	startPos int    // 起始位置
	isFirst  bool   // 是否是第一个块
	isLast   bool   // 是否是最后一个块
	prefix   []byte // 前缀缓冲区
	suffix   []byte // 后缀缓冲区
}

// 解析结果任务
type parseResultTask struct {
	id     int         // 任务ID，用于排序
	result ParseResult // 解析结果
}

// StreamingParser 表示流式解析器
type StreamingParser struct {
	options      ConcurrentParseOptions // 解析选项
	tasks        chan parseTask         // 任务通道
	results      chan parseResultTask   // 结果通道
	errorChan    chan error             // 错误通道
	wg           sync.WaitGroup         // 等待组
	ctx          context.Context        // 上下文
	cancel       context.CancelFunc     // 取消函数
	partialData  map[int][]byte         // 存储部分数据
	partialMutex sync.Mutex             // 部分数据的互斥锁
}

// NewStreamingParser 创建一个新的流式解析器
func NewStreamingParser(options ConcurrentParseOptions) *StreamingParser {
	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)

	return &StreamingParser{
		options:     options,
		tasks:       make(chan parseTask, options.WorkerCount*2),
		results:     make(chan parseResultTask, options.WorkerCount*2),
		errorChan:   make(chan error, options.WorkerCount),
		ctx:         ctx,
		cancel:      cancel,
		partialData: make(map[int][]byte),
	}
}

// ParseStream 从Reader中解析JSON
func (p *StreamingParser) ParseStream(reader io.Reader) (*Value, ParseError) {
	defer p.cancel() // 确保资源被释放

	// 启动工作线程
	for i := 0; i < p.options.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	// 启动结果收集
	resultCh := make(chan ParseResult, 1)
	go p.collectResults(resultCh)

	// 分块读取并分发任务
	buffer := make([]byte, p.options.ChunkSize)
	var totalRead int
	var taskID int
	isFirst := true

	for {
		select {
		case <-p.ctx.Done():
			return nil, PARSE_TIMEOUT
		default:
			n, err := reader.Read(buffer)
			if n > 0 {
				// 创建任务块的副本
				chunk := make([]byte, n)
				copy(chunk, buffer[:n])

				isLast := err == io.EOF

				// 分析当前块是否包含完整JSON或需要与下一块合并
				var prefix, suffix []byte

				// 查找块的边界位置
				if !isFirst && !isLast {
					prefix, suffix = findJsonBoundaries(chunk)
				}

				p.tasks <- parseTask{
					id:       taskID,
					chunk:    chunk,
					startPos: totalRead,
					isFirst:  isFirst,
					isLast:   isLast,
					prefix:   prefix,
					suffix:   suffix,
				}

				taskID++
				totalRead += n
				isFirst = false
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				// 在发生错误时，取消所有操作
				p.cancel()
				close(p.tasks)
				return nil, PARSE_IO_ERROR
			}
		}
	}

	// 关闭任务通道，表示没有更多数据
	close(p.tasks)

	// 等待所有工作线程完成或超时
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(p.results) // 所有工作线程完成后关闭结果通道
		close(done)
	}()

	select {
	case <-done:
		// 工作线程正常完成
	case <-p.ctx.Done():
		// 超时
		return nil, PARSE_TIMEOUT
	}

	// 获取最终结果
	select {
	case result := <-resultCh:
		return result.Value, result.Error
	case <-p.ctx.Done():
		return nil, PARSE_TIMEOUT
	}
}

// 查找JSON数据的自然边界
func findJsonBoundaries(data []byte) ([]byte, []byte) {
	// 简化版本：尝试找到对象或数组的完整部分和不完整部分
	// 在实际实现中，这需要考虑字符串中的大括号和引号等复杂情况

	// 查找最后一个未闭合的大括号或方括号
	stack := make([]byte, 0, 32)
	inString := false
	var lastIncompletePos int = -1

	for i := 0; i < len(data); i++ {
		char := data[i]

		if inString {
			if char == '"' && (i == 0 || data[i-1] != '\\') {
				inString = false
			}
			continue
		}

		if char == '"' {
			inString = true
			continue
		}

		if char == '{' || char == '[' {
			stack = append(stack, char)
			continue
		}

		if (char == '}' && len(stack) > 0 && stack[len(stack)-1] == '{') ||
			(char == ']' && len(stack) > 0 && stack[len(stack)-1] == '[') {
			stack = stack[:len(stack)-1] // 弹出匹配的括号

			// 如果栈空了，说明到这个位置的JSON是完整的
			if len(stack) == 0 {
				lastIncompletePos = -1
			}
			continue
		}

		// 如果栈不为空，更新最后一个不完整位置
		if len(stack) > 0 {
			lastIncompletePos = i
		}
	}

	// 如果找到不完整位置
	if lastIncompletePos >= 0 {
		return data[:lastIncompletePos], data[lastIncompletePos:]
	}

	// 如果没有找到不完整位置但栈不为空，说明整个块都是不完整的
	if len(stack) > 0 {
		return nil, data
	}

	// 整个块都是完整的
	return data, nil
}

// worker 解析任务线程
func (p *StreamingParser) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// 上下文取消时退出
			return
		case task, ok := <-p.tasks:
			if !ok {
				// 任务通道关闭，没有更多任务
				return
			}

			// 获取并处理任务
			result := parseResultTask{
				id: task.id,
				result: ParseResult{
					Value: &Value{},
					Error: ParseError(leptjson.PARSE_OK),
				},
			}

			// 简单情况：完整的单个JSON文档
			if task.isFirst && task.isLast {
				err := Parse(result.result.Value, string(task.chunk))
				if err != nil {
					result.result.Error = ToParseError(err)
				}

				select {
				case p.results <- result:
					// 结果已发送
				case <-p.ctx.Done():
					// 上下文取消
					return
				}
				continue
			}

			// 复杂情况：需要合并多个块
			p.partialMutex.Lock()

			// 第一个块的情况
			if task.isFirst {
				p.partialData[task.id] = task.chunk
				p.partialMutex.Unlock()
				continue
			}

			// 最后一个块的情况
			if task.isLast {
				// 尝试和前一个块合并
				if data, exists := p.partialData[task.id-1]; exists {
					combined := append(data, task.chunk...)
					delete(p.partialData, task.id-1)
					p.partialMutex.Unlock()

					err := Parse(result.result.Value, string(combined))
					if err != nil {
						result.result.Error = ToParseError(err)
					}

					select {
					case p.results <- result:
						// 结果已发送
					case <-p.ctx.Done():
						// 上下文取消
						return
					}
				} else {
					// 单独处理最后块
					p.partialMutex.Unlock()
					err := Parse(result.result.Value, string(task.chunk))
					if err != nil {
						result.result.Error = ToParseError(err)
					}

					select {
					case p.results <- result:
						// 结果已发送
					case <-p.ctx.Done():
						// 上下文取消
						return
					}
				}
				continue
			}

			// 中间块的情况
			if task.prefix != nil && task.suffix != nil {
				// 处理完整部分
				if len(task.prefix) > 0 {
					tempValue := &Value{}
					err := Parse(tempValue, string(task.prefix))
					if err == nil {
						// 发送完整部分
						partResult := parseResultTask{
							id: task.id * 1000, // 使用不同ID范围区分部分结果
							result: ParseResult{
								Value: tempValue,
								Error: ParseError(leptjson.PARSE_OK),
							},
						}
						select {
						case p.results <- partResult:
							// 结果已发送
						case <-p.ctx.Done():
							p.partialMutex.Unlock()
							return
						}
					}
				}

				// 保存不完整部分
				if len(task.suffix) > 0 {
					p.partialData[task.id] = task.suffix
				}
				p.partialMutex.Unlock()
				continue
			}

			// 默认情况：将整个块存储起来等待后续处理
			p.partialData[task.id] = task.chunk
			p.partialMutex.Unlock()
		}
	}
}

// collectResults 收集所有并发解析结果
func (p *StreamingParser) collectResults(resultCh chan<- ParseResult) {
	defer close(resultCh)

	var orderedResults []parseResultTask
	idTracker := make(map[int]bool)
	var nextExpectedID int

	for {
		select {
		case <-p.ctx.Done():
			// 上下文取消，结束收集
			resultCh <- ParseResult{nil, PARSE_TIMEOUT}
			return

		case result, ok := <-p.results:
			if !ok {
				// 结果通道已关闭，处理最后的结果
				finalResult := mergeResults(orderedResults)
				resultCh <- finalResult
				return
			}

			// 记录结果
			idTracker[result.id] = true

			// 插入到有序结果列表
			inserted := false
			for i, r := range orderedResults {
				if result.id < r.id {
					// 插入到合适的位置
					orderedResults = append(orderedResults[:i], append([]parseResultTask{result}, orderedResults[i:]...)...)
					inserted = true
					break
				}
			}

			if !inserted {
				orderedResults = append(orderedResults, result)
			}

			// 检查是否可以处理连续的结果
			for {
				if idTracker[nextExpectedID] {
					nextExpectedID++
				} else {
					break
				}
			}

			// 如果有足够多的结果，并且都是连续的，可以提前返回
			if nextExpectedID > 0 && len(orderedResults) > 0 && orderedResults[0].id == 0 {
				// 尝试查找连续的最后一个ID
				lastConsecutiveIdx := -1
				for i, r := range orderedResults {
					if r.id == i {
						lastConsecutiveIdx = i
					} else {
						break
					}
				}

				// 如果有连续的结果，并且最后一个结果有错误，立即返回错误
				if lastConsecutiveIdx >= 0 {
					for i := 0; i <= lastConsecutiveIdx; i++ {
						if orderedResults[i].result.Error != ParseError(leptjson.PARSE_OK) {
							resultCh <- orderedResults[i].result
							return
						}
					}

					// 合并连续的结果
					if lastConsecutiveIdx > 0 {
						consecutiveResults := orderedResults[:lastConsecutiveIdx+1]
						mergedResult := mergeResults(consecutiveResults)

						// 如果所有结果都处理完毕，或者有明确的错误
						if lastConsecutiveIdx == len(orderedResults)-1 || mergedResult.Error != ParseError(leptjson.PARSE_OK) {
							resultCh <- mergedResult
							return
						}
					}
				}
			}
		}
	}
}

// mergeResults 合并多个解析结果
func mergeResults(results []parseResultTask) ParseResult {
	if len(results) == 0 {
		return ParseResult{
			Value: nil,
			Error: ParseError(leptjson.PARSE_EXPECT_VALUE),
		}
	}

	// 找到第一个有效的结果
	var firstValid *Value
	for _, r := range results {
		if r.result.Error == ParseError(leptjson.PARSE_OK) && r.result.Value != nil {
			firstValid = r.result.Value
			break
		}
	}

	if firstValid == nil {
		// 如果没有有效结果，返回第一个结果
		return results[0].result
	}

	// 尝试合并结果
	result := ParseResult{
		Value: &Value{},
		Error: ParseError(leptjson.PARSE_OK),
	}

	// 深度复制第一个有效结果
	*result.Value = *firstValid

	// 根据值类型决定如何合并
	if result.Value.Type == ARRAY {
		for i := 1; i < len(results); i++ {
			if results[i].result.Error != ParseError(leptjson.PARSE_OK) || results[i].result.Value == nil {
				continue
			}

			// 如果是数组，尝试附加元素
			if results[i].result.Value.Type == ARRAY {
				// 附加数组元素
				for _, elem := range results[i].result.Value.Elements {
					// 深度复制元素
					newElem := &Value{}
					*newElem = *elem
					PushBackArrayElement(result.Value, newElem)
				}
			}
		}
	} else if result.Value.Type == OBJECT {
		// 如果是对象，尝试合并其他结果中的对象成员
		for i := 1; i < len(results); i++ {
			if results[i].result.Error != ParseError(leptjson.PARSE_OK) || results[i].result.Value == nil {
				continue
			}

			// 如果是对象，尝试合并成员
			if results[i].result.Value.Type == OBJECT {
				// 合并对象成员
				for _, member := range results[i].result.Value.Members {
					// 检查是否已存在该键
					existing := FindObjectValue(result.Value, member.Key)
					if existing == nil {
						// 不存在，添加新成员
						newValue := SetObjectValue(result.Value, member.Key)
						*newValue = *member.Value
					}
				}
			}
		}
	}

	return result
}

// ParseConcurrent 并发解析JSON文本
func ParseConcurrent(json string, options ConcurrentParseOptions) (*Value, error) {
	// 创建字符串读取器
	reader := strings.NewReader(json)

	// 使用流式解析器
	parser := NewStreamingParser(options)
	value, parseErr := parser.ParseStream(reader)

	if parseErr != 0 {
		return nil, fmt.Errorf("解析错误: %s", GetErrorString(int(parseErr)))
	}

	return value, nil
}

// 实现更复杂的并发流式解析
// 以下是一个更实际的流式解析实现，处理JSON数组

// ArrayStreamParser 数组流式解析器
type ArrayStreamParser struct {
	tokenizer *JSONTokenizer
	mutex     sync.Mutex
	elements  []*Value
	index     int
}

// JSONTokenizer 简单的JSON词法分析器
type JSONTokenizer struct {
	data   []byte
	pos    int
	length int
	mutex  sync.Mutex
}

// NewJSONTokenizer 创建新的词法分析器
func NewJSONTokenizer(data []byte) *JSONTokenizer {
	return &JSONTokenizer{
		data:   data,
		pos:    0,
		length: len(data),
	}
}

// NextToken 获取下一个Token
func (t *JSONTokenizer) NextToken() (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 跳过空白字符
	for t.pos < t.length && isWhitespace(t.data[t.pos]) {
		t.pos++
	}

	if t.pos >= t.length {
		return "", io.EOF
	}

	switch t.data[t.pos] {
	case '{', '}', '[', ']', ',', ':', '"':
		token := string(t.data[t.pos])
		t.pos++
		return token, nil
	default:
		// 处理数字和关键字
		start := t.pos
		for t.pos < t.length && !isDelimiter(t.data[t.pos]) {
			t.pos++
		}
		return string(t.data[start:t.pos]), nil
	}
}

// 辅助函数：判断是否为空白字符
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// 辅助函数：判断是否为分隔符
func isDelimiter(c byte) bool {
	return c == '{' || c == '}' || c == '[' || c == ']' || c == ',' || c == ':' || c == '"' || isWhitespace(c)
}

// NewArrayStreamParser 创建新的数组流式解析器
func NewArrayStreamParser(data []byte) (*ArrayStreamParser, error) {
	tokenizer := NewJSONTokenizer(data)

	// 确保数据以'['开始
	token, err := tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	if token != "[" {
		return nil, fmt.Errorf("预期'['，实际得到 '%s'", token)
	}

	return &ArrayStreamParser{
		tokenizer: tokenizer,
		elements:  make([]*Value, 0),
		index:     0,
	}, nil
}

// NextElement 获取数组的下一个元素
func (p *ArrayStreamParser) NextElement() (*Value, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 如果已经有解析好的元素，直接返回
	if p.index < len(p.elements) {
		element := p.elements[p.index]
		p.index++
		return element, nil
	}

	// 否则，解析下一个元素
	token, err := p.tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	// 检查数组结束
	if token == "]" {
		return nil, io.EOF
	}

	// 处理逗号
	if token == "," {
		token, err = p.tokenizer.NextToken()
		if err != nil {
			return nil, err
		}
	}

	// 根据token类型创建Value对象
	element := &Value{}

	if token == "null" {
		SetNull(element)
	} else if token == "true" {
		SetBool(element, true)
	} else if token == "false" {
		SetBool(element, false)
	} else if token == "\"" {
		// 处理字符串
		var sb bytes.Buffer
		for {
			char, err := p.tokenizer.NextToken()
			if err != nil {
				return nil, err
			}

			if char == "\"" {
				break
			}

			sb.WriteString(char)
		}
		SetString(element, sb.String())
	} else if token == "{" {
		// 处理对象
		// 这需要更复杂的实现，这里简化处理
		SetObject(element)
	} else if token == "[" {
		// 处理嵌套数组
		// 这需要更复杂的实现，这里简化处理
		SetArray(element)
	} else {
		// 假设是数字
		SetNumber(element, 0) // 实际应该解析数字
	}

	// 保存并返回元素
	p.elements = append(p.elements, element)
	p.index++
	return element, nil
}

// 并发处理数组所有元素
func ProcessArrayConcurrently(data []byte, workerCount int, processor func(*Value) error) error {
	parser, err := NewArrayStreamParser(data)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	taskChan := make(chan *Value)
	resultChan := make(chan int, 1000) // 用于跟踪处理完成的元素索引
	errorChan := make(chan error, workerCount)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // 增加超时时间到120秒
	defer cancel()

	// 启动工作线程
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case element, ok := <-taskChan:
					if !ok {
						return
					}
					if err := processor(element); err != nil {
						select {
						case errorChan <- err:
							// 错误已发送
						default:
							// 错误通道满了，继续处理
						}
					} else {
						// 处理成功，发送索引到结果通道
						if element.Type == NUMBER {
							resultChan <- int(element.N)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// 生成任务
	go func() {
		defer close(taskChan)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				element, err := parser.NextElement()
				if err == io.EOF {
					return
				}
				if err != nil {
					select {
					case errorChan <- err:
						// 错误已发送
					default:
						// 错误通道满了，继续处理
					}
					return
				}

				select {
				case taskChan <- element:
					// 任务已发送
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// 等待所有工作线程完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(resultChan) // 所有工作线程完成后，关闭结果通道
	}()

	select {
	case <-done:
		// 检查是否有错误
		select {
		case err := <-errorChan:
			return err
		default:
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Error 实现error接口，返回错误描述
func (e ParseError) Error() string {
	return GetErrorString(int(e))
}

// ToParseError 将错误转换为ParseError类型
func ToParseError(err error) ParseError {
	if err == nil {
		return ParseError(leptjson.PARSE_OK)
	}
	// 简单处理，实际可以根据错误类型做更详细的映射
	return ParseError(leptjson.PARSE_INVALID_VALUE)
}
