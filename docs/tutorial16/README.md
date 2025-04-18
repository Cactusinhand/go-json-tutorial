# 第16章：命令行工具

在前面的章节中，我们已经实现了一个功能完整的 JSON 库。现在，我们将构建一个实用的命令行工具，让用户能够通过命令行界面来处理和操作 JSON 数据。这种工具在自动化脚本、数据处理管道和调试过程中非常有用。

## 16.1 设计命令行接口

一个好的命令行工具应该有清晰的使用方式和丰富的功能。对于 JSON 处理工具，我们可以设计以下功能：

1. **格式化** - 美化或压缩 JSON 数据
2. **验证** - 检查 JSON 数据的有效性
3. **查询** - 使用 JSONPath 查询 JSON 数据
4. **修改** - 使用 JSON Patch 修改 JSON 数据
5. **转换** - 在不同格式间转换（如 JSON 到 YAML）
6. **合并** - 合并多个 JSON 文件

我们将使用 Linux 风格的命令行参数，例如：

```
jsonutil [选项] <命令> [参数...]
```

## 16.2 程序入口

首先，我们创建程序的入口函数，解析命令行参数：

```c
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "leptjson.h"

void print_usage() {
    printf("用法：jsonutil [选项] <命令> [参数...]\n");
    printf("\n");
    printf("选项：\n");
    printf("  -h, --help     显示帮助信息\n");
    printf("  -v, --version  显示版本信息\n");
    printf("\n");
    printf("命令：\n");
    printf("  format    格式化 JSON 数据\n");
    printf("  validate  验证 JSON 数据\n");
    printf("  query     使用 JSONPath 查询 JSON 数据\n");
    printf("  patch     使用 JSON Patch 修改 JSON 数据\n");
    printf("  convert   在不同格式间转换\n");
    printf("  merge     合并多个 JSON 文件\n");
    printf("\n");
    printf("使用 'jsonutil <命令> --help' 查看特定命令的帮助。\n");
}

int main(int argc, char** argv) {
    if (argc < 2) {
        print_usage();
        return 0;
    }
    
    /* 处理全局选项 */
    if (strcmp(argv[1], "-h") == 0 || strcmp(argv[1], "--help") == 0) {
        print_usage();
        return 0;
    }
    
    if (strcmp(argv[1], "-v") == 0 || strcmp(argv[1], "--version") == 0) {
        printf("jsonutil 版本 1.0.0\n");
        return 0;
    }
    
    /* 处理命令 */
    if (strcmp(argv[1], "format") == 0) {
        return cmd_format(argc - 1, argv + 1);
    }
    else if (strcmp(argv[1], "validate") == 0) {
        return cmd_validate(argc - 1, argv + 1);
    }
    else if (strcmp(argv[1], "query") == 0) {
        return cmd_query(argc - 1, argv + 1);
    }
    else if (strcmp(argv[1], "patch") == 0) {
        return cmd_patch(argc - 1, argv + 1);
    }
    else if (strcmp(argv[1], "convert") == 0) {
        return cmd_convert(argc - 1, argv + 1);
    }
    else if (strcmp(argv[1], "merge") == 0) {
        return cmd_merge(argc - 1, argv + 1);
    }
    else {
        fprintf(stderr, "未知命令：%s\n", argv[1]);
        print_usage();
        return 1;
    }
}
```

## 16.3 实现格式化命令

格式化命令用于美化或压缩 JSON 数据，使其更易读或更节省空间：

```c
int cmd_format(int argc, char** argv) {
    int i;
    int indent = 2;          /* 默认缩进 */
    int compact = 0;         /* 是否压缩 */
    const char* input_file = NULL;
    const char* output_file = NULL;
    FILE* fin;
    FILE* fout;
    char* json_text = NULL;
    size_t json_size = 0;
    lept_value value;
    lept_context context;
    
    /* 解析命令行参数 */
    for (i = 1; i < argc; i++) {
        if (strcmp(argv[i], "--indent") == 0 && i + 1 < argc) {
            indent = atoi(argv[++i]);
            if (indent < 0) {
                fprintf(stderr, "缩进必须是非负数\n");
                return 1;
            }
        }
        else if (strcmp(argv[i], "--compact") == 0) {
            compact = 1;
        }
        else if (strcmp(argv[i], "-i") == 0 && i + 1 < argc) {
            input_file = argv[++i];
        }
        else if (strcmp(argv[i], "-o") == 0 && i + 1 < argc) {
            output_file = argv[++i];
        }
        else if (strcmp(argv[i], "--help") == 0) {
            printf("用法：jsonutil format [选项]\n");
            printf("\n");
            printf("选项：\n");
            printf("  --indent <n>  设置缩进空格数，默认为 2\n");
            printf("  --compact     压缩 JSON，不使用空格和换行\n");
            printf("  -i <文件>     输入文件，默认为标准输入\n");
            printf("  -o <文件>     输出文件，默认为标准输出\n");
            printf("  --help        显示帮助信息\n");
            return 0;
        }
    }
    
    /* 打开输入文件 */
    if (input_file) {
        fin = fopen(input_file, "r");
        if (!fin) {
            fprintf(stderr, "无法打开输入文件：%s\n", input_file);
            return 1;
        }
    }
    else {
        fin = stdin;
    }
    
    /* 读取 JSON 文本 */
    json_text = read_file(fin, &json_size);
    if (input_file)
        fclose(fin);
    
    if (!json_text) {
        fprintf(stderr, "读取输入失败\n");
        return 1;
    }
    
    /* 解析 JSON */
    lept_init(&value);
    if (lept_parse(&value, json_text) != LEPT_PARSE_OK) {
        fprintf(stderr, "解析 JSON 失败\n");
        free(json_text);
        return 1;
    }
    
    free(json_text);
    
    /* 打开输出文件 */
    if (output_file) {
        fout = fopen(output_file, "w");
        if (!fout) {
            fprintf(stderr, "无法打开输出文件：%s\n", output_file);
            lept_free(&value);
            return 1;
        }
    }
    else {
        fout = stdout;
    }
    
    /* 格式化 JSON */
    lept_context_init(&context);
    context.indent = compact ? 0 : indent;
    context.pretty = !compact;
    
    lept_stringify(&value, &context);
    
    /* 输出 JSON */
    fwrite(context.stack, 1, context.top, fout);
    fputc('\n', fout);
    
    if (output_file)
        fclose(fout);
    
    lept_context_free(&context);
    lept_free(&value);
    
    return 0;
}
```

## 16.4 实现验证命令

验证命令用于检查 JSON 数据的有效性，如果提供了 JSON Schema，还会根据 Schema 验证数据结构：

```c
int cmd_validate(int argc, char** argv) {
    int i;
    const char* input_file = NULL;
    const char* schema_file = NULL;
    FILE* fin;
    FILE* fschema = NULL;
    char* json_text = NULL;
    char* schema_text = NULL;
    size_t json_size = 0;
    size_t schema_size = 0;
    lept_value value;
    lept_value schema;
    lept_schema_result result;
    
    /* 解析命令行参数 */
    for (i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-i") == 0 && i + 1 < argc) {
            input_file = argv[++i];
        }
        else if (strcmp(argv[i], "--schema") == 0 && i + 1 < argc) {
            schema_file = argv[++i];
        }
        else if (strcmp(argv[i], "--help") == 0) {
            printf("用法：jsonutil validate [选项]\n");
            printf("\n");
            printf("选项：\n");
            printf("  -i <文件>      输入文件，默认为标准输入\n");
            printf("  --schema <文件> 用于验证的 JSON Schema 文件\n");
            printf("  --help         显示帮助信息\n");
            return 0;
        }
    }
    
    /* 打开输入文件 */
    if (input_file) {
        fin = fopen(input_file, "r");
        if (!fin) {
            fprintf(stderr, "无法打开输入文件：%s\n", input_file);
            return 1;
        }
    }
    else {
        fin = stdin;
    }
    
    /* 读取 JSON 文本 */
    json_text = read_file(fin, &json_size);
    if (input_file)
        fclose(fin);
    
    if (!json_text) {
        fprintf(stderr, "读取输入失败\n");
        return 1;
    }
    
    /* 解析 JSON */
    lept_init(&value);
    if (lept_parse(&value, json_text) != LEPT_PARSE_OK) {
        fprintf(stderr, "解析 JSON 失败\n");
        free(json_text);
        return 1;
    }
    
    free(json_text);
    
    /* 如果提供了 Schema，验证 JSON */
    if (schema_file) {
        fschema = fopen(schema_file, "r");
        if (!fschema) {
            fprintf(stderr, "无法打开 Schema 文件：%s\n", schema_file);
            lept_free(&value);
            return 1;
        }
        
        schema_text = read_file(fschema, &schema_size);
        fclose(fschema);
        
        if (!schema_text) {
            fprintf(stderr, "读取 Schema 失败\n");
            lept_free(&value);
            return 1;
        }
        
        lept_init(&schema);
        if (lept_parse(&schema, schema_text) != LEPT_PARSE_OK) {
            fprintf(stderr, "解析 Schema 失败\n");
            free(schema_text);
            lept_free(&value);
            return 1;
        }
        
        free(schema_text);
        
        if (!lept_validate_schema(&schema, &value, &result)) {
            fprintf(stderr, "验证失败：%s（在 %s）\n", 
                result.message ? result.message : "未知错误",
                result.path ? result.path : "根");
            lept_free(&schema);
            lept_free(&value);
            return 1;
        }
        
        lept_free(&schema);
    }
    
    printf("JSON 有效\n");
    lept_free(&value);
    
    return 0;
}
```

## 16.5 实现查询命令

查询命令使用 JSONPath 从 JSON 数据中提取特定部分：

```c
int cmd_query(int argc, char** argv) {
    int i;
    const char* input_file = NULL;
    const char* output_file = NULL;
    const char* query = NULL;
    FILE* fin;
    FILE* fout;
    char* json_text = NULL;
    size_t json_size = 0;
    lept_value value;
    lept_value result;
    lept_context context;
    
    /* 解析命令行参数 */
    for (i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-i") == 0 && i + 1 < argc) {
            input_file = argv[++i];
        }
        else if (strcmp(argv[i], "-o") == 0 && i + 1 < argc) {
            output_file = argv[++i];
        }
        else if (strcmp(argv[i], "-q") == 0 && i + 1 < argc) {
            query = argv[++i];
        }
        else if (strcmp(argv[i], "--help") == 0) {
            printf("用法：jsonutil query [选项]\n");
            printf("\n");
            printf("选项：\n");
            printf("  -i <文件>  输入文件，默认为标准输入\n");
            printf("  -o <文件>  输出文件，默认为标准输出\n");
            printf("  -q <查询>  JSONPath 查询表达式\n");
            printf("  --help     显示帮助信息\n");
            return 0;
        }
    }
    
    if (!query) {
        fprintf(stderr, "必须提供 JSONPath 查询表达式\n");
        return 1;
    }
    
    /* 打开输入文件 */
    if (input_file) {
        fin = fopen(input_file, "r");
        if (!fin) {
            fprintf(stderr, "无法打开输入文件：%s\n", input_file);
            return 1;
        }
    }
    else {
        fin = stdin;
    }
    
    /* 读取 JSON 文本 */
    json_text = read_file(fin, &json_size);
    if (input_file)
        fclose(fin);
    
    if (!json_text) {
        fprintf(stderr, "读取输入失败\n");
        return 1;
    }
    
    /* 解析 JSON */
    lept_init(&value);
    if (lept_parse(&value, json_text) != LEPT_PARSE_OK) {
        fprintf(stderr, "解析 JSON 失败\n");
        free(json_text);
        return 1;
    }
    
    free(json_text);
    
    /* 执行 JSONPath 查询 */
    lept_init(&result);
    if (lept_json_path_query(&value, query, &result) != LEPT_QUERY_OK) {
        fprintf(stderr, "JSONPath 查询失败\n");
        lept_free(&value);
        return 1;
    }
    
    /* 打开输出文件 */
    if (output_file) {
        fout = fopen(output_file, "w");
        if (!fout) {
            fprintf(stderr, "无法打开输出文件：%s\n", output_file);
            lept_free(&result);
            lept_free(&value);
            return 1;
        }
    }
    else {
        fout = stdout;
    }
    
    /* 输出查询结果 */
    lept_context_init(&context);
    context.indent = 2;
    context.pretty = 1;
    
    lept_stringify(&result, &context);
    
    fwrite(context.stack, 1, context.top, fout);
    fputc('\n', fout);
    
    if (output_file)
        fclose(fout);
    
    lept_context_free(&context);
    lept_free(&result);
    lept_free(&value);
    
    return 0;
}
```

## 16.6 实现补丁命令

补丁命令用于根据 JSON Patch 修改 JSON 数据：

```c
int cmd_patch(int argc, char** argv) {
    int i;
    const char* input_file = NULL;
    const char* patch_file = NULL;
    const char* output_file = NULL;
    FILE* fin;
    FILE* fpatch;
    FILE* fout;
    char* json_text = NULL;
    char* patch_text = NULL;
    size_t json_size = 0;
    size_t patch_size = 0;
    lept_value value;
    lept_value patch;
    lept_value result;
    lept_context context;
    
    /* 解析命令行参数 */
    for (i = 1; i < argc; i++) {
        if (strcmp(argv[i], "-i") == 0 && i + 1 < argc) {
            input_file = argv[++i];
        }
        else if (strcmp(argv[i], "-p") == 0 && i + 1 < argc) {
            patch_file = argv[++i];
        }
        else if (strcmp(argv[i], "-o") == 0 && i + 1 < argc) {
            output_file = argv[++i];
        }
        else if (strcmp(argv[i], "--help") == 0) {
            printf("用法：jsonutil patch [选项]\n");
            printf("\n");
            printf("选项：\n");
            printf("  -i <文件>  输入文件，默认为标准输入\n");
            printf("  -p <文件>  JSON Patch 文件\n");
            printf("  -o <文件>  输出文件，默认为标准输出\n");
            printf("  --help     显示帮助信息\n");
            return 0;
        }
    }
    
    if (!patch_file) {
        fprintf(stderr, "必须提供 JSON Patch 文件\n");
        return 1;
    }
    
    /* 打开输入文件 */
    if (input_file) {
        fin = fopen(input_file, "r");
        if (!fin) {
            fprintf(stderr, "无法打开输入文件：%s\n", input_file);
            return 1;
        }
    }
    else {
        fin = stdin;
    }
    
    /* 读取 JSON 文本 */
    json_text = read_file(fin, &json_size);
    if (input_file)
        fclose(fin);
    
    if (!json_text) {
        fprintf(stderr, "读取输入失败\n");
        return 1;
    }
    
    /* 解析 JSON */
    lept_init(&value);
    if (lept_parse(&value, json_text) != LEPT_PARSE_OK) {
        fprintf(stderr, "解析 JSON 失败\n");
        free(json_text);
        return 1;
    }
    
    free(json_text);
    
    /* 读取 Patch 文件 */
    fpatch = fopen(patch_file, "r");
    if (!fpatch) {
        fprintf(stderr, "无法打开 Patch 文件：%s\n", patch_file);
        lept_free(&value);
        return 1;
    }
    
    patch_text = read_file(fpatch, &patch_size);
    fclose(fpatch);
    
    if (!patch_text) {
        fprintf(stderr, "读取 Patch 失败\n");
        lept_free(&value);
        return 1;
    }
    
    /* 解析 Patch */
    lept_init(&patch);
    if (lept_parse(&patch, patch_text) != LEPT_PARSE_OK) {
        fprintf(stderr, "解析 Patch 失败\n");
        free(patch_text);
        lept_free(&value);
        return 1;
    }
    
    free(patch_text);
    
    /* 应用 Patch */
    lept_init(&result);
    if (lept_json_patch(&value, &patch, &result) != LEPT_PATCH_OK) {
        fprintf(stderr, "应用 Patch 失败\n");
        lept_free(&patch);
        lept_free(&value);
        return 1;
    }
    
    lept_free(&patch);
    lept_free(&value);
    
    /* 打开输出文件 */
    if (output_file) {
        fout = fopen(output_file, "w");
        if (!fout) {
            fprintf(stderr, "无法打开输出文件：%s\n", output_file);
            lept_free(&result);
            return 1;
        }
    }
    else {
        fout = stdout;
    }
    
    /* 输出结果 */
    lept_context_init(&context);
    context.indent = 2;
    context.pretty = 1;
    
    lept_stringify(&result, &context);
    
    fwrite(context.stack, 1, context.top, fout);
    fputc('\n', fout);
    
    if (output_file)
        fclose(fout);
    
    lept_context_free(&context);
    lept_free(&result);
    
    return 0;
}
```

## 16.7 辅助函数

在上面的命令实现中，我们使用了一个 `read_file` 函数来读取文件内容。下面是这个函数的实现：

```c
/* 读取整个文件内容 */
char* read_file(FILE* fp, size_t* size) {
    char* buffer;
    size_t capacity = 1024;
    size_t length = 0;
    
    buffer = (char*)malloc(capacity);
    if (!buffer)
        return NULL;
    
    while (!feof(fp)) {
        if (length + 1 >= capacity) {
            capacity *= 2;
            buffer = (char*)realloc(buffer, capacity);
            if (!buffer)
                return NULL;
        }
        
        length += fread(buffer + length, 1, capacity - length - 1, fp);
    }
    
    buffer[length] = '\0';
    *size = length;
    
    return buffer;
}
```

## 16.8 构建和使用

最后，我们需要编写 Makefile 或构建脚本，将我们的命令行工具和 JSON 库链接起来：

```makefile
CC = gcc
CFLAGS = -Wall -Wextra -g
LDFLAGS =

SRCS = jsonutil.c leptjson.c
OBJS = $(SRCS:.c=.o)
TARGET = jsonutil

.PHONY: all clean

all: $(TARGET)

$(TARGET): $(OBJS)
	$(CC) $(LDFLAGS) -o $@ $^

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

clean:
	rm -f $(OBJS) $(TARGET)
```

构建后，我们可以使用这个工具来处理 JSON 数据：

```bash
# 格式化 JSON 文件
./jsonutil format -i input.json -o formatted.json

# 验证 JSON 文件
./jsonutil validate -i input.json

# 使用 JSONPath 查询
./jsonutil query -i input.json -q "$.store.book[0].title"

# 应用 JSON Patch
./jsonutil patch -i input.json -p patch.json -o output.json
```

## 16.9 测试

我们应该为命令行工具编写测试，确保其正常工作：

```bash
#!/bin/bash

# 测试格式化命令
echo '{"a":1,"b":2}' > test_input.json
./jsonutil format -i test_input.json -o test_output.json
if ! diff test_output.json <(echo -e '{\n  "a": 1,\n  "b": 2\n}'); then
    echo "格式化测试失败"
    exit 1
fi

# 测试验证命令
./jsonutil validate -i test_input.json
if [ $? -ne 0 ]; then
    echo "验证测试失败"
    exit 1
fi

# 测试查询命令
./jsonutil query -i test_input.json -q "$.a" > test_query.json
if ! diff test_query.json <(echo -e '1'); then
    echo "查询测试失败"
    exit 1
fi

# 测试补丁命令
echo '[{"op":"add","path":"/c","value":3}]' > test_patch.json
./jsonutil patch -i test_input.json -p test_patch.json -o test_patched.json
if ! diff test_patched.json <(echo -e '{\n  "a": 1,\n  "b": 2,\n  "c": 3\n}'); then
    echo "补丁测试失败"
    exit 1
fi

echo "所有测试通过"
rm test_input.json test_output.json test_query.json test_patch.json test_patched.json
```

## 16.10 练习

1. 实现转换命令，支持 JSON 与 YAML 或 XML 之间的转换。
2. 实现合并命令，支持合并多个 JSON 文件。
3. 为命令行工具添加更多选项，如颜色输出、错误详情等。
4. 实现一个交互式 JSON 编辑器，允许用户通过命令行界面编辑 JSON 数据。
5. 为工具添加插件系统，允许用户扩展其功能。

## 16.11 下一步

在本章中，我们构建了一个功能完整的 JSON 命令行工具，使用户能够通过命令行界面处理 JSON 数据。在下一章中，我们将探讨如何增强 JSON 库的安全性，包括防御恶意输入、限制资源使用等，使我们的 JSON 库在生产环境中更加可靠和安全。 