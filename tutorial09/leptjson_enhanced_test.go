package leptjson

import (
	"strings"
	"testing"
)

// TestEnhancedErrorInfo 测试增强的错误信息
func TestEnhancedErrorInfo(t *testing.T) {
	tests := []struct {
		name          string
		json          string
		expectedCode  ParseError
		expectedLine  int
		expectedCol   int
		errorContains string
	}{
		{
			name:          "EmptyInput",
			json:          "",
			expectedCode:  PARSE_EXPECT_VALUE,
			expectedLine:  1,
			expectedCol:   1,
			errorContains: "输入为空",
		},
		{
			name:          "InvalidValue",
			json:          "nullx",
			expectedCode:  PARSE_INVALID_VALUE,
			expectedLine:  1,
			expectedCol:   1,
			errorContains: "无效的null值",
		},
		{
			name:          "RootNotSingular",
			json:          "null x",
			expectedCode:  PARSE_ROOT_NOT_SINGULAR,
			expectedLine:  1,
			expectedCol:   6,
			errorContains: "根节点后存在额外内容",
		},
		{
			name:          "NumberTooBig",
			json:          "1e1000",
			expectedCode:  PARSE_NUMBER_TOO_BIG,
			expectedLine:  1,
			expectedCol:   1,
			errorContains: "数字太大",
		},
		{
			name:          "MissingQuotationMark",
			json:          "\"hello",
			expectedCode:  PARSE_MISS_QUOTATION_MARK,
			expectedLine:  1,
			expectedCol:   7,
			errorContains: "字符串未闭合",
		},
		{
			name:          "InvalidEscape",
			json:          "\"\\v\"",
			expectedCode:  PARSE_INVALID_STRING_ESCAPE,
			expectedLine:  1,
			expectedCol:   3,
			errorContains: "无效的转义字符",
		},
		{
			name:          "MultilineError",
			json:          "{\n\"name\": \"value\",\n\"age\": tru\n}",
			expectedCode:  PARSE_INVALID_VALUE,
			expectedLine:  3,
			expectedCol:   9,
			errorContains: "无效的true值",
		},
		{
			name:          "MissingColon",
			json:          "{\"name\" \"value\"}",
			expectedCode:  PARSE_MISS_COLON,
			expectedLine:  1,
			expectedCol:   9,
			errorContains: "缺少冒号",
		},
		{
			name:          "InvalidObject",
			json:          "{\"name\": value}",
			expectedCode:  PARSE_INVALID_VALUE,
			expectedLine:  1,
			expectedCol:   10,
			errorContains: "无效的值",
		},
		{
			name:          "ArrayMissingComma",
			json:          "[1 2]",
			expectedCode:  PARSE_MISS_COMMA_OR_SQUARE_BRACKET,
			expectedLine:  1,
			expectedCol:   4,
			errorContains: "缺少逗号或方括号",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := Value{}
			options := ParseOptions{
				RecoverFromErrors: false,
				AllowComments:     false,
				MaxDepth:          1000,
			}

			// 使用ParseWithOptions替代EnhancedParse
			err := ParseWithOptions(&v, test.json, options)

			if err != test.expectedCode {
				t.Errorf("期望错误码 %v，实际得到 %v", test.expectedCode, err)
				return
			}

			// 创建增强的错误对象以获取详细信息
			ctx := newContext(test.json, options)
			ctx.index = len(test.json) // 模拟解析完成
			enhancedErr := ctx.createError(err, "")

			if enhancedErr == nil {
				t.Fatalf("预期应该返回错误，但返回nil")
			}

			if enhancedErr.Line != test.expectedLine {
				t.Errorf("期望错误在第 %d 行，实际在第 %d 行", test.expectedLine, enhancedErr.Line)
			}

			if enhancedErr.Column != test.expectedCol {
				t.Errorf("期望错误在第 %d 列，实际在第 %d 列", test.expectedCol, enhancedErr.Column)
			}

			if !strings.Contains(enhancedErr.Error(), test.errorContains) {
				t.Errorf("期望错误信息包含 '%s'，实际错误信息：'%s'", test.errorContains, enhancedErr.Error())
			}

			// 测试错误上下文
			if enhancedErr.Context == "" {
				t.Errorf("错误上下文不应为空")
			}

			if enhancedErr.Pointer == "" {
				t.Errorf("错误指针不应为空")
			}

			t.Logf("错误信息: %s", enhancedErr.Error())
		})
	}
}

// 重命名为TestErrorRecoveryEnhanced避免名称冲突
func TestErrorRecoveryEnhanced(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		options    ParseOptions
		shouldPass bool
		resultType ValueType // 使用ValueType代替Type
		arrayLen   int
	}{
		{
			name:       "ArrayWithMissingValue",
			json:       "[1, , 3]",
			options:    ParseOptions{RecoverFromErrors: true},
			shouldPass: true,
			resultType: ARRAY,
			arrayLen:   2, // 应该只解析出[1, 3]
		},
		{
			name:       "ArrayWithInvalidValue",
			json:       "[1, tru, 3]",
			options:    ParseOptions{RecoverFromErrors: true},
			shouldPass: true,
			resultType: ARRAY,
			arrayLen:   2, // 应该只解析出[1, 3]
		},
		{
			name:       "ArrayWithTrailingComma",
			json:       "[1, 2, 3,]",
			options:    ParseOptions{AllowTrailing: true},
			shouldPass: true,
			resultType: ARRAY,
			arrayLen:   3,
		},
		{
			name:       "ObjectWithTrailingComma",
			json:       `{"name": "value", "age": 30,}`,
			options:    ParseOptions{AllowTrailing: true},
			shouldPass: true,
			resultType: OBJECT,
		},
		{
			name:       "ObjectWithInvalidValue",
			json:       `{"name": "value", "age": tru, "city": "Beijing"}`,
			options:    ParseOptions{RecoverFromErrors: true},
			shouldPass: true,
			resultType: OBJECT,
		},
		{
			name:       "DeepNestedExceedMaxDepth",
			json:       `{"a":{"b":{"c":{"d":{"e":{"f":{}}}}}}}`,
			options:    ParseOptions{MaxDepth: 3},
			shouldPass: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := Value{}
			err := ParseWithOptions(&v, test.json, test.options)

			if test.shouldPass {
				if err != PARSE_OK {
					t.Errorf("期望解析成功，但返回错误: %v", err)
				}

				if v.Type != test.resultType {
					t.Errorf("期望结果类型为 %v，实际为 %v", test.resultType, v.Type)
				}

				if test.resultType == ARRAY && len(v.A) != test.arrayLen {
					t.Errorf("期望数组长度为 %d，实际为 %d", test.arrayLen, len(v.A))
				}
			} else {
				if err == PARSE_OK {
					t.Errorf("期望解析失败，但成功解析")
				}
			}
		})
	}
}

// TestCommentSupport 测试注释支持
func TestCommentSupport(t *testing.T) {
	tests := []struct {
		name       string
		json       string
		shouldPass bool
	}{
		{
			name: "SingleLineComment",
			json: `{
				"name": "value", // 这是单行注释
				"age": 30
			}`,
			shouldPass: true,
		},
		{
			name: "MultiLineComment",
			json: `{
				"name": "value", /* 这是
				多行注释 */
				"age": 30
			}`,
			shouldPass: true,
		},
		{
			name: "CommentInArray",
			json: `[
				1, // 第一个元素
				2, /* 第二个元素 */
				3
			]`,
			shouldPass: true,
		},
		{
			name:       "UnterminatedComment",
			json:       `{"name": "value" /* 未结束的注释 `,
			shouldPass: false,
		},
	}

	options := ParseOptions{
		AllowComments: true,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v := Value{}
			err := ParseWithOptions(&v, test.json, options)

			if test.shouldPass {
				if err != PARSE_OK {
					t.Errorf("期望解析成功，但返回错误: %v", err)
				}
			} else {
				if err == PARSE_OK {
					t.Errorf("期望解析失败，但成功解析")
				}
			}
		})
	}
}
