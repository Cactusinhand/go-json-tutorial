# 从零开始的 JSON 库教程（十六）：命令行工具实现

## 命令行工具简介

在本章中，我们为 `leptjson` 库添加了命令行工具功能，使用户能够通过简单的命令行操作来处理和分析 JSON 数据。这个命令行工具提供了多种实用功能，如 JSON 解析、格式化、最小化、统计分析、路径查询和文档比较，大大提高了处理 JSON 数据的效率。

## 主要功能

* **解析 (parse)**: 验证 JSON 文件的格式是否合法
* **格式化 (format)**: 将 JSON 文件格式化，添加适当的缩进和换行，提高可读性
* **最小化 (minify)**: 移除 JSON 文件中的所有不必要的空格，减小文件大小
* **统计 (stats)**: 分析 JSON 文件，提供各种统计信息，如对象数量、数组数量、嵌套深度等
* **查找 (find)**: 使用简化的路径表达式在 JSON 文件中查找特定数据
* **JSONPath (path)**: 使用完整的 JSONPath 语法在 JSON 文件中查询数据，支持复杂查询和过滤条件
* **比较 (compare)**: 比较两个 JSON 文件，查找它们之间的差异
* **验证 (validate)**: 使用 JSON Schema 验证 JSON 文件的结构和内容
* **指针操作 (pointer)**: 使用 JSON Pointer 定位和操作 JSON 文档中的值
* **补丁应用 (patch)**: 使用 JSON Patch 对 JSON 文档应用一系列修改操作
* **合并补丁 (merge-patch)**: 使用 JSON Merge Patch 简化的方式合并 JSON 文档

## 使用方法

使用格式: `leptjson [选项] 命令 [参数]`

### 全局选项

* **--help, -h**: 显示帮助信息
* **--verbose, -v**: 显示详细输出
* **--version**: 显示版本信息

### 命令详解

#### parse - 解析并验证 JSON 文件

```bash
leptjson parse data.json
```

此命令将验证 `data.json` 文件是否是有效的 JSON 格式。如果文件有格式错误，将显示详细的错误信息。

#### format - 格式化 JSON 文件

```bash
leptjson format --indent=2 data.json formatted.json
```

将 `data.json` 格式化并保存为 `formatted.json`，使用 2 个空格作为缩进。如果不指定输出文件，将自动创建一个 `.formatted.json` 后缀的文件。

#### minify - 最小化 JSON 文件

```bash
leptjson minify data.json data.min.json
```

移除 `data.json` 中的所有空白字符，创建一个紧凑的 `data.min.json` 文件。

#### stats - 显示 JSON 统计信息

```bash
leptjson stats data.json
```

分析 `data.json` 文件并显示统计信息，包括：
* 文件大小
* 最大嵌套深度
* 对象数量
* 数组数量
* 键总数和最长键
* 字符串、数字、布尔值和 null 值的数量

可以使用 `--json` 选项以 JSON 格式输出统计信息：

```bash
leptjson stats --json data.json
```

#### find - 查找 JSON 路径

```bash
leptjson find data.json "$.store.book[0].title"
```

使用 JSONPath 表达式 `$.store.book[0].title` 在 `data.json` 中查找匹配的值。

可以使用 `--output` 选项指定输出格式：
* `compact`: 紧凑 JSON（默认）
* `pretty`: 格式化 JSON
* `raw`: 对于简单值，仅输出值本身

```bash
leptjson find --output=pretty data.json "$.store.book[*]"
```

#### path - 使用 JSONPath 查询 JSON 数据

```bash
leptjson path [选项] 文件 JSONPATH表达式
```

使用完整的 JSONPath 语法从 JSON 文件中提取数据。相比 `find` 命令，`path` 命令支持更强大的查询功能。

选项:
- `--output=FORMAT`: 设置输出格式，可选值有 compact (紧凑), pretty (美化), raw (原始), table (表格)
- `--all`: 显示所有匹配结果(默认只显示前10个)
- `--csv=FILE`: 将结果保存为 CSV 文件
- `--no-path`: 不在输出中显示路径信息

支持的 JSONPath 语法:
- `$`: 根对象
- `.property`: 子属性访问
- `['property']`: 带引号的属性访问
- `[index]`: 数组索引访问
- `[start:end:step]`: 数组切片
- `*`: 通配符，匹配所有成员
- `..property`: 递归下降，匹配任意深度的属性
- `[?(@.prop > 10)]`: 过滤表达式
- `[?(@.prop)]`: 存在性检查
- `[?(@.name == 'value')]`: 相等性检查
- `['a','b']`: 多属性选择

示例:
```bash
# 查找所有书籍的作者
leptjson path books.json "$.store.book[*].author"

# 查询价格小于10的所有书籍
leptjson path --output=table books.json "$..book[?(@.price < 10)]"

# 查找所有价格并导出为CSV
leptjson path --csv=prices.csv books.json "$..price"

# 递归查找所有ID
leptjson path users.json "$..id"
```

#### compare - 比较两个 JSON 文件

```bash
leptjson compare file1.json file2.json
```

比较 `file1.json` 和 `file2.json`，显示它们之间的所有差异，包括类型不匹配、值不同和缺失/额外的键。

可以使用 `--json` 选项以 JSON 格式输出差异：

```bash
leptjson compare --json file1.json file2.json
```

#### validate - 使用 JSON Schema 验证 JSON 文件

```bash
leptjson validate schema.json data.json
```

使用 `schema.json` 中定义的 JSON Schema 验证 `data.json` 文件。如果验证失败，将显示详细的错误信息，包括验证失败的位置和原因。

可以使用 `--format` 选项指定输出格式：
* `text`: 人类可读的文本格式（默认）
* `json`: 机器可读的 JSON 格式

```bash
leptjson validate --format=json schema.json data.json
```

#### pointer - 使用 JSON Pointer 操作 JSON 文件

```bash
leptjson pointer data.json "/users/0/name"
```

使用 JSON Pointer (RFC 6901) 查找 `data.json` 中位于路径 `/users/0/name` 的值。

可以使用 `--operation` 选项指定操作类型：
* `get`: 获取值（默认）
* `add`: 添加或替换值
* `remove`: 删除值
* `replace`: 替换值

对于 `add` 和 `replace` 操作，需要使用 `--value` 选项指定要设置的值：

```bash
leptjson pointer --operation=replace --value="John" data.json "/users/0/name"
```

对于修改操作，可以使用 `--output` 选项指定输出文件，默认会覆盖原文件：

```bash
leptjson pointer --operation=add --value="admin" --output=new.json data.json "/users/0/role"
```

#### patch - 使用 JSON Patch 应用修改

```bash
leptjson patch patch.json data.json output.json
```

将 `patch.json` 中定义的 JSON Patch 操作应用到 `data.json`，并将结果保存到 `output.json`。

JSON Patch (RFC 6902) 支持以下操作：
* `add`: 添加值
* `remove`: 删除值
* `replace`: 替换值
* `move`: 移动值
* `copy`: 复制值
* `test`: 测试值是否匹配

可以使用 `--test` 选项仅测试补丁是否可以应用，而不实际修改文件：

```bash
leptjson patch --test patch.json data.json
```

使用 `--in-place` 选项直接修改原文件，而不是创建新文件：

```bash
leptjson patch --in-place patch.json data.json
```

#### merge-patch - 使用 JSON Merge Patch 合并文档

```bash
leptjson merge-patch patch.json data.json output.json
```

将 `patch.json` 中定义的 JSON Merge Patch 应用到 `data.json`，并将结果保存到 `output.json`。

JSON Merge Patch (RFC 7396) 是一种比 JSON Patch 更简单的 JSON 文档合并方式，其主要规则：
* 如果补丁中的值为 null，则从目标中删除该字段
* 如果补丁中的值不为 null，则替换目标中的对应字段
* 如果两边都是对象，则递归合并
* 对于数组，直接替换而不是合并

使用 `--in-place` 选项直接修改原文件，而不是创建新文件：

```bash
leptjson merge-patch --in-place changes.json data.json
```

## 使用示例

### 解析并格式化 JSON 文件

```bash
# 验证 JSON 文件
leptjson parse data.json

# 格式化 JSON 文件
leptjson format --indent=4 data.json pretty.json

# 最小化 JSON 文件
leptjson minify pretty.json data.min.json
```

### 分析和验证 JSON 文件

```bash
# 获取 JSON 统计信息
leptjson stats data.json

# 验证 JSON 是否符合 Schema
leptjson validate user-schema.json user.json

# 查找特定数据
leptjson find data.json "$.users[?(@.age>30)].name"

# 比较两个 JSON 文件
leptjson compare original.json updated.json
```

### 修改 JSON 文件

```bash
# 使用 JSON Pointer 修改特定值
leptjson pointer --operation=replace --value=true data.json "/user/active"

# 使用 JSON Patch 应用多个修改
leptjson patch updates.json data.json data-updated.json

# 使用 JSON Merge Patch 合并文档
leptjson merge-patch merge.json data.json data-merged.json
```

### 复杂查询示例

```bash
# 使用JSONPath查询所有价格小于10的书籍，以表格形式显示
leptjson path --output=table library.json "$..book[?(@.price < 10)]"

# 使用JSONPath递归查找所有作者并保存为CSV文件
leptjson path --csv=authors.csv library.json "$..author"

# 使用JSON Pointer访问特定路径
leptjson pointer users.json "/users/0/name"
```

## 实现细节

命令行工具使用 Go 标准库的 `flag` 包实现命令行参数解析，并利用我们在前面章节中实现的 JSON 解析、序列化、JSONPath、JSON Schema、JSON Pointer 和 JSON Patch 功能。主要组件包括：

1. **JSON 解析器**: 使用我们的 `Parse` 函数验证 JSON 文件
2. **格式化器**: 实现了带缩进的 JSON 输出
3. **统计分析器**: 递归遍历 JSON 结构，计算各种统计信息
4. **JSONPath 查询器**: 使用我们的 JSONPath 实现查找指定路径的值
5. **比较器**: 递归比较两个 JSON 结构，检测所有差异
6. **JSON Schema 验证器**: 实现 JSON Schema 验证功能
7. **JSON Pointer 处理器**: 实现 RFC 6901 定义的 JSON Pointer 功能
8. **JSON Patch 应用器**: 实现 RFC 6902 定义的 JSON Patch 功能

## 测试

在 `go-json-tutorial/tutorial16` 目录下运行测试：

```bash
go test
```

## 构建和安装

在 `go-json-tutorial/tutorial16/main` 目录下构建命令行工具：

```bash
go build -o leptjson
```

或使用 `go install` 安装到系统路径：

```bash
go install
```

## 参考资料

- [Go 标准库 flag 包文档](https://golang.org/pkg/flag/)
- [JSONPath 语法规范](https://goessner.net/articles/JsonPath/)
- [JSON 格式规范](https://www.json.org/)
- [RFC 6901: JSON Pointer](https://tools.ietf.org/html/rfc6901)
- [RFC 6902: JSON Patch](https://tools.ietf.org/html/rfc6902)
- [JSON Schema 规范](https://json-schema.org/specification.html)
- [RFC 7396: JSON Merge Patch](https://tools.ietf.org/html/rfc7396) 