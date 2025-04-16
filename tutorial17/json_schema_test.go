package leptjson

import (
	"strings"
	"testing"
)

// 测试 JSON Schema 创建
func TestNewJSONSchema(t *testing.T) {
	// 测试有效的模式
	validSchema := `{
		"type": "object",
		"properties": {
			"name": {"type": "string"},
			"age": {"type": "integer"}
		}
	}`

	schema, err := NewJSONSchema(validSchema)
	if err != nil {
		t.Errorf("创建有效的 JSON Schema 失败: %v", err)
	}
	if schema == nil || schema.Schema == nil || schema.Schema.Type != OBJECT {
		t.Error("创建的 Schema 结构不正确")
	}

	// 测试无效的模式（非对象）
	invalidSchema := `"not an object"`
	schema, err = NewJSONSchema(invalidSchema)
	if err == nil {
		t.Error("期望对非对象 schema 返回错误，但没有")
	}

	// 测试无效的 JSON
	invalidJSON := `{invalid json`
	schema, err = NewJSONSchema(invalidJSON)
	if err == nil {
		t.Error("期望对无效 JSON 返回错误，但没有")
	}
}

// 测试基本类型验证
func TestBasicTypeValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "字符串类型",
			schema:         `{"type": "string"}`,
			validData:      `"test"`,
			invalidData:    `123`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "数字类型",
			schema:         `{"type": "number"}`,
			validData:      `42.5`,
			invalidData:    `"not a number"`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "整数类型",
			schema:         `{"type": "integer"}`,
			validData:      `42`,
			invalidData:    `42.5`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "布尔类型",
			schema:         `{"type": "boolean"}`,
			validData:      `true`,
			invalidData:    `"true"`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "空类型",
			schema:         `{"type": "null"}`,
			validData:      `null`,
			invalidData:    `0`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "数组类型",
			schema:         `{"type": "array"}`,
			validData:      `[1, 2, 3]`,
			invalidData:    `{"not": "array"}`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "对象类型",
			schema:         `{"type": "object"}`,
			validData:      `{"key": "value"}`,
			invalidData:    `[1, 2, 3]`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "多类型",
			schema:         `{"type": ["string", "number"]}`,
			validData:      `"test"`,
			invalidData:    `true`,
			invalidMessage: "类型不匹配",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 测试数值验证
func TestNumberValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "最小值",
			schema:         `{"type": "number", "minimum": 5}`,
			validData:      `10`,
			invalidData:    `2`,
			invalidMessage: "小于最小值",
		},
		{
			name:           "独占最小值",
			schema:         `{"type": "number", "exclusiveMinimum": 5}`,
			validData:      `5.1`,
			invalidData:    `5`,
			invalidMessage: "应该大于独占最小值",
		},
		{
			name:           "最大值",
			schema:         `{"type": "number", "maximum": 100}`,
			validData:      `50`,
			invalidData:    `150`,
			invalidMessage: "大于最大值",
		},
		{
			name:           "独占最大值",
			schema:         `{"type": "number", "exclusiveMaximum": 100}`,
			validData:      `99.9`,
			invalidData:    `100`,
			invalidMessage: "应该小于独占最大值",
		},
		{
			name:           "倍数",
			schema:         `{"type": "number", "multipleOf": 5}`,
			validData:      `25`,
			invalidData:    `26`,
			invalidMessage: "不是 5 的倍数",
		},
		{
			name:           "组合约束",
			schema:         `{"type": "number", "minimum": 10, "maximum": 100, "multipleOf": 5}`,
			validData:      `55`,
			invalidData:    `57`,
			invalidMessage: "不是 5 的倍数",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 测试字符串验证
func TestStringValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "最小长度",
			schema:         `{"type": "string", "minLength": 3}`,
			validData:      `"test"`,
			invalidData:    `"ab"`,
			invalidMessage: "小于最小长度",
		},
		{
			name:           "最大长度",
			schema:         `{"type": "string", "maxLength": 5}`,
			validData:      `"test"`,
			invalidData:    `"too long"`,
			invalidMessage: "大于最大长度",
		},
		{
			name:           "模式",
			schema:         `{"type": "string", "pattern": "^[a-z]+$"}`,
			validData:      `"abc"`,
			invalidData:    `"ABC"`,
			invalidMessage: "不匹配模式",
		},
		{
			name:           "格式-电子邮件",
			schema:         `{"type": "string", "format": "email"}`,
			validData:      `"user@example.com"`,
			invalidData:    `"invalid-email"`,
			invalidMessage: "电子邮件格式",
		},
		{
			name:           "格式-日期时间",
			schema:         `{"type": "string", "format": "date-time"}`,
			validData:      `"2023-01-01T12:00:00Z"`,
			invalidData:    `"2023-01-01"`,
			invalidMessage: "日期时间格式",
		},
		{
			name:           "组合约束",
			schema:         `{"type": "string", "minLength": 5, "maxLength": 10, "pattern": "^[a-z]+$"}`,
			validData:      `"abcdef"`,
			invalidData:    `"Ab123"`,
			invalidMessage: "不匹配模式",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 测试数组验证
func TestArrayValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "最小元素数",
			schema:         `{"type": "array", "minItems": 2}`,
			validData:      `[1, 2, 3]`,
			invalidData:    `[1]`,
			invalidMessage: "数组长度小于最小长度",
		},
		{
			name:           "最大元素数",
			schema:         `{"type": "array", "maxItems": 3}`,
			validData:      `[1, 2, 3]`,
			invalidData:    `[1, 2, 3, 4]`,
			invalidMessage: "数组长度大于最大长度",
		},
		{
			name:           "元素唯一性",
			schema:         `{"type": "array", "uniqueItems": true}`,
			validData:      `[1, 2, 3]`,
			invalidData:    `[1, 2, 1]`,
			invalidMessage: "数组包含重复项",
		},
		{
			name:           "元素类型验证",
			schema:         `{"type": "array", "items": {"type": "string"}}`,
			validData:      `["a", "b", "c"]`,
			invalidData:    `["a", 2, "c"]`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "列表验证",
			schema:         `{"type": "array", "items": [{"type": "number"}, {"type": "string"}]}`,
			validData:      `[42, "text"]`,
			invalidData:    `["text", 42]`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "包含验证",
			schema:         `{"type": "array", "contains": {"type": "number", "minimum": 10}}`,
			validData:      `[1, 15, 3]`,
			invalidData:    `[1, 2, 3]`,
			invalidMessage: "没有元素满足",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 测试对象验证
func TestObjectValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "最小属性数",
			schema:         `{"type": "object", "minProperties": 2}`,
			validData:      `{"a": 1, "b": 2, "c": 3}`,
			invalidData:    `{"a": 1}`,
			invalidMessage: "小于最小数量",
		},
		{
			name:           "最大属性数",
			schema:         `{"type": "object", "maxProperties": 2}`,
			validData:      `{"a": 1, "b": 2}`,
			invalidData:    `{"a": 1, "b": 2, "c": 3}`,
			invalidMessage: "大于最大数量",
		},
		{
			name:           "必需属性",
			schema:         `{"type": "object", "required": ["name", "age"]}`,
			validData:      `{"name": "John", "age": 30}`,
			invalidData:    `{"name": "John"}`,
			invalidMessage: "缺少必需属性",
		},
		{
			name:           "属性类型验证",
			schema:         `{"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}}`,
			validData:      `{"name": "John", "age": 30}`,
			invalidData:    `{"name": "John", "age": "thirty"}`,
			invalidMessage: "类型不匹配",
		},
		{
			name:           "额外属性",
			schema:         `{"type": "object", "properties": {"name": {"type": "string"}}, "additionalProperties": false}`,
			validData:      `{"name": "John"}`,
			invalidData:    `{"name": "John", "age": 30}`,
			invalidMessage: "不允许存在",
		},
		{
			name:           "模式属性",
			schema:         `{"type": "object", "patternProperties": {"^S_": {"type": "string"}}, "additionalProperties": false}`,
			validData:      `{"S_name": "John", "S_code": "ABC"}`,
			invalidData:    `{"S_name": "John", "age": 30}`,
			invalidMessage: "不允许存在",
		},
		{
			name:           "依赖关系",
			schema:         `{"type": "object", "dependencies": {"credit_card": ["billing_address"]}}`,
			validData:      `{"name": "John", "credit_card": "1234-5678-9012-3456", "billing_address": "123 Main St"}`,
			invalidData:    `{"name": "John", "credit_card": "1234-5678-9012-3456"}`,
			invalidMessage: "依赖于属性",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 测试逻辑组合验证
func TestLogicalValidation(t *testing.T) {
	cases := []struct {
		name           string
		schema         string
		validData      string
		invalidData    string
		invalidMessage string
	}{
		{
			name:           "allOf",
			schema:         `{"allOf": [{"type": "object"}, {"required": ["name"]}, {"required": ["age"]}]}`,
			validData:      `{"name": "John", "age": 30}`,
			invalidData:    `{"name": "John"}`,
			invalidMessage: "缺少必需属性",
		},
		{
			name:           "anyOf",
			schema:         `{"anyOf": [{"type": "string"}, {"type": "number"}]}`,
			validData:      `"test"`,
			invalidData:    `true`,
			invalidMessage: "不符合 anyOf",
		},
		{
			name:           "oneOf",
			schema:         `{"oneOf": [{"type": "number", "multipleOf": 5}, {"type": "number", "multipleOf": 3}]}`,
			validData:      `10`,
			invalidData:    `15`,
			invalidMessage: "匹配了 2 个 oneOf 模式",
		},
		{
			name:           "not",
			schema:         `{"not": {"type": "string"}}`,
			validData:      `42`,
			invalidData:    `"test"`,
			invalidMessage: "不应该匹配 not 模式",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建 schema
			schema, err := NewJSONSchema(tc.schema)
			if err != nil {
				t.Fatalf("创建 schema 失败: %v", err)
			}

			// 解析有效数据
			validValue := &Value{}
			if err := Parse(validValue, tc.validData); err != PARSE_OK {
				t.Fatalf("解析有效数据失败: %v", err)
			}

			// 验证有效数据
			result := schema.Validate(validValue)
			if !result.Valid {
				t.Errorf("有效数据验证失败: %v", result.Errors)
			}

			// 解析无效数据
			invalidValue := &Value{}
			if err := Parse(invalidValue, tc.invalidData); err != PARSE_OK {
				t.Fatalf("解析无效数据失败: %v", err)
			}

			// 验证无效数据
			result = schema.Validate(invalidValue)
			if result.Valid {
				t.Error("无效数据通过验证")
			} else {
				foundExpectedError := false
				for _, err := range result.Errors {
					if strings.Contains(err.Message, tc.invalidMessage) {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("验证错误不包含预期消息 '%s'，实际错误: %v", tc.invalidMessage, result.Errors)
				}
			}
		})
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && len(s) >= len(substr) && s[0:len(substr)] == substr || s[len(s)-len(substr):] == substr
}
