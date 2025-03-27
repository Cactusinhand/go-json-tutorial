package leptjson

import (
	"testing"
)

// 测试JSON指针的解析
func TestParseJSONPointer(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   JSONPointerError
		wantParts []string
	}{
		{
			name:      "空指针",
			input:     "",
			wantErr:   POINTER_OK,
			wantParts: []string{},
		},
		{
			name:      "根指针",
			input:     "/",
			wantErr:   POINTER_OK,
			wantParts: []string{""},
		},
		{
			name:      "简单路径",
			input:     "/foo/bar",
			wantErr:   POINTER_OK,
			wantParts: []string{"foo", "bar"},
		},
		{
			name:      "数组索引",
			input:     "/foo/0/bar",
			wantErr:   POINTER_OK,
			wantParts: []string{"foo", "0", "bar"},
		},
		{
			name:      "转义字符",
			input:     "/foo~1bar/~0baz",
			wantErr:   POINTER_OK,
			wantParts: []string{"foo/bar", "~baz"},
		},
		{
			name:    "无效格式",
			input:   "foo/bar",
			wantErr: POINTER_INVALID_FORMAT,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSONPointer(tt.input)

			if err != tt.wantErr {
				t.Errorf("ParseJSONPointer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == POINTER_OK {
				if len(got.tokens) != len(tt.wantParts) {
					t.Errorf("ParseJSONPointer() tokens length = %v, want %v", len(got.tokens), len(tt.wantParts))
					return
				}

				for i, token := range got.tokens {
					if token != tt.wantParts[i] {
						t.Errorf("ParseJSONPointer() token[%d] = %v, want %v", i, token, tt.wantParts[i])
					}
				}
			}
		})
	}
}

// 测试获取JSON值
func TestJSONPointerGet(t *testing.T) {
	// 准备测试数据
	// {"foo": {"bar": 42}, "baz": [0, 1, {"qux": "hello"}]}
	doc := &Value{}
	SetObject(doc)

	foo := SetObjectValue(doc, "foo")
	SetObject(foo)
	bar := SetObjectValue(foo, "bar")
	SetNumber(bar, 42)

	baz := SetObjectValue(doc, "baz")
	SetArray(baz, 3)
	SetNumber(PushBackArrayElement(baz), 0)
	SetNumber(PushBackArrayElement(baz), 1)
	quxObj := PushBackArrayElement(baz)
	SetObject(quxObj)
	qux := SetObjectValue(quxObj, "qux")
	SetString(qux, "hello")

	tests := []struct {
		name      string
		pointer   string
		wantType  ValueType
		wantValue interface{}
		wantErr   JSONPointerError
	}{
		{
			name:     "根文档",
			pointer:  "",
			wantType: OBJECT,
			wantErr:  POINTER_OK,
		},
		{
			name:      "对象属性",
			pointer:   "/foo/bar",
			wantType:  NUMBER,
			wantValue: 42.0,
			wantErr:   POINTER_OK,
		},
		{
			name:      "数组元素",
			pointer:   "/baz/1",
			wantType:  NUMBER,
			wantValue: 1.0,
			wantErr:   POINTER_OK,
		},
		{
			name:      "嵌套路径",
			pointer:   "/baz/2/qux",
			wantType:  STRING,
			wantValue: "hello",
			wantErr:   POINTER_OK,
		},
		{
			name:    "不存在的键",
			pointer: "/foo/unknown",
			wantErr: POINTER_KEY_NOT_FOUND,
		},
		{
			name:    "超出范围的索引",
			pointer: "/baz/5",
			wantErr: POINTER_INDEX_OUT_OF_RANGE,
		},
		{
			name:    "无效索引",
			pointer: "/baz/not-a-number",
			wantErr: POINTER_INDEX_OUT_OF_RANGE,
		},
		{
			name:    "无效目标类型",
			pointer: "/foo/bar/impossible",
			wantErr: POINTER_INVALID_TARGET,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pointer, pErr := ParseJSONPointer(tt.pointer)
			if pErr != POINTER_OK {
				t.Errorf("无法解析JSON指针: %v", pErr)
				return
			}

			got, err := pointer.Get(doc)

			if err != tt.wantErr {
				t.Errorf("JSONPointer.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == POINTER_OK {
				if got.Type != tt.wantType {
					t.Errorf("JSONPointer.Get() type = %v, want %v", got.Type, tt.wantType)
				}

				switch tt.wantType {
				case NUMBER:
					if GetNumber(got) != tt.wantValue.(float64) {
						t.Errorf("JSONPointer.Get() value = %v, want %v", GetNumber(got), tt.wantValue)
					}
				case STRING:
					if GetString(got) != tt.wantValue.(string) {
						t.Errorf("JSONPointer.Get() value = %v, want %v", GetString(got), tt.wantValue)
					}
				}
			}
		})
	}
}

// 测试设置JSON值
func TestJSONPointerSet(t *testing.T) {
	// 准备基础JSON文档
	doc := &Value{}
	SetObject(doc)

	foo := SetObjectValue(doc, "foo")
	SetObject(foo)

	baz := SetObjectValue(doc, "baz")
	SetArray(baz, 2)
	SetNumber(PushBackArrayElement(baz), 0)
	SetNumber(PushBackArrayElement(baz), 1)

	tests := []struct {
		name     string
		pointer  string
		setValue func() *Value
		wantErr  JSONPointerError
		check    func(t *testing.T, doc *Value)
	}{
		{
			name:    "设置对象属性",
			pointer: "/foo/bar",
			setValue: func() *Value {
				v := &Value{}
				SetNumber(v, 42)
				return v
			},
			wantErr: POINTER_OK,
			check: func(t *testing.T, doc *Value) {
				v, err := GetValueByPointer(doc, "/foo/bar")
				if err != nil {
					t.Errorf("检查失败: %v", err)
					return
				}
				if v.Type != NUMBER || GetNumber(v) != 42 {
					t.Errorf("设置值不正确, got %v", v)
				}
			},
		},
		{
			name:    "更新数组元素",
			pointer: "/baz/1",
			setValue: func() *Value {
				v := &Value{}
				SetString(v, "updated")
				return v
			},
			wantErr: POINTER_OK,
			check: func(t *testing.T, doc *Value) {
				v, err := GetValueByPointer(doc, "/baz/1")
				if err != nil {
					t.Errorf("检查失败: %v", err)
					return
				}
				if v.Type != STRING || GetString(v) != "updated" {
					t.Errorf("设置值不正确, got %v", v)
				}
			},
		},
		{
			name:    "创建不存在的路径",
			pointer: "/foo/newpath/deep",
			setValue: func() *Value {
				v := &Value{}
				SetBoolean(v, true)
				return v
			},
			wantErr: POINTER_KEY_NOT_FOUND,
		},
		{
			name:    "索引超出范围",
			pointer: "/baz/5",
			setValue: func() *Value {
				v := &Value{}
				SetNull(v)
				return v
			},
			wantErr: POINTER_INDEX_OUT_OF_RANGE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 对每个测试创建文档的副本
			testDoc := &Value{}
			Copy(testDoc, doc)

			// 解析JSON指针
			pointer, _ := ParseJSONPointer(tt.pointer)

			// 设置值
			err := pointer.Set(testDoc, tt.setValue())

			if err != tt.wantErr {
				t.Errorf("JSONPointer.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == POINTER_OK && tt.check != nil {
				tt.check(t, testDoc)
			}
		})
	}
}
