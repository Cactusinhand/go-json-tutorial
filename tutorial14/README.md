# 从零开始的 JSON 库教程（十四）：JSON Merge Patch

## JSON Merge Patch 简介

JSON Merge Patch 是一种简单直观的 JSON 文档修改方法，定义在 [RFC 7396](https://tools.ietf.org/html/rfc7396) 中。与 JSON Patch（RFC 6902）相比，它提供了一种更简洁的方式来表示对 JSON 文档的更新操作，特别适用于 HTTP PATCH 请求。

JSON Merge Patch 的核心思想是使用一个类似目标文档结构的 JSON 对象来描述修改。修改规则如下：

1. 如果 Merge Patch 是一个对象，那么对于其中的每个成员：
   - 如果值为 `null`，从目标文档中删除该成员
   - 如果值是对象，递归应用上述规则
   - 其他情况（包括数组），直接替换目标文档中的对应值

2. 如果 Merge Patch 不是一个对象（例如是数组、字符串或 `null`），则整个目标文档被替换为该值

## 与 JSON Patch 的比较

| 特性 | JSON Merge Patch | JSON Patch |
|------|-----------------|------------|
| 格式 | 单个 JSON 对象 | JSON 数组，包含操作对象 |
| 操作类型 | 隐含的添加/替换/删除 | 明确的 add/remove/replace/move/copy/test |
| 数组处理 | 只能替换整个数组 | 可以修改数组中的特定元素 |
| 表达能力 | 较弱，不支持复杂操作 | 较强，支持复杂的精确操作 |
| 易用性 | 简单直观 | 相对复杂 |
| 适用场景 | 简单的文档更新 | 复杂的文档转换 |

## 使用示例

假设我们有以下 JSON 文档：

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

我们可以应用以下 JSON Merge Patch：

```json
{
  "title": "Hello!",
  "author": {
    "familyName": null
  },
  "tags": ["example", "changed"],
  "phoneNumber": "+01-123-456-7890"
}
```

应用后，JSON 文档将变为：

```json
{
  "title": "Hello!",
  "author": {
    "givenName": "John"
  },
  "tags": ["example", "changed"],
  "content": "This will be unchanged",
  "phoneNumber": "+01-123-456-7890"
}
```

注意：
- `title` 被替换为新值
- `author.familyName` 被删除（因为其值为 `null`）
- `tags` 数组被完全替换
- `phoneNumber` 被添加
- `content` 保持不变（因为 Patch 中未提及）

## 实现目标

在本章中，我们将实现一个符合 RFC 7396 的 JSON Merge Patch 处理器，支持以下功能：

1. 解析和验证 JSON Merge Patch 文档
2. 应用 Merge Patch 到 JSON 文档
3. 生成两个文档之间的差异作为 JSON Merge Patch

## 实现方法

我们将创建一个 `JSONMergePatch` 类型，它封装了 JSON Merge Patch 文档并提供以下方法：

1. `NewJSONMergePatch` - 创建一个新的 JSON Merge Patch 对象
2. `Apply` - 将 Merge Patch 应用到目标文档
3. `String` - 将 Merge Patch 转换为字符串
4. `CreateMergePatch` - 从源文档和目标文档创建 JSON Merge Patch

实现过程中需要注意以下几点：
- 递归处理嵌套对象
- 正确处理 `null` 值（用于删除属性）
- 适当的错误处理和类型检查

## JSON Merge Patch 的局限性

虽然 JSON Merge Patch 简单直观，但它也有一些限制：

1. 不能在对象中表示删除所有属性（如果 patch 是一个空对象 `{}`，它不会修改目标）
2. 不能在数组内进行部分更新，只能替换整个数组
3. 不支持移动或复制操作
4. 不能区分设置值为 `null` 和删除该值

这些限制使得 JSON Merge Patch 适合简单的更新场景，但对于复杂的文档转换，可能需要使用更强大的 JSON Patch。

## 在 RESTful API 中的应用

JSON Merge Patch 特别适合用于 RESTful API 的 PATCH 请求。HTTP PATCH 方法（RFC 5789）用于对资源进行部分更新，而 JSON Merge Patch 提供了一种简单的方式来表示这些更新。

当使用 JSON Merge Patch 时，HTTP 请求应该使用 `Content-Type: application/merge-patch+json` 头。

## 参考资料

- [RFC 7396: JSON Merge Patch](https://tools.ietf.org/html/rfc7396)
- [RFC 5789: PATCH Method for HTTP](https://tools.ietf.org/html/rfc5789)
- [JSON Merge Patch 和 JSON Patch 比较](https://erosb.github.io/post/json-patch-vs-merge-patch/)
- [使用 JSON Merge Patch 进行 HTTP API 更新](https://williamdurand.fr/2014/02/14/please-do-not-patch-like-an-idiot/) 