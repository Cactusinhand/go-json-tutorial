# 从零开始的 JSON 库教程（十三）：JSON Patch 实现

## JSON Patch 简介

JSON Patch 是一种用于描述 JSON 文档修改的格式，定义在 [RFC 6902](https://tools.ietf.org/html/rfc6902) 中。它提供了一种标准化的方式来表示对 JSON 文档的操作，如添加、删除、替换、移动或复制值，以及测试值是否存在或匹配预期。

JSON Patch 文档是一个 JSON 数组，每个元素都是一个表示单个操作的对象。每个操作对象至少有两个成员：
- `op`: 操作类型，如 "add"、"remove"、"replace"、"move"、"copy" 或 "test"
- `path`: 一个 JSON Pointer，指定要操作的目标位置

根据操作类型的不同，还可能需要其他字段，如：
- `value`: 用于 "add"、"replace" 或 "test" 操作
- `from`: 用于 "move" 或 "copy" 操作，指定源位置

## 操作类型

### 1. add
将值添加到对象或数组。对于数组，索引可以是"-"表示在数组末尾添加。

```json
{ "op": "add", "path": "/a/b/c", "value": [ "foo", "bar" ] }
```

### 2. remove
从对象或数组中移除值。

```json
{ "op": "remove", "path": "/a/b/c" }
```

### 3. replace
替换值。

```json
{ "op": "replace", "path": "/a/b/c", "value": 42 }
```

### 4. move
将值从一个位置移动到另一个位置。

```json
{ "op": "move", "from": "/a/b/c", "path": "/a/b/d" }
```

### 5. copy
从一个位置复制值到另一个位置。

```json
{ "op": "copy", "from": "/a/b/c", "path": "/a/b/e" }
```

### 6. test
测试值是否等于提供的值。

```json
{ "op": "test", "path": "/a/b/c", "value": "foo" }
```

## 使用示例

假设我们有一个 JSON 文档：

```json
{
  "biscuits": [
    { "name": "Digestive" },
    { "name": "Choco Leibniz" }
  ]
}
```

我们可以应用以下 JSON Patch：

```json
[
  { "op": "add", "path": "/biscuits/1", "value": { "name": "Ginger Nut" } },
  { "op": "remove", "path": "/biscuits/0" },
  { "op": "replace", "path": "/biscuits/0/name", "value": "Chocolate Digestive" },
  { "op": "copy", "from": "/biscuits/0", "path": "/best_biscuit" }
]
```

应用后，JSON 文档将变为：

```json
{
  "biscuits": [
    { "name": "Chocolate Digestive" }
  ],
  "best_biscuit": { "name": "Chocolate Digestive" }
}
```

## 本章实现目标

在本章中，我们将实现一个符合 RFC 6902 的 JSON Patch 处理器，支持以下功能：

1. 解析 JSON Patch 文档
2. 验证 Patch 操作的有效性
3. 应用 Patch 到 JSON 文档
4. 生成两个文档之间的差异作为 JSON Patch

此外，我们还将在下一章实现 JSON Merge Patch（RFC 7396），它是一种更简单但功能较弱的 JSON 文档修改方法。

## 实现计划

1. 创建基本的 JSON Patch 数据结构
2. 实现解析 JSON Patch 文档的功能
3. 实现各种操作类型的处理逻辑
4. 添加错误处理和验证
5. 实现生成 Patch 的功能
6. 完善文档和示例

## JSON Patch vs JSON Merge Patch

JSON Patch 和 JSON Merge Patch 是两种不同的 JSON 文档修改方法：

- **JSON Patch**：更详细和精确，支持多种操作类型，适合复杂的修改。
- **JSON Merge Patch**：更简单直观，但功能较弱，主要用于简单的更新操作。

在本章我们将实现 JSON Patch，在下一章将实现 JSON Merge Patch，以提供完整的 JSON 修改能力。

## 参考资料

- [RFC 6902: JavaScript Object Notation (JSON) Patch](https://tools.ietf.org/html/rfc6902)
- [RFC 6901: JavaScript Object Notation (JSON) Pointer](https://tools.ietf.org/html/rfc6901)
- [JSON Patch 官方网站](http://jsonpatch.com/)
- [JSON Patch 格式](https://jsonpatch.com/)
- [JSON Merge Patch](https://tools.ietf.org/html/rfc7396) 