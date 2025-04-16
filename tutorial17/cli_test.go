package leptjson

import (
	"os"
	"testing"
)

func TestCalculateStats(t *testing.T) {
	// 创建一个测试JSON值
	v := &Value{}
	SetObject(v)

	// 添加一个字符串字段
	strVal := SetObjectValue(v, "name")
	SetString(strVal, "测试数据")

	// 添加一个数字字段
	numVal := SetObjectValue(v, "value")
	SetNumber(numVal, 42.5)

	// 添加一个布尔字段
	boolVal := SetObjectValue(v, "active")
	SetBoolean(boolVal, true)

	// 添加一个null字段
	nullVal := SetObjectValue(v, "empty")
	SetNull(nullVal)

	// 添加一个数组字段
	arrVal := SetObjectValue(v, "items")
	SetArray(arrVal, 3)

	item1 := PushBackArrayElement(arrVal)
	SetNumber(item1, 1)

	item2 := PushBackArrayElement(arrVal)
	SetNumber(item2, 2)

	item3 := PushBackArrayElement(arrVal)
	SetNumber(item3, 3)

	// 添加一个嵌套对象
	nestedObj := SetObjectValue(v, "nested")
	SetObject(nestedObj)
	nestedField := SetObjectValue(nestedObj, "very_long_field_name")
	SetString(nestedField, "nested value")

	// 计算统计信息
	stats := calculateStats(v)

	// 验证统计信息
	if stats.ObjectCount != 2 {
		t.Errorf("对象计数错误: 期望 2, 实际 %d", stats.ObjectCount)
	}

	if stats.ArrayCount != 1 {
		t.Errorf("数组计数错误: 期望 1, 实际 %d", stats.ArrayCount)
	}

	if stats.StringCount != 2 {
		t.Errorf("字符串计数错误: 期望 2, 实际 %d", stats.StringCount)
	}

	if stats.NumberCount != 4 {
		t.Errorf("数字计数错误: 期望 4, 实际 %d", stats.NumberCount)
	}

	if stats.BooleanCount != 1 {
		t.Errorf("布尔值计数错误: 期望 1, 实际 %d", stats.BooleanCount)
	}

	if stats.NullCount != 1 {
		t.Errorf("null计数错误: 期望 1, 实际 %d", stats.NullCount)
	}

	if stats.MaxDepth != 2 {
		t.Errorf("最大深度错误: 期望 2, 实际 %d", stats.MaxDepth)
	}

	if stats.KeyCount != 7 {
		t.Errorf("键计数错误: 期望 7, 实际 %d", stats.KeyCount)
	}

	if stats.MaxKeyLength != 20 || stats.LongestKey != "very_long_field_name" {
		t.Errorf("最长键错误: 期望 'very_long_field_name' (20字符), 实际 '%s' (%d字符)",
			stats.LongestKey, stats.MaxKeyLength)
	}
}

func TestFormatJSON(t *testing.T) {
	// 创建一个测试JSON值
	v := &Value{}
	SetObject(v)

	// 添加一些字段
	nameVal := SetObjectValue(v, "name")
	SetString(nameVal, "测试")

	numVal := SetObjectValue(v, "value")
	SetNumber(numVal, 42)

	// 格式化JSON
	formatted, err := formatJSON(v, "  ")

	// 验证结果
	if err != nil {
		t.Errorf("格式化失败: %v", err)
	}

	expected := "{\n  \"name\": \"测试\",\n  \"value\": 42\n}"
	if formatted != expected {
		t.Errorf("格式化结果错误\n期望:\n%s\n\n实际:\n%s", expected, formatted)
	}
}

func TestCompareJSON(t *testing.T) {
	// 创建两个相似但有差异的JSON值
	v1 := &Value{}
	SetObject(v1)
	name1 := SetObjectValue(v1, "name")
	SetString(name1, "原始值")
	num1 := SetObjectValue(v1, "value")
	SetNumber(num1, 42)

	v2 := &Value{}
	SetObject(v2)
	name2 := SetObjectValue(v2, "name")
	SetString(name2, "新值")
	num2 := SetObjectValue(v2, "value")
	SetNumber(num2, 42)
	extra := SetObjectValue(v2, "extra")
	SetBoolean(extra, true)

	// 比较JSON
	differences := compareJSON(v1, v2)

	// 验证差异
	if len(differences) != 2 {
		t.Errorf("差异数量错误: 期望 2, 实际 %d", len(differences))
	}

	// 检查是否包含预期的差异
	expectedDiff1 := "路径 $.name: 字符串不同 (\"原始值\" vs \"新值\")"
	expectedDiff2 := "路径 $: 第二个JSON有键 'extra'，但第一个没有"

	found1, found2 := false, false
	for _, diff := range differences {
		if diff == expectedDiff1 {
			found1 = true
		}
		if diff == expectedDiff2 {
			found2 = true
		}
	}

	if !found1 {
		t.Errorf("未找到预期的差异: %s", expectedDiff1)
	}

	if !found2 {
		t.Errorf("未找到预期的差异: %s", expectedDiff2)
	}
}

// 为了避免在运行测试时实际执行命令行操作，我们不测试CLI命令的执行函数
// 相反，我们只测试核心功能函数

func TestMinifyJSON(t *testing.T) {
	// 创建一个测试JSON值
	v := &Value{}
	SetObject(v)

	// 添加一些字段
	nameVal := SetObjectValue(v, "name")
	SetString(nameVal, "测试")

	arrVal := SetObjectValue(v, "array")
	SetArray(arrVal, 2)

	item1 := PushBackArrayElement(arrVal)
	SetNumber(item1, 1)

	item2 := PushBackArrayElement(arrVal)
	SetNumber(item2, 2)

	// 最小化JSON
	minified, err := minifyJSON(v)

	// 验证结果
	if err != nil {
		t.Errorf("最小化失败: %v", err)
	}

	expected := "{\"name\":\"测试\",\"array\":[1,2]}"
	if minified != expected {
		t.Errorf("最小化结果错误\n期望: %s\n实际: %s", expected, minified)
	}
}

func TestLoadAndSaveJSON(t *testing.T) {
	// 创建临时文件
	tempFile, err := os.CreateTemp("", "test-*.json")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// 写入JSON内容
	jsonContent := "{\"test\":true}"
	_, err = tempFile.WriteString(jsonContent)
	tempFile.Close()
	if err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}

	// 测试加载
	v, err := loadJSON(tempFile.Name(), false)
	if err != nil {
		t.Fatalf("加载JSON失败: %v", err)
	}

	// 验证加载的内容
	if v.Type != OBJECT || len(v.O) != 1 {
		t.Errorf("加载的JSON内容错误")
	}

	// 获取test字段
	var testVal *Value
	for _, member := range v.O {
		if member.K == "test" {
			testVal = member.V
			break
		}
	}

	if testVal == nil || testVal.Type != TRUE {
		t.Errorf("test字段值错误")
	}

	// 测试保存
	outputFile := tempFile.Name() + ".out"
	defer os.Remove(outputFile)

	err = saveJSON(outputFile, jsonContent, false)
	if err != nil {
		t.Fatalf("保存JSON失败: %v", err)
	}

	// 读取保存的内容
	savedContent, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("读取保存的文件失败: %v", err)
	}

	if string(savedContent) != jsonContent {
		t.Errorf("保存的内容错误\n期望: %s\n实际: %s", jsonContent, string(savedContent))
	}
}

// 添加JSONPath测试
func TestJSONPath(t *testing.T) {
	// 创建一个测试JSON值
	v := &Value{}
	SetObject(v)

	// 创建store对象
	store := SetObjectValue(v, "store")
	SetObject(store)

	// 创建books数组
	books := SetObjectValue(store, "book")
	SetArray(books, 0)

	// 添加图书1
	book1 := PushBackArrayElement(books)
	SetObject(book1)
	category1 := SetObjectValue(book1, "category")
	SetString(category1, "reference")
	author1 := SetObjectValue(book1, "author")
	SetString(author1, "Nigel Rees")
	title1 := SetObjectValue(book1, "title")
	SetString(title1, "Sayings of the Century")
	price1 := SetObjectValue(book1, "price")
	SetNumber(price1, 8.95)

	// 添加图书2
	book2 := PushBackArrayElement(books)
	SetObject(book2)
	category2 := SetObjectValue(book2, "category")
	SetString(category2, "fiction")
	author2 := SetObjectValue(book2, "author")
	SetString(author2, "Evelyn Waugh")
	title2 := SetObjectValue(book2, "title")
	SetString(title2, "Sword of Honour")
	price2 := SetObjectValue(book2, "price")
	SetNumber(price2, 12.99)

	// 测试基本路径
	path, err := NewJSONPath("$.store.book[0].title")
	if err != nil {
		t.Errorf("创建JSONPath失败: %v", err)
	}

	results, err := path.Query(v)
	if err != nil {
		t.Errorf("执行查询失败: %v", err)
	}

	if len(results) != 1 || results[0].Type != STRING || results[0].S != "Sayings of the Century" {
		t.Errorf("基本路径查询结果错误")
	}

	// 测试通配符
	path, err = NewJSONPath("$.store.book[*].author")
	if err != nil {
		t.Errorf("创建JSONPath失败: %v", err)
	}

	results, err = path.Query(v)
	if err != nil {
		t.Errorf("执行查询失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("通配符查询结果数量错误: 期望 2, 实际 %d", len(results))
	}

	// 测试递归下降
	path, err = NewJSONPath("$..author")
	if err != nil {
		t.Errorf("创建JSONPath失败: %v", err)
	}

	results, err = path.Query(v)
	if err != nil {
		t.Errorf("执行查询失败: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("递归下降查询结果数量错误: 期望 2, 实际 %d", len(results))
	}
}

// 添加JSON Pointer测试
func TestJSONPointer(t *testing.T) {
	// 创建一个测试JSON值
	v := &Value{}
	SetObject(v)

	// 添加一些属性
	nameVal := SetObjectValue(v, "name")
	SetString(nameVal, "John")

	// 添加嵌套对象
	addressVal := SetObjectValue(v, "address")
	SetObject(addressVal)
	cityVal := SetObjectValue(addressVal, "city")
	SetString(cityVal, "New York")

	// 添加数组
	phonesVal := SetObjectValue(v, "phones")
	SetArray(phonesVal, 0)
	phone1 := PushBackArrayElement(phonesVal)
	SetString(phone1, "123456789")
	phone2 := PushBackArrayElement(phonesVal)
	SetString(phone2, "987654321")

	// 测试解析和获取值
	pointer, err := NewJSONPointer("/name")
	if err != nil {
		t.Errorf("创建JSON Pointer失败: %v", err)
	}

	value, _, err := ResolvePointer(v, pointer)
	if err != nil {
		t.Errorf("解析指针失败: %v", err)
	}

	if value.Type != STRING || value.S != "John" {
		t.Errorf("指针解析结果错误")
	}

	// 测试嵌套属性
	pointer, err = NewJSONPointer("/address/city")
	if err != nil {
		t.Errorf("创建JSON Pointer失败: %v", err)
	}

	value, _, err = ResolvePointer(v, pointer)
	if err != nil {
		t.Errorf("解析指针失败: %v", err)
	}

	if value.Type != STRING || value.S != "New York" {
		t.Errorf("指针解析结果错误")
	}

	// 测试数组索引
	pointer, err = NewJSONPointer("/phones/1")
	if err != nil {
		t.Errorf("创建JSON Pointer失败: %v", err)
	}

	value, _, err = ResolvePointer(v, pointer)
	if err != nil {
		t.Errorf("解析指针失败: %v", err)
	}

	if value.Type != STRING || value.S != "987654321" {
		t.Errorf("指针解析结果错误")
	}

	// 测试添加操作
	pointer, err = NewJSONPointer("/age")
	if err != nil {
		t.Errorf("创建JSON Pointer失败: %v", err)
	}

	newVal := &Value{}
	SetNumber(newVal, 30)

	err = PointerAdd(v, pointer, newVal)
	if err != nil {
		t.Errorf("添加操作失败: %v", err)
	}

	value, _, err = ResolvePointer(v, pointer)
	if err != nil {
		t.Errorf("解析指针失败: %v", err)
	}

	if value.Type != NUMBER || value.N != 30 {
		t.Errorf("添加操作结果错误")
	}

	// 测试移除操作
	err = PointerRemove(v, pointer)
	if err != nil {
		t.Errorf("移除操作失败: %v", err)
	}

	_, _, err = ResolvePointer(v, pointer)
	if err == nil {
		t.Errorf("移除操作未生效")
	}

	// 测试替换操作
	pointer, err = NewJSONPointer("/name")
	if err != nil {
		t.Errorf("创建JSON Pointer失败: %v", err)
	}
	replaceVal := &Value{}
	SetString(replaceVal, "Replaced John")
	err = PointerReplace(v, pointer, replaceVal)
	if err != nil {
		t.Errorf("替换操作失败: %v", err)
	}
	value, _, err = ResolvePointer(v, pointer)
	if err != nil {
		t.Errorf("替换后解析指针失败: %v", err)
	}
	if value.Type != STRING || value.S != "Replaced John" {
		t.Errorf("替换操作结果错误")
	}
}

// 添加JSON Patch测试
func TestJSONPatch(t *testing.T) {
	// 创建一个测试JSON文档
	doc := &Value{}
	SetObject(doc)

	// 添加一些初始属性
	nameVal := SetObjectValue(doc, "name")
	SetString(nameVal, "Original")

	numVal := SetObjectValue(doc, "value")
	SetNumber(numVal, 10)

	// 创建一个JSON Patch文档
	patchDoc := &Value{}
	SetArray(patchDoc, 0)

	// 添加操作: 替换name
	op1 := PushBackArrayElement(patchDoc)
	SetObject(op1)
	opType1 := SetObjectValue(op1, "op")
	SetString(opType1, "replace")
	path1 := SetObjectValue(op1, "path")
	SetString(path1, "/name")
	value1 := SetObjectValue(op1, "value")
	SetString(value1, "Updated")

	// 添加操作: 添加新字段
	op2 := PushBackArrayElement(patchDoc)
	SetObject(op2)
	opType2 := SetObjectValue(op2, "op")
	SetString(opType2, "add")
	path2 := SetObjectValue(op2, "path")
	SetString(path2, "/new")
	value2 := SetObjectValue(op2, "value")
	SetBoolean(value2, true)

	// 添加操作: 删除字段
	op3 := PushBackArrayElement(patchDoc)
	SetObject(op3)
	opType3 := SetObjectValue(op3, "op")
	SetString(opType3, "remove")
	path3 := SetObjectValue(op3, "path")
	SetString(path3, "/value")

	// 解析补丁
	operations, err := parsePatch(patchDoc)
	if err != nil {
		t.Errorf("解析JSON Patch失败: %v", err)
	}

	if len(operations) != 3 {
		t.Errorf("操作数量错误: 期望 3, 实际 %d", len(operations))
	}

	// 应用补丁
	err = applyPatch(doc, operations, false)
	if err != nil {
		t.Errorf("应用JSON Patch失败: %v", err)
	}

	// 验证结果
	// 1. name应该被替换
	namePointer, _ := NewJSONPointer("/name")
	nameValue, _, err := ResolvePointer(doc, namePointer)
	if err != nil || nameValue.Type != STRING || nameValue.S != "Updated" {
		t.Errorf("replace操作结果错误")
	}

	// 2. 应该添加了新字段
	newPointer, _ := NewJSONPointer("/new")
	newValue, _, err := ResolvePointer(doc, newPointer)
	if err != nil || newValue.Type != TRUE {
		t.Errorf("add操作结果错误")
	}

	// 3. value应该被移除
	valuePointer, _ := NewJSONPointer("/value")
	_, _, err = ResolvePointer(doc, valuePointer)
	if err == nil {
		t.Errorf("remove操作结果错误")
	}
}

// 添加JSON Merge Patch测试
func TestJSONMergePatch(t *testing.T) {
	// 创建一个测试JSON文档
	target := &Value{}
	SetObject(target)

	// 添加一些初始属性
	nameVal := SetObjectValue(target, "name")
	SetString(nameVal, "Original")

	numVal := SetObjectValue(target, "value")
	SetNumber(numVal, 10)

	objVal := SetObjectValue(target, "nested")
	SetObject(objVal)
	field1 := SetObjectValue(objVal, "field1")
	SetString(field1, "keep")
	field2 := SetObjectValue(objVal, "field2")
	SetString(field2, "delete")

	// 创建一个JSON Merge Patch
	patch := &Value{}
	SetObject(patch)

	// 替换name
	newName := SetObjectValue(patch, "name")
	SetString(newName, "Updated")

	// 删除value
	nullValue := SetObjectValue(patch, "value")
	SetNull(nullValue)

	// 修改nested对象
	newObj := SetObjectValue(patch, "nested")
	SetObject(newObj)
	// 保留field1（不设置）
	// 删除field2
	delField2 := SetObjectValue(newObj, "field2")
	SetNull(delField2)
	// 添加field3
	addField3 := SetObjectValue(newObj, "field3")
	SetString(addField3, "new")

	// 应用Merge Patch
	err := applyMergePatch(target, patch)
	if err != nil {
		t.Errorf("应用JSON Merge Patch失败: %v", err)
	}

	// 验证结果
	// 1. name应该被替换
	for _, m := range target.O {
		if m.K == "name" {
			if m.V.Type != STRING || m.V.S != "Updated" {
				t.Errorf("name字段更新错误")
			}
		}
	}

	// 2. value应该被删除
	found := false
	for _, m := range target.O {
		if m.K == "value" {
			found = true
			break
		}
	}
	if found {
		t.Errorf("value字段未被删除")
	}

	// 3. nested.field1应该保留
	// 4. nested.field2应该被删除
	// 5. nested.field3应该被添加
	var nested *Value
	for _, m := range target.O {
		if m.K == "nested" {
			nested = m.V
			break
		}
	}

	if nested == nil || nested.Type != OBJECT {
		t.Errorf("nested对象丢失或类型错误")
	} else {
		foundField1, foundField2, foundField3 := false, false, false
		field3Value := ""

		for _, m := range nested.O {
			switch m.K {
			case "field1":
				foundField1 = true
				if m.V.Type != STRING || m.V.S != "keep" {
					t.Errorf("field1值错误")
				}
			case "field2":
				foundField2 = true
			case "field3":
				foundField3 = true
				if m.V.Type == STRING {
					field3Value = m.V.S
				}
			}
		}

		if !foundField1 {
			t.Errorf("field1丢失")
		}
		if foundField2 {
			t.Errorf("field2未被删除")
		}
		if !foundField3 || field3Value != "new" {
			t.Errorf("field3未被添加或值错误")
		}
	}
}

// 测试JSON Schema验证核心逻辑
func TestValidateWithSchema(t *testing.T) {
	// 准备测试数据
	schemaJSON := `{
		"type": "object",
		"required": ["name", "age"],
		"properties": {
			"name": {"type": "string", "minLength": 2},
			"age": {"type": "integer", "minimum": 18}
		}
	}`
	validDataJSON := `{"name": "Alice", "age": 30}`
	invalidDataJSON := `{"name": "B", "age": 15}`

	// 解析Schema和数据
	schema := &Value{}
	if errCode := Parse(schema, schemaJSON); errCode != PARSE_OK {
		t.Fatalf("解析Schema失败: %s", errCode.Error())
	}

	validData := &Value{}
	if errCode := Parse(validData, validDataJSON); errCode != PARSE_OK {
		t.Fatalf("解析有效数据失败: %s", errCode.Error())
	}

	invalidData := &Value{}
	if errCode := Parse(invalidData, invalidDataJSON); errCode != PARSE_OK {
		t.Fatalf("解析无效数据失败: %s", errCode.Error())
	}

	// 验证有效数据
	resultValid := validateWithSchema(schema, validData)
	if !resultValid.Valid {
		t.Errorf("有效数据验证失败，期望通过，实际错误: %v", resultValid.Errors)
	}

	// 验证无效数据
	resultInvalid := validateWithSchema(schema, invalidData)
	if resultInvalid.Valid {
		t.Errorf("无效数据验证成功，期望失败")
	}
	if len(resultInvalid.Errors) != 2 {
		t.Errorf("无效数据验证错误数量错误，期望 2，实际 %d", len(resultInvalid.Errors))
	}
	// 可以进一步检查具体的错误信息
	expectedError1 := "位于'$.name'的字符串长度1小于最小长度2"
	expectedError2 := "位于'$.age'的数值15小于最小值18"
	foundError1, foundError2 := false, false
	for _, e := range resultInvalid.Errors {
		if e == expectedError1 {
			foundError1 = true
		}
		if e == expectedError2 {
			foundError2 = true
		}
	}
	if !foundError1 || !foundError2 {
		t.Errorf(`无效数据验证的错误信息不匹配。
期望包含: '%s' 和 '%s'
实际错误: %v`, expectedError1, expectedError2, resultInvalid.Errors)
	}
}

// 测试 findByPath 函数
func TestFindByPath(t *testing.T) {
	// 创建测试JSON
	doc := &Value{}
	jsonStr := `{
		"store": {
			"book": [
				{"title": "Title A", "price": 10},
				{"title": "Title B", "price": 20}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		}
	}`
	if errCode := Parse(doc, jsonStr); errCode != PARSE_OK {
		t.Fatalf("解析测试JSON失败: %s", errCode.Error())
	}

	tests := []struct {
		path     string
		expected *Value
		wantErr  bool
	}{
		{
			path:     "$.store.bicycle.color",
			expected: &Value{Type: STRING, S: "red"},
			wantErr:  false,
		},
		{
			path:     "$.store.book[1].title",
			expected: &Value{Type: STRING, S: "Title B"},
			wantErr:  false,
		},
		{
			path:     "$.store.book[0].price",
			expected: &Value{Type: NUMBER, N: 10},
			wantErr:  false,
		},
		{
			path:     "$.store.book[2]", // Index out of bounds
			expected: nil,
			wantErr:  true,
		},
		{
			path:     "$.store.nonexistent", // Key not found
			expected: nil,
			wantErr:  true,
		},
		{
			path:     "$.store.book.title", // Accessing property on array
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result, err := findByPath(doc, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !ValuesEqual(result, tt.expected) {
				resultStr, _ := Stringify(result)
				expectedStr, _ := Stringify(tt.expected)
				t.Errorf("findByPath() got = %s, want %s", resultStr, expectedStr)
			}
		})
	}
}
