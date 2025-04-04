package leptjson

import (
	"strings"
	"testing"
)

// 测试创建一个有效的 JSON Patch
func TestNewJSONPatch(t *testing.T) {
	testCases := []struct {
		name        string
		patchJSON   string
		shouldErr   bool
		errContains string
		opCount     int
	}{
		{
			name:      "Empty Patch",
			patchJSON: "[]",
			shouldErr: false,
			opCount:   0,
		},
		{
			name:      "Add Operation",
			patchJSON: `[{"op":"add","path":"/foo","value":"bar"}]`,
			shouldErr: false,
			opCount:   1,
		},
		{
			name:      "Multiple Operations",
			patchJSON: `[{"op":"add","path":"/foo","value":"bar"},{"op":"remove","path":"/baz"}]`,
			shouldErr: false,
			opCount:   2,
		},
		{
			name:        "Not an Array",
			patchJSON:   `{"op":"add","path":"/foo","value":"bar"}`,
			shouldErr:   true,
			errContains: "JSON Patch 必须是一个数组",
		},
		{
			name:        "Not an Object",
			patchJSON:   `["not an object"]`,
			shouldErr:   true,
			errContains: "JSON Patch 操作必须是一个对象",
		},
		{
			name:        "Missing op",
			patchJSON:   `[{"path":"/foo","value":"bar"}]`,
			shouldErr:   true,
			errContains: "缺少有效的 'op' 字段",
		},
		{
			name:        "Invalid op",
			patchJSON:   `[{"op":"invalid","path":"/foo","value":"bar"}]`,
			shouldErr:   true,
			errContains: "无效的操作类型",
		},
		{
			name:        "Missing path",
			patchJSON:   `[{"op":"add","value":"bar"}]`,
			shouldErr:   true,
			errContains: "缺少有效的 'path' 字段",
		},
		{
			name:        "Move without from",
			patchJSON:   `[{"op":"move","path":"/foo"}]`,
			shouldErr:   true,
			errContains: "缺少有效的 'from' 字段",
		},
		{
			name:        "Add without value",
			patchJSON:   `[{"op":"add","path":"/foo"}]`,
			shouldErr:   true,
			errContains: "缺少 'value' 字段",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patch, err := NewJSONPatchFromString(tc.patchJSON)

			if tc.shouldErr {
				if err == nil {
					t.Fatalf("期望错误，但没有得到错误")
				}
				if tc.errContains != "" && err.Error() != "" {
					if err.Error() == "" || !containsString(err.Error(), tc.errContains) {
						t.Fatalf("期望错误消息包含 %q，但得到 %q", tc.errContains, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Fatalf("期望成功，但得到错误: %v", err)
				}
				if patch == nil {
					t.Fatalf("期望非 nil 的补丁，但得到 nil")
				}
				if len(patch.Operations) != tc.opCount {
					t.Fatalf("期望 %d 个操作，但得到 %d 个", tc.opCount, len(patch.Operations))
				}
			}
		})
	}
}

// 测试应用 JSON Patch 操作
func TestApplyPatch(t *testing.T) {
	testCases := []struct {
		name        string
		document    string
		patch       string
		expected    string
		shouldErr   bool
		errContains string
	}{
		{
			name:     "Add to Object",
			document: `{"foo":"bar"}`,
			patch:    `[{"op":"add","path":"/baz","value":"qux"}]`,
			expected: `{"foo":"bar","baz":"qux"}`,
		},
		{
			name:     "Add to Array",
			document: `["foo","bar"]`,
			patch:    `[{"op":"add","path":"/1","value":"qux"}]`,
			expected: `["foo","qux","bar"]`,
		},
		{
			name:     "Remove from Object",
			document: `{"foo":"bar","baz":"qux"}`,
			patch:    `[{"op":"remove","path":"/baz"}]`,
			expected: `{"foo":"bar"}`,
		},
		{
			name:     "Remove from Array",
			document: `["foo","bar","baz"]`,
			patch:    `[{"op":"remove","path":"/1"}]`,
			expected: `["foo","baz"]`,
		},
		{
			name:     "Replace in Object",
			document: `{"foo":"bar","baz":"qux"}`,
			patch:    `[{"op":"replace","path":"/baz","value":"quux"}]`,
			expected: `{"foo":"bar","baz":"quux"}`,
		},
		{
			name:     "Move in Object",
			document: `{"foo":{"bar":"baz"},"qux":{"corge":"grault"}}`,
			patch:    `[{"op":"move","from":"/foo/bar","path":"/qux/thud"}]`,
			expected: `{"foo":{},"qux":{"corge":"grault","thud":"baz"}}`,
		},
		{
			name:     "Copy in Object",
			document: `{"foo":{"bar":"baz"},"qux":{"corge":"grault"}}`,
			patch:    `[{"op":"copy","from":"/foo/bar","path":"/qux/thud"}]`,
			expected: `{"foo":{"bar":"baz"},"qux":{"corge":"grault","thud":"baz"}}`,
		},
		{
			name:     "Test Success",
			document: `{"foo":"bar"}`,
			patch:    `[{"op":"test","path":"/foo","value":"bar"}]`,
			expected: `{"foo":"bar"}`,
		},
		{
			name:        "Test Failure",
			document:    `{"foo":"bar"}`,
			patch:       `[{"op":"test","path":"/foo","value":"qux"}]`,
			shouldErr:   true,
			errContains: "测试失败：目标值 \"bar\" 不等于预期值 \"qux\"",
		},
		{
			name:        "Invalid Path in Add",
			document:    `{"foo":"bar"}`,
			patch:       `[{"op":"add","path":"/baz/bat","value":"qux"}]`,
			shouldErr:   true,
			errContains: "添加操作失败: 对象中未找到指定的键",
		},
		{
			name:        "Invalid Path in Remove",
			document:    `{"foo":"bar"}`,
			patch:       `[{"op":"remove","path":"/baz"}]`,
			shouldErr:   true,
			errContains: "移除操作失败: 对象中未找到指定的键",
		},
		{
			name:        "Move to own Child",
			document:    `{"foo":{"bar":{"baz":"qux"}}}`,
			patch:       `[{"op":"move","from":"/foo","path":"/foo/bar/bam"}]`,
			shouldErr:   true,
			errContains: "添加移动值到目标失败: 对象中未找到指定的键",
		},
		{
			name:     "Complex Operations",
			document: `{"foo":["bar","baz"],"qux":{"corge":"grault"}}`,
			patch:    `[{"op":"remove","path":"/foo/0"},{"op":"add","path":"/foo/-","value":"quux"},{"op":"replace","path":"/qux/corge","value":"garply"}]`,
			expected: `{"foo":["baz","quux"],"qux":{"corge":"garply"}}`,
		},
		{
			name:     "Add to End of Array",
			document: `["foo","bar"]`,
			patch:    `[{"op":"add","path":"/-","value":"baz"}]`,
			expected: `["foo","bar","baz"]`,
		},
		{
			name:     "Add Nested Object",
			document: `{"foo":"bar"}`,
			patch:    `[{"op":"add","path":"/baz","value":{"qux":"quux"}}]`,
			expected: `{"foo":"bar","baz":{"qux":"quux"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析原始文档
			doc := &Value{}
			if err := Parse(doc, tc.document); err != PARSE_OK {
				t.Fatalf("解析原始文档失败: %s", GetErrorMessage(err))
			}

			// 创建并应用补丁
			patch, err := NewJSONPatchFromString(tc.patch)
			if err != nil {
				t.Fatalf("创建 JSON Patch 失败: %v", err)
			}

			err = patch.Apply(doc)

			if tc.shouldErr {
				if err == nil {
					t.Fatalf("期望错误，但没有得到错误")
				}
				if tc.errContains != "" {
					if err.Error() == "" || !containsString(err.Error(), tc.errContains) {
						t.Fatalf("期望错误消息包含 %q，但得到 %q", tc.errContains, err.Error())
					}
				}
			} else {
				if err != nil {
					t.Fatalf("期望成功，但得到错误: %v", err)
				}

				// 解析期望的结果
				expected := &Value{}
				if err := Parse(expected, tc.expected); err != PARSE_OK {
					t.Fatalf("解析期望结果失败: %s", GetErrorMessage(err))
				}

				// 检查结果是否正确
				if !Equal(doc, expected) {
					docStr, _ := Stringify(doc)
					expectedStr, _ := Stringify(expected)
					t.Fatalf("结果不匹配\n期望: %s\n实际: %s", expectedStr, docStr)
				}
			}
		})
	}
}

// 测试 JSON Patch 转字符串
func TestJSONPatchString(t *testing.T) {
	testCases := []struct {
		name     string
		patch    string
		expected string
	}{
		{
			name:     "Simple Patch",
			patch:    `[{"op":"add","path":"/foo","value":"bar"}]`,
			expected: `[{"op":"add","path":"/foo","value":"bar"}]`,
		},
		{
			name:     "Multiple Operations",
			patch:    `[{"op":"add","path":"/foo","value":"bar"},{"op":"remove","path":"/baz"}]`,
			expected: `[{"op":"add","path":"/foo","value":"bar"},{"op":"remove","path":"/baz"}]`,
		},
		{
			name:     "Copy Operation",
			patch:    `[{"op":"copy","from":"/foo/bar","path":"/qux/thud"}]`,
			expected: `[{"op":"copy","from":"/foo/bar","path":"/qux/thud"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patch, err := NewJSONPatchFromString(tc.patch)
			if err != nil {
				t.Fatalf("创建 JSON Patch 失败: %v", err)
			}

			patchStr, err := patch.String()
			if err != nil {
				t.Fatalf("转换 JSON Patch 为字符串失败: %v", err)
			}

			// 将结果解析为 JSON 值进行比较，避免格式差异
			expected := &Value{}
			if err := Parse(expected, tc.expected); err != PARSE_OK {
				t.Fatalf("解析期望结果失败: %s", GetErrorMessage(err))
			}

			result := &Value{}
			if err := Parse(result, patchStr); err != PARSE_OK {
				t.Fatalf("解析结果为 JSON 失败: %s", GetErrorMessage(err))
			}

			if !Equal(result, expected) {
				t.Fatalf("结果不匹配\n期望: %s\n实际: %s", tc.expected, patchStr)
			}
		})
	}
}

// 辅助函数：检查字符串是否包含子字符串
func containsString(s, substr string) bool {
	return s != "" && substr != "" && s != substr && strings.Contains(s, substr)
}

// 测试生成 JSON Patch
func TestCreatePatch(t *testing.T) {
	testCases := []struct {
		name     string
		source   string
		target   string
		expected string // 预期的 Patch 可能不唯一，这里我们只检查应用后是否得到目标文档
	}{
		{
			name:     "Add Property",
			source:   `{"foo":"bar"}`,
			target:   `{"foo":"bar","baz":"qux"}`,
			expected: `[{"op":"add","path":"/baz","value":"qux"}]`,
		},
		{
			name:     "Remove Property",
			source:   `{"foo":"bar","baz":"qux"}`,
			target:   `{"foo":"bar"}`,
			expected: `[{"op":"remove","path":"/baz"}]`,
		},
		{
			name:     "Replace Property",
			source:   `{"foo":"bar","baz":"qux"}`,
			target:   `{"foo":"bar","baz":"quux"}`,
			expected: `[{"op":"replace","path":"/baz","value":"quux"}]`,
		},
		{
			name:     "Replace Entire Document",
			source:   `{"foo":"bar"}`,
			target:   `["foo","bar"]`,
			expected: `[{"op":"replace","path":"/","value":["foo","bar"]}]`,
		},
		{
			name:     "Complex Change",
			source:   `{"foo":["bar","baz"],"qux":{"corge":"grault"}}`,
			target:   `{"foo":["baz","quux"],"qux":{"corge":"garply"}}`,
			expected: `[{"op":"replace","path":"/foo/0","value":"baz"},{"op":"replace","path":"/foo/1","value":"quux"},{"op":"replace","path":"/qux/corge","value":"garply"}]`,
		},
		{
			name:     "Array Length Change",
			source:   `["foo","bar","baz"]`,
			target:   `["foo","bar"]`,
			expected: `[{"op":"remove","path":"/2"}]`,
		},
		{
			name:     "Add to End of Array",
			source:   `["foo","bar"]`,
			target:   `["foo","bar","baz"]`,
			expected: `[{"op":"add","path":"/2","value":"baz"}]`,
		},
		{
			name:     "Nested Changes",
			source:   `{"foo":{"bar":{"baz":"qux"}}}`,
			target:   `{"foo":{"bar":{"baz":"quux","quuz":"corge"}}}`,
			expected: `[{"op":"replace","path":"/foo/bar/baz","value":"quux"},{"op":"add","path":"/foo/bar/quuz","value":"corge"}]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析源文档和目标文档
			source := &Value{}
			if err := Parse(source, tc.source); err != PARSE_OK {
				t.Fatalf("解析源文档失败: %s", GetErrorMessage(err))
			}

			target := &Value{}
			if err := Parse(target, tc.target); err != PARSE_OK {
				t.Fatalf("解析目标文档失败: %s", GetErrorMessage(err))
			}

			// 生成补丁
			patch, err := CreatePatch(source, target)
			if err != nil {
				t.Fatalf("创建 JSON Patch 失败: %v", err)
			}

			// 将补丁转换为字符串
			patchStr, err := patch.String()
			if err != nil {
				t.Fatalf("转换 JSON Patch 为字符串失败: %v", err)
			}

			t.Logf("生成的补丁: %s", patchStr)

			// 解析原始源文档的副本，用于应用补丁
			sourceDoc := &Value{}
			if err := Parse(sourceDoc, tc.source); err != PARSE_OK {
				t.Fatalf("解析源文档副本失败: %s", GetErrorMessage(err))
			}

			// 将生成的补丁应用到源文档副本
			if err := patch.Apply(sourceDoc); err != nil {
				t.Fatalf("应用 JSON Patch 失败: %v", err)
			}

			// 检查应用补丁后是否得到目标文档
			if !Equal(sourceDoc, target) {
				sourceStr, _ := Stringify(sourceDoc)
				targetStr, _ := Stringify(target)
				t.Fatalf("补丁应用后结果不匹配\n期望: %s\n实际: %s", targetStr, sourceStr)
			}

			// 可选：验证生成的补丁是否符合预期
			// 注意：生成的补丁可能不唯一，所以这只是一个额外的检查
			expected := &Value{}
			if err := Parse(expected, tc.expected); err != PARSE_OK {
				t.Logf("解析期望补丁失败，跳过额外验证: %s", GetErrorMessage(err))
				return
			}

			generated := &Value{}
			if err := Parse(generated, patchStr); err != PARSE_OK {
				t.Logf("解析生成补丁失败，跳过额外验证: %s", GetErrorMessage(err))
				return
			}

			// 如果预期的补丁和生成的补丁完全相同，这是理想情况
			// 但由于补丁不唯一，我们只是记录它们是否匹配，而不使测试失败
			if !Equal(generated, expected) {
				t.Logf("生成的补丁与预期补丁不完全匹配，但功能正确")
				t.Logf("预期: %s", tc.expected)
				t.Logf("生成: %s", patchStr)
			}
		})
	}
}
