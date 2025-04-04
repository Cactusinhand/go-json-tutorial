package leptjson

import (
	"strings"
	"testing"
)

func TestNewJSONMergePatch(t *testing.T) {
	tests := []struct {
		name         string
		documentJSON string
		wantErr      bool
	}{
		{
			name:         "空对象补丁",
			documentJSON: `{}`,
			wantErr:      false,
		},
		{
			name: "简单对象补丁",
			documentJSON: `{
				"foo": "bar",
				"baz": {
					"qux": 123
				}
			}`,
			wantErr: false,
		},
		{
			name:         "含null的补丁",
			documentJSON: `{"foo": null}`,
			wantErr:      false,
		},
		{
			name:         "非对象或null补丁 (也有效)",
			documentJSON: `"a string patch"`,
			wantErr:      false,
		},
		{
			name:         "数组补丁 (也有效)",
			documentJSON: `[1, 2, 3]`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documentVal := &Value{}
			if errCode := Parse(documentVal, tt.documentJSON); errCode != PARSE_OK {
				t.Fatalf("解析文档失败: %s", errCode.Error())
			}

			got, err := NewJSONMergePatch(documentVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJSONMergePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !Equal(got.Document, documentVal) {
				gotStr, _ := Stringify(got.Document)
				wantStr, _ := Stringify(documentVal)
				t.Errorf("NewJSONMergePatch() got = %s, want %s", gotStr, wantStr)
			}
		})
	}
}

func TestApplyMergePatch(t *testing.T) {
	tests := []struct {
		name                string
		targetJSON          string
		patchJSON           string
		expectedJSON        string
		expectError         bool
		expectErrorContains string
	}{
		{
			name:         "添加属性",
			targetJSON:   `{"foo": "bar"}`,
			patchJSON:    `{"baz": "qux"}`,
			expectedJSON: `{"foo": "bar", "baz": "qux"}`,
		},
		{
			name:         "修改属性",
			targetJSON:   `{"foo": "bar"}`,
			patchJSON:    `{"foo": "qux"}`,
			expectedJSON: `{"foo": "qux"}`,
		},
		{
			name:         "删除属性",
			targetJSON:   `{"foo": "bar", "baz": "qux"}`,
			patchJSON:    `{"baz": null}`,
			expectedJSON: `{"foo": "bar"}`,
		},
		{
			name:         "递归合并对象",
			targetJSON:   `{"foo": {"bar": "baz", "qux": "quux"}}`,
			patchJSON:    `{"foo": {"bar": "updated"}}`,
			expectedJSON: `{"foo": {"bar": "updated", "qux": "quux"}}`,
		},
		{
			name:         "递归删除嵌套属性",
			targetJSON:   `{"foo": {"bar": "baz", "qux": "quux"}}`,
			patchJSON:    `{"foo": {"qux": null}}`,
			expectedJSON: `{"foo": {"bar": "baz"}}`,
		},
		{
			name:         "替换数组",
			targetJSON:   `{"foo": ["bar", "baz"]}`,
			patchJSON:    `{"foo": ["qux", "quux"]}`,
			expectedJSON: `{"foo": ["qux", "quux"]}`,
		},
		{
			name:         "替换整个对象为数组",
			targetJSON:   `{"foo": {"bar": "baz"}}`,
			patchJSON:    `{"foo": ["qux", "quux"]}`,
			expectedJSON: `{"foo": ["qux", "quux"]}`,
		},
		{
			name:         "复杂示例 - 混合操作",
			targetJSON:   `{"title": "Hello", "author": {"name": "John Doe", "email": "john@example.com"}, "tags": ["news", "tech"], "views": 100}`,
			patchJSON:    `{"title": "Hello World", "author": {"email": null}, "tags": ["news", "technology"], "views": null, "comments": 10}`,
			expectedJSON: `{"title": "Hello World", "author": {"name": "John Doe"}, "tags": ["news", "technology"], "comments": 10}`,
		},
		{
			name:         "空补丁不修改文档",
			targetJSON:   `{"foo": "bar"}`,
			patchJSON:    `{}`,
			expectedJSON: `{"foo": "bar"}`,
		},
		{
			name:         "补丁为null",
			targetJSON:   `{"foo": "bar"}`,
			patchJSON:    `null`,
			expectedJSON: `null`,
		},
		{
			name:         "目标为null",
			targetJSON:   `null`,
			patchJSON:    `{"foo": "bar"}`,
			expectedJSON: `{"foo": "bar"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targetVal := &Value{}
			if errCode := Parse(targetVal, tt.targetJSON); errCode != PARSE_OK {
				t.Fatalf("解析目标文档失败: %s", errCode.Error())
			}

			patchVal := &Value{}
			if errCode := Parse(patchVal, tt.patchJSON); errCode != PARSE_OK {
				t.Fatalf("解析补丁文档失败: %s", errCode.Error())
			}

			expectedVal := &Value{}
			if errCode := Parse(expectedVal, tt.expectedJSON); errCode != PARSE_OK {
				t.Fatalf("解析期望结果失败: %s", errCode.Error())
			}

			patch, err := NewJSONMergePatch(patchVal)
			if err != nil {
				t.Fatalf("创建补丁失败: %v", err)
			}

			resultVal, applyErr := patch.Apply(targetVal)

			if tt.expectError {
				if applyErr == nil {
					t.Errorf("Apply() 期望错误但没有得到错误")
				}
				if tt.expectErrorContains != "" && (applyErr == nil || !strings.Contains(applyErr.Error(), tt.expectErrorContains)) {
					t.Errorf("Apply() 错误信息 %q 不包含期望的 %q", applyErr, tt.expectErrorContains)
				}
			} else {
				if applyErr != nil {
					t.Errorf("Apply() 期望成功但得到错误: %v", applyErr)
					return
				}
				if !Equal(resultVal, expectedVal) {
					resultStr, _ := Stringify(resultVal)
					expectedStr, _ := Stringify(expectedVal)
					t.Errorf("Apply() 结果不匹配\n预期: %s\n实际: %s", expectedStr, resultStr)
				}
			}
		})
	}
}

func TestJSONMergePatchString(t *testing.T) {
	tests := []struct {
		name         string
		documentJSON string
		want         string
		wantErr      bool
	}{
		{
			name:         "Null Patch",
			documentJSON: `null`,
			want:         "null",
			wantErr:      false,
		},
		{
			name:         "Empty Object Patch",
			documentJSON: `{}`,
			want:         "{}",
			wantErr:      false,
		},
		{
			name:         "Simple Object Patch",
			documentJSON: `{"a": 1, "b": "hello"}`,
			want:         `{"a":1,"b":"hello"}`,
			wantErr:      false,
		},
		{
			name:         "String Patch",
			documentJSON: `"patch string"`,
			want:         `"patch string"`,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			documentVal := &Value{}
			if errCode := Parse(documentVal, tt.documentJSON); errCode != PARSE_OK {
				t.Fatalf("解析文档失败: %s", errCode.Error())
			}

			p := &JSONMergePatch{Document: documentVal}
			got, err := p.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateMergePatch(t *testing.T) {
	tests := []struct {
		name              string
		sourceJSON        string
		targetJSON        string
		expectedPatchJSON string
		expectError       bool
	}{
		{
			name:              "No Change",
			sourceJSON:        `{"a": 1}`,
			targetJSON:        `{"a": 1}`,
			expectedPatchJSON: `{}`,
		},
		{
			name:              "Add Property",
			sourceJSON:        `{"a": 1}`,
			targetJSON:        `{"a": 1, "b": 2}`,
			expectedPatchJSON: `{"b": 2}`,
		},
		{
			name:              "Remove Property",
			sourceJSON:        `{"a": 1, "b": 2}`,
			targetJSON:        `{"a": 1}`,
			expectedPatchJSON: `{"b": null}`,
		},
		{
			name:              "Change Property",
			sourceJSON:        `{"a": 1}`,
			targetJSON:        `{"a": 2}`,
			expectedPatchJSON: `{"a": 2}`,
		},
		{
			name:              "Nested Change Add",
			sourceJSON:        `{"a": {"b": 1}}`,
			targetJSON:        `{"a": {"b": 1, "c": 2}}`,
			expectedPatchJSON: `{"a": {"c": 2}}`,
		},
		{
			name:              "Nested Change Remove",
			sourceJSON:        `{"a": {"b": 1, "c": 2}}`,
			targetJSON:        `{"a": {"b": 1}}`,
			expectedPatchJSON: `{"a": {"c": null}}`,
		},
		{
			name:              "Nested Change Modify",
			sourceJSON:        `{"a": {"b": 1}}`,
			targetJSON:        `{"a": {"b": 2}}`,
			expectedPatchJSON: `{"a": {"b": 2}}`,
		},
		{
			name:              "Replace Object with Value",
			sourceJSON:        `{"a": {"b": 1}}`,
			targetJSON:        `{"a": 2}`,
			expectedPatchJSON: `{"a": 2}`,
		},
		{
			name:              "Replace Value with Object",
			sourceJSON:        `{"a": 1}`,
			targetJSON:        `{"a": {"b": 2}}`,
			expectedPatchJSON: `{"a": {"b": 2}}`,
		},
		{
			name:              "Replace Array",
			sourceJSON:        `{"a": [1, 2]}`,
			targetJSON:        `{"a": [3, 4]}`,
			expectedPatchJSON: `{"a": [3, 4]}`,
		},
		{
			name:              "Target is Null",
			sourceJSON:        `{"a": 1}`,
			targetJSON:        `null`,
			expectedPatchJSON: `null`,
		},
		{
			name:              "Source is Null",
			sourceJSON:        `null`,
			targetJSON:        `{"a": 1}`,
			expectedPatchJSON: `{"a": 1}`,
		},
		{
			name:              "Complex Diff",
			sourceJSON:        `{"title": "Hello", "author": {"name": "John Doe", "email": "john@example.com"}, "tags": ["news", "tech"], "views": 100}`,
			targetJSON:        `{"title": "Hello World", "author": {"name": "John Doe"}, "tags": ["news", "technology"], "comments": 10}`,
			expectedPatchJSON: `{"title": "Hello World", "author": {"email": null}, "tags": ["news", "technology"], "views": null, "comments": 10}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceVal := &Value{}
			if errCode := Parse(sourceVal, tt.sourceJSON); errCode != PARSE_OK {
				t.Fatalf("解析源文档失败: %s", errCode.Error())
			}
			targetVal := &Value{}
			if errCode := Parse(targetVal, tt.targetJSON); errCode != PARSE_OK {
				t.Fatalf("解析目标文档失败: %s", errCode.Error())
			}
			expectedPatchVal := &Value{}
			if errCode := Parse(expectedPatchVal, tt.expectedPatchJSON); errCode != PARSE_OK {
				t.Fatalf("解析期望补丁失败: %s", errCode.Error())
			}

			patch, err := CreateMergePatch(sourceVal, targetVal)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateMergePatch() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if err != nil {
				return
			}

			if !Equal(patch.Document, expectedPatchVal) {
				patchStr, _ := patch.String()
				expectedPatchStr, _ := Stringify(expectedPatchVal)
				t.Errorf("CreateMergePatch() 生成的补丁不匹配\n预期: %s\n实际: %s", expectedPatchStr, patchStr)
			}

			resultVal, applyErr := patch.Apply(sourceVal)
			if applyErr != nil {
				t.Errorf("Apply() 应用补丁时发生意外错误: %v", applyErr)
				return
			}

			if !Equal(resultVal, targetVal) {
				resultStr, _ := Stringify(resultVal)
				targetStr, _ := Stringify(targetVal)
				t.Errorf("Apply() 应用补丁后结果不等于目标文档\n目标: %s\n实际: %s", targetStr, resultStr)
			}
		})
	}
}
