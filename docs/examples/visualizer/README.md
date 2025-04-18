# JSON 可视化解析器

这个示例展示了如何实现一个交互式的 JSON 解析可视化工具，可以帮助学习者理解 JSON 解析的过程。

## 功能特点

- 逐步展示 JSON 解析过程
- 可视化解析状态和内部数据结构
- 动态高亮当前处理的 JSON 片段
- 相应的源代码同步高亮显示
- 内存状态变化的跟踪

## 组件架构

### 1. 输入区域

用户可以输入或选择预定义的 JSON 字符串进行解析。

### 2. 可视化面板

展示解析过程的主要区域，包括：
- 字符扫描动画
- 语法树构建
- 状态机转换
- 错误处理

### 3. 代码展示区

显示与当前解析步骤对应的源代码，并高亮当前执行的代码行。

### 4. 控制面板

提供控制解析过程的功能：
- 开始/暂停
- 下一步
- 上一步
- 重置
- 调整速度

## 技术实现

前端实现将使用以下技术：
- React 用于构建 UI 组件
- D3.js 用于可视化效果
- Monaco Editor 用于代码展示
- CSS 动画用于平滑过渡效果

关键实现点：
1. 将解析过程分解为离散步骤
2. 为每个步骤记录状态变化
3. 实现状态之间的平滑过渡动画
4. 同步代码和可视化状态

## 示例代码

### 解析步骤定义

```typescript
interface ParseStep {
  position: number;        // 当前解析位置
  currentChar: string;     // 当前字符
  state: string;           // 解析器状态
  valueStack: any[];       // 值栈
  currentValue: any;       // 当前构建的值
  error: string | null;    // 错误信息
  codeHighlight: {         // 高亮的代码行
    start: number;
    end: number;
  };
}
```

### 动画控制器

```typescript
class VisualizerController {
  private steps: ParseStep[] = [];
  private currentStepIndex: number = -1;
  private speed: number = 1000; // ms per step
  private isPlaying: boolean = false;
  private timer: any = null;
  
  constructor(private jsonInput: string, private onStepChange: (step: ParseStep) => void) {
    this.generateSteps();
  }
  
  private generateSteps() {
    // 解析 JSON 并记录每个步骤
    // ...实现省略...
  }
  
  public start() {
    this.isPlaying = true;
    this.playNext();
  }
  
  public pause() {
    this.isPlaying = false;
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }
  
  public next() {
    if (this.currentStepIndex < this.steps.length - 1) {
      this.currentStepIndex++;
      this.onStepChange(this.steps[this.currentStepIndex]);
    }
  }
  
  public previous() {
    if (this.currentStepIndex > 0) {
      this.currentStepIndex--;
      this.onStepChange(this.steps[this.currentStepIndex]);
    }
  }
  
  private playNext() {
    if (!this.isPlaying) return;
    
    this.next();
    
    if (this.currentStepIndex < this.steps.length - 1) {
      this.timer = setTimeout(() => this.playNext(), this.speed);
    } else {
      this.isPlaying = false;
    }
  }
  
  public setSpeed(speed: number) {
    this.speed = speed;
  }
  
  public reset() {
    this.pause();
    this.currentStepIndex = -1;
    this.onStepChange(null);
  }
}
```

## 可视化效果示例

在完整的实现中，解析过程将展示类似以下的可视化效果：

```
示例 JSON:  { "name": "John", "age": 30 }

步骤 1:
[{] "name": "John", "age": 30 }
↑
状态: 开始解析对象
代码: return p.parseObject(v)

步骤 2:
{ ["name"]: "John", "age": 30 }
  ↑↑↑↑↑↑
状态: 解析对象键
代码: key := p.parseString()

步骤 3:
{ "name"[: ]"John", "age": 30 }
        ↑↑
状态: 期望冒号
代码: if p.next() != ':' { return ParseMissColon }

... 更多步骤 ...

最终状态:
内存中的值:
{
  "type": "object",
  "value": {
    "name": {
      "type": "string",
      "value": "John"
    },
    "age": {
      "type": "number",
      "value": 30
    }
  }
}
```

## 在Web项目中的整合

这个可视化工具将作为Web项目的核心组件之一，整合到每个教程章节中。用户可以：

1. 阅读理论知识
2. 查看可视化演示
3. 修改示例并观察不同结果
4. 完成相关练习

通过这种互动方式，用户可以直观理解 JSON 解析的工作原理，加深对理论知识的理解。

## 下一步计划

1. 实现基本的 JSON 解析可视化
2. 添加各种 JSON 类型的专门可视化（数字、字符串、数组、对象等）
3. 增加错误处理的可视化
4. 添加内存使用和性能分析可视化
5. 整合到主web项目中 