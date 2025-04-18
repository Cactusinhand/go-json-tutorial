# JSON 解析器开发练习指南

本文档提供一系列分级的练习和实践项目，帮助学习者循序渐进地掌握 JSON 解析器的开发技能。

## 初级练习（基础概念与解析）

### 练习 1-1：JSON 词法分析器
**目标**：实现一个简单的 JSON 词法分析器，将 JSON 字符串转换为标记序列。
**要求**：
- 识别基本标记：`{`, `}`, `[`, `]`, `:`, `,`, `"string"`, `number`, `true`, `false`, `null`
- 正确处理空白字符
- 能够报告基本的词法错误（如未闭合的字符串）

**提示**：
```c
enum token_type {
    TOKEN_NULL,
    TOKEN_TRUE,
    TOKEN_FALSE,
    TOKEN_NUMBER,
    TOKEN_STRING,
    TOKEN_ARRAY_START,
    TOKEN_ARRAY_END,
    TOKEN_OBJECT_START,
    TOKEN_OBJECT_END,
    TOKEN_COMMA,
    TOKEN_COLON,
    TOKEN_END
};

struct token {
    enum token_type type;
    char* start;
    size_t length;
    // 可能需要添加额外字段，如行号、列号等
};
```

### 练习 1-2：解析基本类型
**目标**：实现解析 JSON 基本类型（null、布尔值、数字）的函数。
**要求**：
- 正确解析 `null`、`true` 和 `false`
- 解析整数和浮点数
- 处理可能的数值格式错误

**提示**：
```c
typedef enum {
    JSON_NULL,
    JSON_BOOLEAN,
    JSON_NUMBER,
    JSON_STRING,
    JSON_ARRAY,
    JSON_OBJECT
} json_type;

typedef struct {
    json_type type;
    union {
        int boolean;
        double number;
        // 其他类型将在后续练习中添加
    } value;
} json_value;

int parse_null(const char* json, json_value* value);
int parse_boolean(const char* json, json_value* value);
int parse_number(const char* json, json_value* value);
```

### 练习 1-3：字符串解析
**目标**：实现 JSON 字符串的解析，包括转义字符的处理。
**要求**：
- 处理基本转义序列：`\"`, `\\`, `\/`, `\b`, `\f`, `\n`, `\r`, `\t`
- 处理 Unicode 转义序列：`\uXXXX`
- 正确处理 UTF-8 编码
- 检测并报告无效的转义序列

**示例代码框架**：
```c
// 字符串表示结构
typedef struct {
    char* s;
    size_t len;
} json_string;

// 字符串解析函数
int parse_string(const char* json, json_string* str);

// Unicode 字符处理函数
int parse_unicode_hex(const char* s, unsigned* u);
```

### 练习 1-4：完整解析器实现
**目标**：整合前面的练习，实现一个完整的 JSON 解析器。
**要求**：
- 解析复合类型：数组和对象
- 实现递归下降解析算法
- 提供错误报告机制

**思路指导**：
1. 首先实现一个解析单个 JSON 值的函数
2. 然后实现数组和对象的解析函数
3. 使用递归方式处理嵌套结构
4. 添加错误处理机制

## 中级练习（功能扩展与优化）

### 练习 2-1：错误处理增强
**目标**：增强解析器的错误处理能力。
**要求**：
- 提供详细的错误信息，包括错误类型、位置和上下文
- 实现错误恢复机制，尝试在出错后继续解析
- 设计并实现一个完整的错误报告系统

**示例实现**：
```c
typedef enum {
    PARSE_OK,
    PARSE_EXPECT_VALUE,
    PARSE_INVALID_VALUE,
    PARSE_ROOT_NOT_SINGULAR,
    PARSE_NUMBER_TOO_BIG,
    PARSE_MISS_QUOTATION_MARK,
    PARSE_INVALID_STRING_ESCAPE,
    PARSE_INVALID_STRING_CHAR,
    PARSE_INVALID_UNICODE_HEX,
    PARSE_INVALID_UNICODE_SURROGATE,
    PARSE_MISS_COMMA_OR_SQUARE_BRACKET,
    PARSE_MISS_KEY,
    PARSE_MISS_COLON,
    PARSE_MISS_COMMA_OR_CURLY_BRACKET
} parse_error;

typedef struct {
    parse_error code;
    size_t position;
    const char* context;
    size_t context_length;
} parse_error_info;

void error_report(parse_error_info* error);
```

### 练习 2-2：内存管理优化
**目标**：优化解析器的内存管理。
**要求**：
- 实现简单的内存池
- 优化字符串的内存分配
- 实现小字符串优化（SSO）

**参考实现**：
```c
// 简单内存池
typedef struct {
    char* buffer;
    size_t size;
    size_t used;
} memory_pool;

void* memory_pool_alloc(memory_pool* pool, size_t size);
void memory_pool_free(memory_pool* pool);

// 小字符串优化
#define SSO_MAX_SIZE 15

typedef struct {
    union {
        struct {
            char* ptr;
            size_t length;
        } long_str;
        struct {
            char buffer[SSO_MAX_SIZE + 1];
            unsigned char length;
        } short_str;
    } data;
    unsigned char is_short;
} json_string_sso;
```

### 练习 2-3：DOM API 设计
**目标**：设计并实现 JSON 文档对象模型（DOM）API。
**要求**：
- 提供访问 JSON 值的方法（获取类型、获取值）
- 提供修改 JSON 值的方法（设置值、增加/删除/修改数组元素或对象成员）
- 设计合理的内存管理策略

**示例 API**：
```c
// 获取值类型
json_type json_get_type(const json_value* value);

// 获取基本类型值
int json_get_boolean(const json_value* value);
double json_get_number(const json_value* value);
const char* json_get_string(const json_value* value);

// 数组操作
size_t json_get_array_size(const json_value* value);
json_value* json_get_array_element(const json_value* value, size_t index);
void json_set_array_element(json_value* value, size_t index, json_value* element);
void json_array_push_back(json_value* value, json_value* element);

// 对象操作
size_t json_get_object_size(const json_value* value);
const char* json_get_object_key(const json_value* value, size_t index);
json_value* json_get_object_value(const json_value* value, size_t index);
json_value* json_find_object_value(const json_value* value, const char* key);
void json_set_object_value(json_value* value, const char* key, json_value* val);
```

### 练习 2-4：性能优化
**目标**：提高解析器的性能。
**要求**：
- 使用 `strtod()` 或自定义算法优化数字解析
- 使用查找表优化字符类型判断
- 减少函数调用层级
- 实现基准测试框架，对优化前后进行性能对比

**性能测试框架示例**：
```c
#include <time.h>

typedef struct {
    clock_t start, end;
    double cpu_time_used;
} benchmark;

void benchmark_start(benchmark* b) {
    b->start = clock();
}

void benchmark_end(benchmark* b) {
    b->end = clock();
    b->cpu_time_used = ((double) (b->end - b->start)) / CLOCKS_PER_SEC;
}

void benchmark_parse(const char* json, size_t length, int iterations) {
    benchmark b;
    benchmark_start(&b);
    
    for (int i = 0; i < iterations; i++) {
        json_value value;
        parse(json, length, &value);
        json_free(&value);
    }
    
    benchmark_end(&b);
    printf("Parse: %.9f s\n", b.cpu_time_used);
}
```

## 高级练习（完整应用与扩展）

### 练习 3-1：JSON Schema 验证器
**目标**：实现符合 JSON Schema 规范的验证器。
**要求**：
- 支持基本类型验证
- 支持数值范围验证
- 支持字符串格式和模式验证
- 支持数组和对象的结构验证
- 支持复杂组合验证（allOf, anyOf, oneOf, not）

**实现建议**：
```c
typedef enum {
    SCHEMA_TYPE_ANY,
    SCHEMA_TYPE_NULL,
    SCHEMA_TYPE_BOOLEAN,
    SCHEMA_TYPE_NUMBER,
    SCHEMA_TYPE_STRING,
    SCHEMA_TYPE_ARRAY,
    SCHEMA_TYPE_OBJECT
} schema_type;

typedef struct schema_validator schema_validator;

schema_validator* schema_create(const json_value* schema);
void schema_free(schema_validator* validator);
int schema_validate(schema_validator* validator, const json_value* value, char** error_msg);
```

### 练习 3-2：JSONPath 查询实现
**目标**：实现 JSONPath 查询语言。
**要求**：
- 支持基本路径表达式（如 `$.store.book[0].title`）
- 支持通配符（如 `$.store.book[*].author`）
- 支持递归搜索（如 `$..author`）
- 支持过滤表达式（如 `$.store.book[?(@.price < 10)]`）

**示例实现**：
```c
typedef struct jsonpath_query jsonpath_query;

jsonpath_query* jsonpath_compile(const char* path);
void jsonpath_free(jsonpath_query* query);
json_value* jsonpath_execute(jsonpath_query* query, const json_value* root);
```

### 练习 3-3：完整 JSON 库
**目标**：将前面的练习整合为一个完整的 JSON 处理库。
**要求**：
- 整合所有功能：解析、生成、验证、查询
- 设计清晰一致的 API
- 编写详细的文档和示例
- 添加单元测试和性能测试
- 遵循良好的工程实践

**项目结构建议**：
```
|- json/
|  |- include/
|  |  |- json.h             // 主头文件
|  |  |- json_parser.h      // 解析器
|  |  |- json_generator.h   // 生成器
|  |  |- json_schema.h      // Schema 验证
|  |  |- json_path.h        // JSONPath
|  |- src/
|  |  |- json_parser.c
|  |  |- json_generator.c
|  |  |- json_schema.c
|  |  |- json_path.c
|  |- test/
|  |  |- test_parser.c
|  |  |- test_generator.c
|  |  |- test_schema.c
|  |  |- test_path.c
|  |  |- test_performance.c
|  |- examples/
|  |  |- basic_usage.c
|  |  |- schema_validation.c
|  |  |- path_query.c
|  |- doc/
|  |  |- api_reference.md
|  |  |- user_guide.md
|  |- CMakeLists.txt
```

## 实战项目

### 项目 1：JSON 配置文件解析器
**目标**：实现一个应用程序配置文件解析器，使用 JSON 格式。
**要求**：
- 读取 JSON 格式的配置文件
- 提供类型安全的配置访问 API
- 支持默认值
- 支持配置验证
- 支持配置热加载

### 项目 2：JSON-RPC 服务框架
**目标**：实现一个基于 JSON-RPC 协议的服务框架。
**要求**：
- 符合 JSON-RPC 2.0 规范
- 支持方法注册和调用
- 支持请求批处理
- 提供客户端和服务器实现
- 实现简单的错误处理机制

### 项目 3：JSON 数据库
**目标**：实现一个简单的基于 JSON 的数据库。
**要求**：
- 支持 JSON 文档的增删改查
- 实现简单的索引机制
- 支持类似 SQL 的查询语言
- 提供事务支持
- 实现数据持久化

## 学习资源

### 参考文档
- [JSON 官方网站](https://www.json.org/)
- [RFC 8259: The JavaScript Object Notation (JSON) Data Interchange Format](https://tools.ietf.org/html/rfc8259)
- [JSON Schema 规范](https://json-schema.org/)
- [JSONPath 规范](https://goessner.net/articles/JsonPath/)
- [JSON-RPC 2.0 规范](https://www.jsonrpc.org/specification)

### 开源项目参考
- [RapidJSON](https://github.com/Tencent/rapidjson)
- [cJSON](https://github.com/DaveGamble/cJSON)
- [Jansson](https://github.com/akheron/jansson)
- [JSON for Modern C++](https://github.com/nlohmann/json)

### 书籍推荐
- 《数据结构与算法分析》
- 《编程珠玑》
- 《深入理解计算机系统》
- 《设计模式：可复用面向对象软件的基础》

## 总结

通过完成这些练习和项目，你将逐步掌握 JSON 解析器的设计和实现技巧，从基础的解析功能到高级的查询和验证功能，最终能够构建一个完整的 JSON 处理库。每个练习都提供了详细的指导和示例代码，帮助你循序渐进地学习。

记住，解析器开发是一个综合性的工程，需要考虑正确性、健壮性、性能和可用性等多个方面。通过实践和不断改进，你将能够开发出高质量的 JSON 处理工具。 