# 第15章：结构体与标签支持

在前面的章节中，我们已经实现了一个功能完整的 JSON 库，包括解析、生成、访问、修改等基本功能，以及一些高级特性。本章我们将探讨如何将 JSON 数据与编程语言中的结构体类型进行转换，这对于实际应用中的数据处理非常重要。

## 15.1 JSON 数据与结构体的映射

在实际应用中，我们通常需要将 JSON 数据转换为特定的数据结构，或将程序中的数据结构序列化为 JSON 格式。这种转换可以通过手动设置每个字段来实现，但这种方式繁琐且容易出错。因此，我们需要设计一种自动映射机制。

### 15.1.1 反射机制

反射是一种在运行时检查、操作代码中的类型和值的机制。通过反射，我们可以在运行时了解结构体的字段名称、类型等信息，从而实现自动映射。

```c
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "leptjson.h"

/* 简单的反射机制 */
typedef enum {
    LEPT_REFLECT_TYPE_INT,
    LEPT_REFLECT_TYPE_DOUBLE,
    LEPT_REFLECT_TYPE_STRING,
    LEPT_REFLECT_TYPE_BOOLEAN,
    LEPT_REFLECT_TYPE_STRUCT,
    LEPT_REFLECT_TYPE_ARRAY    /* 新增：数组类型支持 */
} lept_reflect_type;

typedef struct lept_reflect_field lept_reflect_field;

struct lept_reflect_field {
    const char* name;           /* 字段名 */
    size_t offset;              /* 字段在结构体中的偏移量 */
    lept_reflect_type type;     /* 字段类型 */
    const void* type_info;      /* 对于 STRUCT 类型，指向子结构体的反射信息 */
    const char* json_name;      /* JSON 中的字段名，可与结构体字段名不同 */
    int ignore;                 /* 是否在序列化/反序列化时忽略此字段 */
    int is_required;            /* 新增：字段是否必需 */
    const char* default_value;  /* 新增：默认值 */
};

typedef struct {
    size_t struct_size;         /* 结构体大小 */
    size_t field_count;         /* 字段数量 */
    lept_reflect_field* fields; /* 字段信息数组 */
    const char* struct_name;    /* 新增：结构体名称，用于更好的错误信息 */
} lept_reflect_struct;
```

### 15.1.2 结构体定义宏

为了简化结构体反射信息的定义，我们可以设计一系列宏：

```c
/* 开始定义结构体反射信息 */
#define LEPT_STRUCT_BEGIN(type) \
    lept_reflect_struct type##_reflect = { \
        sizeof(type), \
        0, \
        NULL, \
        #type \
    }; \
    void type##_reflect_init() { \
        static lept_reflect_field fields[] = {

/* 定义整数字段 */
#define LEPT_STRUCT_INT(type, field) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_INT, NULL, NULL, 0, 0, NULL},

/* 定义浮点数字段 */
#define LEPT_STRUCT_DOUBLE(type, field) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_DOUBLE, NULL, NULL, 0, 0, NULL},

/* 定义字符串字段 */
#define LEPT_STRUCT_STRING(type, field) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_STRING, NULL, NULL, 0, 0, NULL},

/* 定义布尔字段 */
#define LEPT_STRUCT_BOOLEAN(type, field) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_BOOLEAN, NULL, NULL, 0, 0, NULL},

/* 定义结构体字段 */
#define LEPT_STRUCT_STRUCT(type, field, field_type) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_STRUCT, &field_type##_reflect, NULL, 0, 0, NULL},

/* 定义带标签的整数字段 */
#define LEPT_STRUCT_INT_TAG(type, field, json_name, ignore) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_INT, NULL, json_name, ignore, 0, NULL},

/* 定义带标签和默认值的整数字段 */
#define LEPT_STRUCT_INT_FULL(type, field, json_name, ignore, required, default_val) \
    {#field, offsetof(type, field), LEPT_REFLECT_TYPE_INT, NULL, json_name, ignore, required, default_val},

/* 其他类型的带标签宏类似... */

/* 结束定义结构体反射信息 */
#define LEPT_STRUCT_END(type) \
        }; \
        type##_reflect.field_count = sizeof(fields) / sizeof(fields[0]); \
        type##_reflect.fields = fields; \
    }
```

## 15.2 标签支持

在许多编程语言中，我们可以通过标签（tag）来为结构体字段提供额外的元数据。这些标签可以用于自定义 JSON 序列化和反序列化的行为，如字段重命名、忽略特定字段等。

### 15.2.1 标签解析

标签通常包含以下选项：

- **名称映射**：指定 JSON 中的字段名与结构体字段名不同
- **忽略选项**：在特定条件下忽略字段
- **必需字段**：指定字段必须存在于 JSON 中
- **默认值**：当 JSON 中字段不存在时使用的默认值
- **验证规则**：对字段值进行验证的规则

### 15.2.2 C++ 标签支持

对于 C++，我们可以使用模板和属性来提供更优雅的标签支持：

```cpp
#include <string>
#include <typeinfo>
#include <map>

// 字段元数据
template <typename T>
struct FieldMeta {
    std::string json_name;
    bool ignore = false;
    bool required = false;
    T default_value = T();
    
    FieldMeta<T>& name(const std::string& name) {
        json_name = name;
        return *this;
    }
    
    FieldMeta<T>& omit() {
        ignore = true;
        return *this;
    }
    
    FieldMeta<T>& require() {
        required = true;
        return *this;
    }
    
    FieldMeta<T>& defaultVal(const T& val) {
        default_value = val;
        return *this;
    }
};

// 类型映射
template <typename Class>
class TypeMap {
private:
    std::map<std::string, void*> meta_map;
    
public:
    template <typename T>
    TypeMap& field(T Class::* member, FieldMeta<T> meta = FieldMeta<T>()) {
        // 存储字段元数据
        meta_map[typeid(member).name()] = new FieldMeta<T>(meta);
        return *this;
    }
    
    template <typename T>
    FieldMeta<T>* get_meta(T Class::* member) {
        return static_cast<FieldMeta<T>*>(meta_map[typeid(member).name()]);
    }
};

// 使用示例
struct Person {
    std::string name;
    int age;
    bool is_student;
    std::string email;
};

TypeMap<Person> person_map;

void init_reflection() {
    person_map
        .field(&Person::name, FieldMeta<std::string>().require())
        .field(&Person::age, FieldMeta<int>().name("person_age"))
        .field(&Person::is_student)
        .field(&Person::email, FieldMeta<std::string>().name("email_address").defaultVal(""));
}
```

## 15.3 序列化与反序列化

有了反射机制和标签支持，我们可以实现通用的序列化和反序列化函数：

### 15.3.1 序列化

```c
int lept_serialize_struct(const void* structure, const lept_reflect_struct* reflect, lept_value* json_value) {
    size_t i;
    lept_set_object(json_value);
    
    for (i = 0; i < reflect->field_count; i++) {
        const lept_reflect_field* field = &reflect->fields[i];
        const void* field_addr = (const char*)structure + field->offset;
        lept_value field_value;
        
        /* 如果字段被标记为忽略，则跳过 */
        if (field->ignore)
            continue;
            
        lept_init(&field_value);
        
        /* 根据字段类型，设置 JSON 值 */
        switch (field->type) {
            case LEPT_REFLECT_TYPE_INT:
                lept_set_number(&field_value, *(int*)field_addr);
                break;
            case LEPT_REFLECT_TYPE_DOUBLE:
                lept_set_number(&field_value, *(double*)field_addr);
                break;
            case LEPT_REFLECT_TYPE_STRING:
                {
                    const char* str = *(const char**)field_addr;
                    if (str)
                        lept_set_string(&field_value, str, strlen(str));
                    else
                        lept_set_null(&field_value);
                }
                break;
            case LEPT_REFLECT_TYPE_BOOLEAN:
                lept_set_boolean(&field_value, *(int*)field_addr);
                break;
            case LEPT_REFLECT_TYPE_STRUCT:
                {
                    const lept_reflect_struct* sub_reflect = (const lept_reflect_struct*)field->type_info;
                    lept_serialize_struct(field_addr, sub_reflect, &field_value);
                }
                break;
            case LEPT_REFLECT_TYPE_ARRAY:
                /* 处理数组类型 */
                /* ... */
                break;
            default:
                lept_free(&field_value);
                return LEPT_SERIALIZE_INVALID_TYPE;
        }
        
        /* 将字段添加到 JSON 对象中 */
        lept_set_object_value(json_value, field->json_name ? field->json_name : field->name, &field_value);
    }
    
    return LEPT_SERIALIZE_OK;
}
```

### 15.3.2 反序列化

```c
int lept_deserialize_struct(void* structure, const lept_reflect_struct* reflect, const lept_value* json_value) {
    size_t i;
    
    /* 确保 JSON 值是对象类型 */
    if (json_value->type != LEPT_OBJECT)
        return LEPT_DESERIALIZE_INVALID_TYPE;
    
    /* 遍历结构体的每个字段 */
    for (i = 0; i < reflect->field_count; i++) {
        const lept_reflect_field* field = &reflect->fields[i];
        void* field_addr = (char*)structure + field->offset;
        
        /* 如果字段被标记为忽略，则跳过 */
        if (field->ignore)
            continue;
            
        /* 在 JSON 对象中查找字段 */
        const char* json_field_name = field->json_name ? field->json_name : field->name;
        const lept_value* json_field = lept_find_object_value(json_value, json_field_name);
        
        /* 如果字段不存在但是必需的，返回错误 */
        if (!json_field && field->is_required)
            return LEPT_DESERIALIZE_MISSING_REQUIRED;
            
        /* 如果字段不存在但有默认值，使用默认值 */
        if (!json_field && field->default_value) {
            /* 设置默认值 */
            switch (field->type) {
                case LEPT_REFLECT_TYPE_INT:
                    *(int*)field_addr = atoi(field->default_value);
                    break;
                case LEPT_REFLECT_TYPE_DOUBLE:
                    *(double*)field_addr = atof(field->default_value);
                    break;
                case LEPT_REFLECT_TYPE_STRING:
                    {
                        char* str = strdup(field->default_value);
                        *(char**)field_addr = str;
                    }
                    break;
                case LEPT_REFLECT_TYPE_BOOLEAN:
                    *(int*)field_addr = strcmp(field->default_value, "true") == 0;
                    break;
                default:
                    break;
            }
            continue;
        }
        
        /* 如果字段不存在，跳过 */
        if (!json_field)
            continue;
            
        /* 根据字段类型，设置结构体字段值 */
        switch (field->type) {
            case LEPT_REFLECT_TYPE_INT:
                if (json_field->type == LEPT_NUMBER)
                    *(int*)field_addr = (int)json_field->u.n;
                break;
            case LEPT_REFLECT_TYPE_DOUBLE:
                if (json_field->type == LEPT_NUMBER)
                    *(double*)field_addr = json_field->u.n;
                break;
            case LEPT_REFLECT_TYPE_STRING:
                if (json_field->type == LEPT_STRING) {
                    char* str = malloc(json_field->u.s.len + 1);
                    memcpy(str, json_field->u.s.s, json_field->u.s.len);
                    str[json_field->u.s.len] = '\0';
                    *(char**)field_addr = str;
                }
                break;
            case LEPT_REFLECT_TYPE_BOOLEAN:
                if (json_field->type == LEPT_TRUE)
                    *(int*)field_addr = 1;
                else if (json_field->type == LEPT_FALSE)
                    *(int*)field_addr = 0;
                break;
            case LEPT_REFLECT_TYPE_STRUCT:
                if (json_field->type == LEPT_OBJECT) {
                    const lept_reflect_struct* sub_reflect = (const lept_reflect_struct*)field->type_info;
                    lept_deserialize_struct(field_addr, sub_reflect, json_field);
                }
                break;
            case LEPT_REFLECT_TYPE_ARRAY:
                /* 处理数组类型 */
                /* ... */
                break;
            default:
                return LEPT_DESERIALIZE_INVALID_TYPE;
        }
    }
    
    return LEPT_DESERIALIZE_OK;
}
```

## 15.4 类型安全的序列化与反序列化

为了提高类型安全性，我们可以使用模板编程来确保在编译时捕获类型错误：

```cpp
// 类型安全的序列化函数
template <typename T>
int serialize(const T& obj, lept_value* json) {
    // 确保T类型的反射信息已经初始化
    T_reflect_init();
    return lept_serialize_struct(&obj, &T_reflect, json);
}

// 类型安全的反序列化函数
template <typename T>
int deserialize(T& obj, const lept_value* json) {
    // 确保T类型的反射信息已经初始化
    T_reflect_init();
    return lept_deserialize_struct(&obj, &T_reflect, json);
}

// 辅助函数：直接从JSON字符串反序列化
template <typename T>
int from_json(T& obj, const char* json_str) {
    lept_value v;
    lept_init(&v);
    int ret = lept_parse(&v, json_str);
    if (ret != LEPT_PARSE_OK) {
        lept_free(&v);
        return ret;
    }
    
    ret = deserialize(obj, &v);
    lept_free(&v);
    return ret;
}

// 辅助函数：直接序列化为JSON字符串
template <typename T>
char* to_json(const T& obj) {
    lept_value v;
    lept_init(&v);
    int ret = serialize(obj, &v);
    if (ret != LEPT_SERIALIZE_OK) {
        lept_free(&v);
        return NULL;
    }
    
    char* json_str = lept_stringify(&v, NULL);
    lept_free(&v);
    return json_str;
}
```

## 15.5 使用示例

下面是如何使用我们的结构体与标签支持的示例：

```c
/* 定义结构体 */
typedef struct {
    char* name;
    int age;
    int is_student;
    char* email;
} Person;

/* 初始化反射信息 */
LEPT_STRUCT_BEGIN(Person)
    LEPT_STRUCT_STRING(Person, name)
    LEPT_STRUCT_INT(Person, age)
    LEPT_STRUCT_BOOLEAN(Person, is_student)
    LEPT_STRUCT_STRING_TAG(Person, email, "email_address", 0)  /* 重命名字段 */
LEPT_STRUCT_END(Person)

/* 使用反射进行序列化 */
void example_serialize() {
    Person person = {"John", 30, 1, "john@example.com"};
    lept_value json;
    char* json_str;
    
    Person_reflect_init();  /* 初始化反射信息 */
    lept_init(&json);
    
    lept_serialize_struct(&person, &Person_reflect, &json);
    json_str = lept_stringify(&json, NULL);
    
    printf("JSON: %s\n", json_str);
    
    free(json_str);
    lept_free(&json);
    free(person.name);
    free(person.email);
}

/* 使用反射进行反序列化 */
void example_deserialize() {
    const char* json_str = "{\"name\":\"John\",\"age\":30,\"is_student\":true,\"email_address\":\"john@example.com\"}";
    lept_value json;
    Person person = {NULL, 0, 0, NULL};
    
    Person_reflect_init();  /* 初始化反射信息 */
    lept_init(&json);
    lept_parse(&json, json_str);
    
    lept_deserialize_struct(&person, &Person_reflect, &json);
    
    printf("Name: %s, Age: %d, Is Student: %s, Email: %s\n",
           person.name, person.age, person.is_student ? "Yes" : "No", person.email);
    
    lept_free(&json);
    free(person.name);
    free(person.email);
}
```

## 15.6 实际应用场景

结构体与标签支持在多种场景下都非常有用：

### 15.6.1 配置文件处理

使用 JSON 作为配置文件格式，通过结构体映射实现简洁的配置读写：

```c
typedef struct {
    int port;
    char* host;
    int max_connections;
    double timeout;
    struct {
        int enabled;
        char* log_file;
        int log_level;
    } logging;
} ServerConfig;

LEPT_STRUCT_BEGIN(ServerConfig)
    LEPT_STRUCT_INT_FULL(ServerConfig, port, "port", 0, 1, "8080")
    LEPT_STRUCT_STRING(ServerConfig, host)
    LEPT_STRUCT_INT_TAG(ServerConfig, max_connections, "max_conn", 0)
    LEPT_STRUCT_DOUBLE(ServerConfig, timeout)
    /* 嵌套结构体 */
    /* ... */
LEPT_STRUCT_END(ServerConfig)

void load_config() {
    ServerConfig config = {0};
    FILE* fp = fopen("config.json", "r");
    if (fp) {
        /* 读取文件内容 */
        /* ... */
        from_json(config, json_content);
        fclose(fp);
    }
    
    printf("Server will run on %s:%d\n", config.host, config.port);
    /* 使用配置 */
}
```

### 15.6.2 API 交互

与 Web API 交互时，自动映射 JSON 请求和响应：

```c
typedef struct {
    char* username;
    char* password;
} LoginRequest;

typedef struct {
    char* token;
    int expires_in;
    char* user_id;
} LoginResponse;

/* 初始化反射信息 */
/* ... */

void api_login(const char* username, const char* password) {
    LoginRequest req = {(char*)username, (char*)password};
    char* json_req = to_json(req);
    
    /* 发送 HTTP 请求 */
    /* ... */
    
    /* 接收响应 */
    const char* json_resp = /* ... */;
    
    LoginResponse resp = {0};
    from_json(resp, json_resp);
    
    printf("Login successful! Token: %s (expires in %d seconds)\n",
           resp.token, resp.expires_in);
    
    /* 清理资源 */
    free(json_req);
    free(resp.token);
    free(resp.user_id);
}
```

### 15.6.3 数据库映射

将数据库记录映射到结构体：

```c
typedef struct {
    int id;
    char* title;
    char* content;
    char* author;
    char* created_at;
} Article;

/* 初始化反射信息 */
/* ... */

void fetch_articles() {
    /* 执行数据库查询 */
    /* ... */
    
    /* 假设查询结果是一个JSON数组 */
    const char* json_results = /* ... */;
    
    lept_value v;
    lept_init(&v);
    lept_parse(&v, json_results);
    
    /* 假设结果是一个数组 */
    if (v.type == LEPT_ARRAY) {
        Article* articles = malloc(v.u.a.size * sizeof(Article));
        
        for (size_t i = 0; i < v.u.a.size; i++) {
            memset(&articles[i], 0, sizeof(Article));
            deserialize(articles[i], &v.u.a.e[i]);
            
            printf("Article #%d: %s by %s\n",
                   articles[i].id, articles[i].title, articles[i].author);
        }
        
        /* 清理资源 */
        for (size_t i = 0; i < v.u.a.size; i++) {
            free(articles[i].title);
            free(articles[i].content);
            free(articles[i].author);
            free(articles[i].created_at);
        }
        free(articles);
    }
    
    lept_free(&v);
}
```

## 15.7 自动生成反射代码

手动编写反射代码可能会很繁琐，我们可以设计一个工具，自动从结构体定义生成反射代码。这个工具可以解析C源代码，提取结构体信息，并生成相应的反射初始化代码。

此外，我们还可以支持通过注释来指定标签：

```c
typedef struct {
    char* name;              /* json:"name" */
    int age;                 /* json:"age" */
    int is_student;          /* json:"is_student" */
    char* email;             /* json:"email_address,omitempty" */
    double salary;           /* json:"-" */  /* 忽略此字段 */
} Person;
```

## 15.8 练习

1. 扩展反射机制，支持数组类型，包括定长数组和动态数组。
2. 为反射机制添加更多标签选项，如默认值、验证规则等。
3. 实现一个简单的代码生成工具，自动从结构体定义生成反射代码。
4. 为反射机制添加对循环引用的检测。
5. 扩展反射机制，支持从 JSON 数组反序列化为结构体数组。
6. 实现一个基于模板的 C++ 包装器，提供更友好的 API。

## 15.9 下一步

本章中，我们实现了结构体与 JSON 之间的自动转换机制，大大简化了数据处理工作。在下一章中，我们将探讨如何构建一个完整的 JSON 命令行工具，用于处理和操作 JSON 数据。 