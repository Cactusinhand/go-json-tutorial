package tutorial14

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestNewJSONMergePatch(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		wantErr  bool
	}{
		{
			name:     "空补丁",
			document: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name: "简单补丁",
			document: map[string]interface{}{
				"foo": "bar",
				"baz": map[string]interface{}{
					"qux": 123,
				},
			},
			wantErr: false,
		},
		{
			name:     "含null的补丁",
			document: map[string]interface{}{"foo": nil},
			wantErr:  false,
		},
		{
			name:     "无效JSON类型",
			document: func() {},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJSONMergePatch(tt.document)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJSONMergePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && !reflect.DeepEqual(got.Document, tt.document) {
				t.Errorf("NewJSONMergePatch() got = %v, want %v", got.Document, tt.document)
			}
		})
	}
}

func TestApplyMergePatch(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		patch       string
		expected    string
		expectError bool
	}{
		{
			name:        "添加属性",
			target:      `{"foo": "bar"}`,
			patch:       `{"baz": "qux"}`,
			expected:    `{"foo": "bar", "baz": "qux"}`,
			expectError: false,
		},
		{
			name:        "修改属性",
			target:      `{"foo": "bar"}`,
			patch:       `{"foo": "qux"}`,
			expected:    `{"foo": "qux"}`,
			expectError: false,
		},
		{
			name:        "删除属性",
			target:      `{"foo": "bar", "baz": "qux"}`,
			patch:       `{"baz": null}`,
			expected:    `{"foo": "bar"}`,
			expectError: false,
		},
		{
			name:        "递归合并对象",
			target:      `{"foo": {"bar": "baz", "qux": "quux"}}`,
			patch:       `{"foo": {"bar": "updated"}}`,
			expected:    `{"foo": {"bar": "updated", "qux": "quux"}}`,
			expectError: false,
		},
		{
			name:        "递归删除嵌套属性",
			target:      `{"foo": {"bar": "baz", "qux": "quux"}}`,
			patch:       `{"foo": {"qux": null}}`,
			expected:    `{"foo": {"bar": "baz"}}`,
			expectError: false,
		},
		{
			name:        "替换数组",
			target:      `{"foo": ["bar", "baz"]}`,
			patch:       `{"foo": ["qux", "quux"]}`,
			expected:    `{"foo": ["qux", "quux"]}`,
			expectError: false,
		},
		{
			name:        "替换整个对象为数组",
			target:      `{"foo": {"bar": "baz"}}`,
			patch:       `{"foo": ["qux", "quux"]}`,
			expected:    `{"foo": ["qux", "quux"]}`,
			expectError: false,
		},
		{
			name:        "复杂示例 - 混合操作",
			target:      `{"title": "Hello", "author": {"name": "John Doe", "email": "john@example.com"}, "tags": ["news", "tech"], "views": 100}`,
			patch:       `{"title": "Hello World", "author": {"email": null}, "tags": ["news", "technology"], "views": null, "comments": 10}`,
			expected:    `{"title": "Hello World", "author": {"name": "John Doe"}, "tags": ["news", "technology"], "comments": 10}`,
			expectError: false,
		},
		{
			name:        "空补丁不修改文档",
			target:      `{"foo": "bar"}`,
			patch:       `{}`,
			expected:    `{"foo": "bar"}`,
			expectError: false,
		},
		{
			name:        "补丁为null",
			target:      `{"foo": "bar"}`,
			patch:       `null`,
			expected:    `null`,
			expectError: false,
		},
		{
			name:        "目标为null",
			target:      `null`,
			patch:       `{"foo": "bar"}`,
			expected:    `{"foo": "bar"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析测试数据
			var targetObj interface{}
			err := json.Unmarshal([]byte(tt.target), &targetObj)
			if err != nil {
				t.Fatalf("解析目标文档失败: %v", err)
			}

			var patchObj interface{}
			err = json.Unmarshal([]byte(tt.patch), &patchObj)
			if err != nil {
				t.Fatalf("解析补丁文档失败: %v", err)
			}

			var expectedObj interface{}
			err = json.Unmarshal([]byte(tt.expected), &expectedObj)
			if err != nil {
				t.Fatalf("解析期望结果失败: %v", err)
			}

			// 创建并应用补丁
			patch, err := NewJSONMergePatch(patchObj)
			if err != nil {
				t.Fatalf("创建补丁失败: %v", err)
			}

			result, err := patch.Apply(targetObj)
			if (err != nil) != tt.expectError {
				t.Errorf("Apply() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// 验证结果
			if !reflect.DeepEqual(result, expectedObj) {
				resultStr, _ := json.Marshal(result)
				expectedStr, _ := json.Marshal(expectedObj)
				t.Errorf("Apply() 结果不匹配\n预期: %s\n实际: %s", expectedStr, resultStr)
			}
		})
	}
}

func TestJSONMergePatchString(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		want     string
		wantErr  bool
	}{
		{
			name:     "Null Patch",
			document: nil,
			want:     "null",
			wantErr:  false,
		},
		{
			name: "Simple Patch",
			document: map[string]interface{}{
				"foo": "bar",
			},
			want:    `{"foo":"bar"}`,
			wantErr: false,
		},
		{
			name: "Complex Patch",
			document: map[string]interface{}{
				"foo": "bar",
				"baz": map[string]interface{}{
					"qux":   123,
					"array": []interface{}{1, 2, 3},
				},
			},
			want:    `{"baz":{"array":[1,2,3],"qux":123},"foo":"bar"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &JSONMergePatch{Document: tt.document}
			got, err := p.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// 标准化 JSON 字符串进行比较
			var gotObj, wantObj interface{}
			json.Unmarshal([]byte(got), &gotObj)
			json.Unmarshal([]byte(tt.want), &wantObj)

			if !reflect.DeepEqual(gotObj, wantObj) {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateMergePatch(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		target        string
		expectedPatch string
		expectError   bool
	}{
		{
			name:          "添加属性",
			source:        `{"foo": "bar"}`,
			target:        `{"foo": "bar", "baz": "qux"}`,
			expectedPatch: `{"baz": "qux"}`,
			expectError:   false,
		},
		{
			name:          "删除属性",
			source:        `{"foo": "bar", "baz": "qux"}`,
			target:        `{"foo": "bar"}`,
			expectedPatch: `{"baz": null}`,
			expectError:   false,
		},
		{
			name:          "修改属性",
			source:        `{"foo": "bar"}`,
			target:        `{"foo": "qux"}`,
			expectedPatch: `{"foo": "qux"}`,
			expectError:   false,
		},
		{
			name:          "修改嵌套属性",
			source:        `{"foo": {"bar": "baz", "qux": "quux"}}`,
			target:        `{"foo": {"bar": "updated", "qux": "quux"}}`,
			expectedPatch: `{"foo": {"bar": "updated"}}`,
			expectError:   false,
		},
		{
			name:          "删除嵌套属性",
			source:        `{"foo": {"bar": "baz", "qux": "quux"}}`,
			target:        `{"foo": {"bar": "baz"}}`,
			expectedPatch: `{"foo": {"qux": null}}`,
			expectError:   false,
		},
		{
			name:          "添加嵌套属性",
			source:        `{"foo": {"bar": "baz"}}`,
			target:        `{"foo": {"bar": "baz", "qux": "quux"}}`,
			expectedPatch: `{"foo": {"qux": "quux"}}`,
			expectError:   false,
		},
		{
			name:          "替换数组",
			source:        `{"foo": ["bar", "baz"]}`,
			target:        `{"foo": ["qux", "quux"]}`,
			expectedPatch: `{"foo": ["qux", "quux"]}`,
			expectError:   false,
		},
		{
			name:          "替换类型",
			source:        `{"foo": {"bar": "baz"}}`,
			target:        `{"foo": ["qux", "quux"]}`,
			expectedPatch: `{"foo": ["qux", "quux"]}`,
			expectError:   false,
		},
		{
			name:          "目标为null",
			source:        `{"foo": "bar"}`,
			target:        `null`,
			expectedPatch: `null`,
			expectError:   false,
		},
		{
			name:          "源为null",
			source:        `null`,
			target:        `{"foo": "bar"}`,
			expectedPatch: `{"foo": "bar"}`,
			expectError:   false,
		},
		{
			name:          "复杂示例 - RFC 7396 案例",
			source:        `{"title": "Hello", "author": {"name": "John Doe", "email": "john@example.com"}, "tags": ["news", "tech"], "views": 100}`,
			target:        `{"title": "Hello World", "author": {"name": "John Doe"}, "tags": ["news", "technology"], "comments": 10}`,
			expectedPatch: `{"comments": 10, "tags": ["news", "technology"], "title": "Hello World", "author": {"email": null}, "views": null}`,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 解析测试数据
			var sourceObj interface{}
			if err := json.Unmarshal([]byte(tt.source), &sourceObj); err != nil {
				t.Fatalf("解析源文档失败: %v", err)
			}

			var targetObj interface{}
			if err := json.Unmarshal([]byte(tt.target), &targetObj); err != nil {
				t.Fatalf("解析目标文档失败: %v", err)
			}

			var expectedPatchObj interface{}
			if err := json.Unmarshal([]byte(tt.expectedPatch), &expectedPatchObj); err != nil {
				t.Fatalf("解析期望补丁失败: %v", err)
			}

			// 创建补丁
			patch, err := CreateMergePatch(sourceObj, targetObj)
			if (err != nil) != tt.expectError {
				t.Errorf("CreateMergePatch() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if err != nil {
				return
			}

			// 验证生成的补丁
			patchStr, _ := patch.String()
			fmt.Printf("生成的补丁: %s\n", patchStr)

			// 标准化 JSON 对象进行深度比较
			if !reflect.DeepEqual(normalizePatch(patch.Document), normalizePatch(expectedPatchObj)) {
				expectedPatchStr, _ := json.Marshal(expectedPatchObj)
				patchStr, _ := json.Marshal(patch.Document)
				t.Errorf("CreateMergePatch() 补丁不匹配\n预期: %s\n实际: %s", expectedPatchStr, patchStr)
			}

			// 验证应用补丁后的结果
			result, err := patch.Apply(sourceObj)
			if err != nil {
				t.Errorf("应用生成的补丁失败: %v", err)
				return
			}

			// 标准化 JSON 对象进行深度比较
			if !reflect.DeepEqual(normalizePatch(result), normalizePatch(targetObj)) {
				resultStr, _ := json.Marshal(result)
				targetStr, _ := json.Marshal(targetObj)
				t.Errorf("应用生成的补丁后结果不匹配\n预期: %s\n实际: %s", targetStr, resultStr)
			}
		})
	}
}

// normalizePatch 标准化 JSON 对象，忽略属性顺序和空对象差异
func normalizePatch(obj interface{}) interface{} {
	if obj == nil {
		return nil
	}

	switch v := obj.(type) {
	case map[string]interface{}:
		// 如果是空对象，直接返回
		if len(v) == 0 {
			return v
		}

		// 对所有嵌套对象递归处理
		result := make(map[string]interface{})
		for key, value := range v {
			result[key] = normalizePatch(value)
		}
		return result

	case []interface{}:
		// 处理数组中的每个元素
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = normalizePatch(item)
		}
		return result

	default:
		return obj
	}
}
