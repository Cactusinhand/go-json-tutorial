import React, { useState, useEffect, useRef } from 'react';
import './JsonVisualizer.css';

// 模拟的解析步骤
const generateParseSteps = (jsonInput) => {
  const steps = [];
  let position = 0;
  
  // 简化版的解析步骤生成，实际实现会更复杂
  while (position < jsonInput.length) {
    // 跳过空白字符
    while (position < jsonInput.length && 
           [' ', '\t', '\n', '\r'].includes(jsonInput[position])) {
      position++;
    }
    
    if (position >= jsonInput.length) break;
    
    const char = jsonInput[position];
    let state = '';
    let codeSnippet = '';
    
    // 根据当前字符确定解析状态
    if (char === '{') {
      state = '开始解析对象';
      codeSnippet = 'return p.parseObject(v)';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (char === '}') {
      state = '结束解析对象';
      codeSnippet = 'v.Type = TypeObject';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (char === '[') {
      state = '开始解析数组';
      codeSnippet = 'return p.parseArray(v)';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (char === ']') {
      state = '结束解析数组';
      codeSnippet = 'v.Type = TypeArray';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (char === '"') {
      // 寻找字符串结束位置
      let endPos = position + 1;
      while (endPos < jsonInput.length) {
        if (jsonInput[endPos] === '\\') {
          endPos += 2;
          continue;
        }
        if (jsonInput[endPos] === '"') {
          break;
        }
        endPos++;
      }
      
      if (endPos >= jsonInput.length) {
        state = '错误：未闭合的字符串';
        codeSnippet = 'return ParseMissQuotationMark';
      } else {
        state = '解析字符串';
        codeSnippet = 'return p.parseString(v)';
      }
      
      steps.push({
        position,
        currentChar: jsonInput.substring(position, endPos + 1),
        state,
        codeSnippet,
        highlightLength: endPos - position + 1
      });
      position = endPos + 1;
    } else if (char === ':') {
      state = '解析冒号';
      codeSnippet = 'if p.next() != \':\' { return ParseMissColon }';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (char === ',') {
      state = '解析逗号';
      codeSnippet = 'if p.next() != \',\' { return ParseMissComma }';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    } else if (['-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'].includes(char)) {
      // 寻找数字结束位置
      let endPos = position;
      while (endPos < jsonInput.length && 
             ['-', '+', '.', 'e', 'E', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'].includes(jsonInput[endPos])) {
        endPos++;
      }
      
      state = '解析数字';
      codeSnippet = 'return p.parseNumber(v)';
      steps.push({
        position,
        currentChar: jsonInput.substring(position, endPos),
        state,
        codeSnippet,
        highlightLength: endPos - position
      });
      position = endPos;
    } else if (jsonInput.substring(position, position + 4) === 'true') {
      state = '解析布尔值 true';
      codeSnippet = 'return p.parseTrue(v)';
      steps.push({
        position,
        currentChar: 'true',
        state,
        codeSnippet,
        highlightLength: 4
      });
      position += 4;
    } else if (jsonInput.substring(position, position + 5) === 'false') {
      state = '解析布尔值 false';
      codeSnippet = 'return p.parseFalse(v)';
      steps.push({
        position,
        currentChar: 'false',
        state,
        codeSnippet,
        highlightLength: 5
      });
      position += 5;
    } else if (jsonInput.substring(position, position + 4) === 'null') {
      state = '解析 null';
      codeSnippet = 'return p.parseNull(v)';
      steps.push({
        position,
        currentChar: 'null',
        state,
        codeSnippet,
        highlightLength: 4
      });
      position += 4;
    } else {
      // 无效字符
      state = '错误：无效的 JSON 字符';
      codeSnippet = 'return ParseInvalidValue';
      steps.push({
        position,
        currentChar: char,
        state,
        codeSnippet,
        highlightLength: 1
      });
      position++;
    }
  }
  
  return steps;
};

const JsonVisualizer = () => {
  const [jsonInput, setJsonInput] = useState('{ "name": "John", "age": 30, "isStudent": false, "courses": ["Math", "Computer Science"] }');
  const [parseSteps, setParseSteps] = useState([]);
  const [currentStepIndex, setCurrentStepIndex] = useState(-1);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(1000);
  const timerRef = useRef(null);
  
  useEffect(() => {
    // 生成解析步骤
    setParseSteps(generateParseSteps(jsonInput));
    setCurrentStepIndex(-1);
  }, [jsonInput]);
  
  useEffect(() => {
    // 清理定时器
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);
  
  const start = () => {
    setIsPlaying(true);
  };
  
  const pause = () => {
    setIsPlaying(false);
    if (timerRef.current) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  };
  
  const next = () => {
    if (currentStepIndex < parseSteps.length - 1) {
      setCurrentStepIndex(prev => prev + 1);
    } else {
      pause();
    }
  };
  
  const previous = () => {
    if (currentStepIndex > 0) {
      setCurrentStepIndex(prev => prev - 1);
    }
  };
  
  const reset = () => {
    pause();
    setCurrentStepIndex(-1);
  };
  
  useEffect(() => {
    if (isPlaying) {
      timerRef.current = setTimeout(() => {
        next();
      }, speed);
    }
  }, [isPlaying, currentStepIndex]);
  
  const handleSpeedChange = (e) => {
    setSpeed(Number(e.target.value));
  };
  
  const renderHighlightedJson = () => {
    if (currentStepIndex === -1 || !parseSteps[currentStepIndex]) {
      return <pre>{jsonInput}</pre>;
    }
    
    const step = parseSteps[currentStepIndex];
    const beforeHighlight = jsonInput.substring(0, step.position);
    const highlight = jsonInput.substring(step.position, step.position + step.highlightLength);
    const afterHighlight = jsonInput.substring(step.position + step.highlightLength);
    
    return (
      <pre>
        {beforeHighlight}
        <span className="highlight">{highlight}</span>
        {afterHighlight}
      </pre>
    );
  };
  
  return (
    <div className="json-visualizer">
      <h2>JSON 解析可视化器</h2>
      
      <div className="input-section">
        <h3>JSON 输入</h3>
        <textarea
          value={jsonInput}
          onChange={(e) => setJsonInput(e.target.value)}
          rows={5}
          className="json-input"
        />
        <button onClick={() => setParseSteps(generateParseSteps(jsonInput))}>
          重新生成步骤
        </button>
      </div>
      
      <div className="visualization-section">
        <h3>可视化</h3>
        <div className="json-display">
          {renderHighlightedJson()}
        </div>
        
        <div className="parse-info">
          {currentStepIndex >= 0 && parseSteps[currentStepIndex] && (
            <>
              <h4>当前步骤: {currentStepIndex + 1} / {parseSteps.length}</h4>
              <p><strong>状态:</strong> {parseSteps[currentStepIndex].state}</p>
              <div className="code-snippet">
                <h4>代码:</h4>
                <pre>{parseSteps[currentStepIndex].codeSnippet}</pre>
              </div>
            </>
          )}
        </div>
      </div>
      
      <div className="controls">
        <button onClick={reset} disabled={currentStepIndex === -1}>
          重置
        </button>
        <button onClick={previous} disabled={currentStepIndex <= 0}>
          上一步
        </button>
        {isPlaying ? (
          <button onClick={pause}>暂停</button>
        ) : (
          <button onClick={start} disabled={currentStepIndex >= parseSteps.length - 1}>
            播放
          </button>
        )}
        <button onClick={next} disabled={currentStepIndex >= parseSteps.length - 1}>
          下一步
        </button>
        
        <div className="speed-control">
          <label>
            速度:
            <input
              type="range"
              min="100"
              max="2000"
              step="100"
              value={speed}
              onChange={handleSpeedChange}
            />
            {speed} ms
          </label>
        </div>
      </div>
    </div>
  );
};

export default JsonVisualizer; 