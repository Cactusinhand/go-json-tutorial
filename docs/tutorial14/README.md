# 第14章：JSON Merge Patch

JSON Merge Patch 是一种简化的 JSON 文档更新方法，相比 JSON Patch 更加直观。本章我们将学习 JSON Merge Patch 的规范，并实现一个 Go 语言版本的 JSON Merge Patch 处理器。

## 14.1 JSON Merge Patch 简介

JSON Merge Patch 由 [RFC 7396](https://tools.ietf.org/html/rfc7396) 定义，提供了一种更简单的方式来更新 JSON 文档。与 JSON Patch 不同，它不使用操作数组，而是使用一个与目标文档结构相似的 JSON 对象来描述更改。

JSON Merge Patch 的基本规则是：
- 如果 Merge Patch 中的值是 `null`，则从目标文档中删除该成员
- 如果 Merge Patch 中的值是对象，则递归应用合并
- 其他情况下，用 Merge Patch 中的值替换目标文档中的值

例如，对于以下目标文档：

```json
{
  "title": "Goodbye!",
  "author": {
    "givenName": "John",
    "familyName": "Doe"
  },
  "tags": ["example", "sample"],
  "content": "This will be unchanged"
}
```

应用此 Merge Patch：

```json
{
  "title": "Hello!",
  "author": {
    "familyName": null
  },
  "tags": ["example"]
}
```

结果将是：

```json
{
  "title": "Hello!",
  "author": {
    "givenName": "John"
  },
  "tags": ["example"],
  "content": "This will be unchanged"
}
```

## 14.2 JSON Merge Patch vs. JSON Patch

JSON Merge Patch 相比 JSON Patch 有以下特点：

| 特性 | JSON Merge Patch | JSON Patch |
|------|-----------------|------------|
| 语法 | 简单，与JSON文档结构相似 | 更复杂，使用操作数组 |
| 表达能力 | 有限，无法表达数组操作 | 强大，可以精确控制修改 |
| 学习曲线 | 浅，易于理解和使用 | 稍陡，需要学习多种操作类型 |
| 数组处理 | 只能替换整个数组 | 可以精确修改数组元素 |
| 适用场景 | 简单的文档更新 | 复杂的、精确的文档修改 |

## 14.3 实现 JSON Merge Patch

我们将实现一个 Go 语言版本的 JSON Merge Patch 处理器。以下是基本结构和错误类型：

```go
// MergePatchError 表示JSON Merge Patch过程中的错误类型
type MergePatchError int

const (
    MergePatchOK MergePatchError = iota
    MergePatchInvalidParameter
    MergePatchNotObject
    MergePatchMemoryError
)

// 错误信息
func (e MergePatchError) Error() string {
    switch e {
    case MergePatchOK:
        return "no error"
    case MergePatchInvalidParameter:
        return "invalid parameter"
    case MergePatchNotObject:
        return "document is not an object"
    case MergePatchMemoryError:
        return "memory allocation error"
    default:
        return "unknown error"
    }
}
```

### 14.3.1 实现合并算法

现在实现核心算法，即应用 Merge Patch 到目标文档：

```go
// ApplyMergePatch 将JSON Merge Patch应用到目标文档
func ApplyMergePatch(target, patch *Value) error {
    // 如果patch是null，直接将目标设置为null
    if patch.Type == Null {
        target.SetNull()
        return nil
    }
    
    // 如果patch不是对象，直接用patch替换target
    if patch.Type != Object {
        if err := target.Copy(patch); err != nil {
            return MergePatchMemoryError
        }
        return nil
    }
    
    // 如果target不是对象，将其初始化为空对象
    if target.Type != Object {
        target.SetObject()
    }
    
    // 遍历patch对象的所有成员
    for i := 0; i < patch.ObjectSize(); i++ {
        key := patch.GetObjectKey(i)
        patchValue := patch.GetObjectValue(i)
        
        // 如果patch成员的值是null，从target中删除此成员
        if patchValue.Type == Null {
            RemoveObjectMember(target, key)
            continue
        }
        
        // 获取目标对象中的成员
        targetValue := target.FindMember(key)
        
        // 如果目标中不存在此成员，或patch成员不是对象，直接设置/替换
        if targetValue == nil || patchValue.Type != Object {
            newValue := new(Value)
            newValue.Copy(patchValue)
            target.SetObjectMember(key, newValue)
        } else {
            // 递归应用合并
            if err := ApplyMergePatch(targetValue, patchValue); err != nil {
                return err
            }
        }
    }
    
    return nil
}

// RemoveObjectMember 从对象中移除指定成员
func RemoveObjectMember(obj *Value, key string) {
    if obj.Type != Object {
        return
    }
    
    // 在Go中，我们可以直接从map中删除元素
    delete(obj.ObjectMembers, key)
}
```

### 14.3.2 使用示例

以下是一个使用我们实现的 JSON Merge Patch 功能的示例：

```go
func ExampleJSONMergePatch() {
    // 创建目标文档
    target := new(Value)
    if err := Parse(target, `{
        "title": "Goodbye!",
        "author": {
            "givenName": "John",
            "familyName": "Doe"
        },
        "tags": ["example", "sample"],
        "content": "This will be unchanged"
    }`); err != nil {
        fmt.Println("解析目标文档失败:", err)
        return
    }
    
    // 创建Patch文档
    patch := new(Value)
    if err := Parse(patch, `{
        "title": "Hello!",
        "author": {
            "familyName": null
        },
        "tags": ["example"]
    }`); err != nil {
        fmt.Println("解析Patch文档失败:", err)
        return
    }
    
    // 应用Patch
    if err := ApplyMergePatch(target, patch); err != nil {
        fmt.Println("应用Patch失败:", err)
        return
    }
    
    // 输出结果
    result, err := Stringify(target)
    if err != nil {
        fmt.Println("序列化结果失败:", err)
        return
    }
    
    fmt.Println("合并结果:", result)
    // 输出: {"title":"Hello!","author":{"givenName":"John"},"tags":["example"],"content":"This will be unchanged"}
}
```

## 14.4 生成 JSON Merge Patch

我们也可以实现一个函数，用于从两个 JSON 文档生成 Merge Patch：

```go
// GenerateMergePatch 创建从source到target的JSON Merge Patch
func GenerateMergePatch(source, target *Value) (*Value, error) {
    // 如果目标是null或目标类型与源不同，直接返回目标
    if target.Type == Null || source.Type != target.Type {
        result := new(Value)
        if err := result.Copy(target); err != nil {
            return nil, MergePatchMemoryError
        }
        return result, nil
    }
    
    // 如果两者都不是对象，且值相同，返回null（表示无变化）
    if source.Type != Object {
        if source.Equals(target) {
            result := new(Value)
            result.SetNull()
            return result, nil
        }
        result := new(Value)
        if err := result.Copy(target); err != nil {
            return nil, MergePatchMemoryError
        }
        return result, nil
    }
    
    // 如果都是对象，创建一个空的Patch对象
    patch := new(Value)
    patch.SetObject()
    
    // 处理在source中存在但在target中不存在或已更改的成员
    for i := 0; i < source.ObjectSize(); i++ {
        key := source.GetObjectKey(i)
        sourceValue := source.GetObjectValue(i)
        targetValue := target.FindMember(key)
        
        // 如果成员在target中不存在，在patch中设置为null
        if targetValue == nil {
            nullValue := new(Value)
            nullValue.SetNull()
            patch.SetObjectMember(key, nullValue)
            continue
        }
        
        // 对子对象递归生成patch
        subPatch, err := GenerateMergePatch(sourceValue, targetValue)
        if err != nil {
            return nil, err
        }
        
        // 如果子patch不是null（表示有变化），添加到结果中
        if subPatch.Type != Null {
            patch.SetObjectMember(key, subPatch)
        }
    }
    
    // 处理在target中存在但在source中不存在的成员
    for i := 0; i < target.ObjectSize(); i++ {
        key := target.GetObjectKey(i)
        if source.FindMember(key) == nil {
            targetValue := target.GetObjectValue(i)
            newValue := new(Value)
            if err := newValue.Copy(targetValue); err != nil {
                return nil, MergePatchMemoryError
            }
            patch.SetObjectMember(key, newValue)
        }
    }
    
    // 如果patch为空对象（无变化），返回null
    if patch.ObjectSize() == 0 {
        patch.SetNull()
    }
    
    return patch, nil
}
```

## 14.5 HTTP 和 JSON Merge Patch

JSON Merge Patch 被定义为 HTTP PATCH 方法的标准媒体类型 `application/merge-patch+json`。下面是一个处理 HTTP PATCH 请求的 Go 函数示例：

```go
func handleMergePatchRequest(w http.ResponseWriter, r *http.Request) {
    // 只处理PATCH请求
    if r.Method != http.MethodPatch {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // 检查Content-Type
    contentType := r.Header.Get("Content-Type")
    if contentType != "application/merge-patch+json" {
        http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
        return
    }
    
    // 读取请求体
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }
    
    // 假设我们有一个存储的文档
    storedDoc := getStoredDocument() // 实际中应从数据库或存储系统获取
    
    // 解析patch文档
    patch := new(Value)
    if err := Parse(patch, string(body)); err != nil {
        http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    // 应用patch
    if err := ApplyMergePatch(storedDoc, patch); err != nil {
        http.Error(w, "Error applying patch: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 保存更新后的文档
    if err := saveDocument(storedDoc); err != nil {
        http.Error(w, "Error saving document", http.StatusInternalServerError)
        return
    }
    
    // 返回更新后的文档
    result, err := Stringify(storedDoc)
    if err != nil {
        http.Error(w, "Error converting to JSON", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(result))
}
```

## 14.6 练习

1. 实现 `ApplyMergePatch` 函数，并编写单元测试来验证其正确性。

2. 实现 `GenerateMergePatch` 函数，用于从两个 JSON 文档生成 Merge Patch。

3. 创建一个简单的 HTTP 服务器，处理 JSON Merge Patch 请求。

4. 扩展 JSON Merge Patch 实现，添加详细的错误报告。

5. 比较 JSON Merge Patch 和 JSON Patch 的性能差异。

## 14.7 下一步

本章我们学习了 JSON Merge Patch，这是一种简单直观的 JSON 文档更新方法。虽然它不如 JSON Patch 灵活，但在许多简单场景下使用更加方便。

在下一章中，我们将探讨如何实现 Go 结构体与 JSON 的转换，这将使我们的 JSON 库与 Go 的类型系统无缝集成。 