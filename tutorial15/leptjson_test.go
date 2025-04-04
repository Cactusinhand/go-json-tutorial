package leptjson // 确保包声明在最前面

import (
	"strings"
	"testing" // 引入 testing 包
)

// --- Marshal 测试 ---

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantJSON string
		wantErr  bool
	}{
		// 基本类型
		{"Null", nil, "null", false},
		{"True", true, "true", false},
		{"False", false, "false", false},
		{"Int", 123, "123", false},
		{"Float", 123.456, "123.456", false},
		{"String", "hello", `"hello"`, false},
		{"Empty String", "", `""`, false},
		{"String with Quotes", `"hello"`, `"\"hello\""`, false},
		{"String with Escapes", "\b\f\n\r\t\"\\/", `"\b\f\n\r\t\"\\/"`, false},

		// 切片/数组
		{"Empty Slice", []int{}, "[]", false},
		{"Int Slice", []int{1, 2, 3}, "[1,2,3]", false},
		{"String Slice", []string{"a", "b"}, `["a","b"]`, false},
		{"Mixed Slice", []interface{}{1, "a", true, nil}, `[1,"a",true,null]`, false},
		{"Nested Slice", [][]int{{1, 2}, {3, 4}}, "[[1,2],[3,4]]", false},

		// Map (string key)
		{"Empty Map", map[string]interface{}{}, "{}", false},
		{"Simple Map", map[string]interface{}{"a": 1, "b": "hello"}, `{"a":1,"b":"hello"}`, false}, // 注意 key 顺序不保证
		{"Nested Map", map[string]interface{}{"a": map[string]interface{}{"b": 1}}, `{"a":{"b":1}}`, false},

		// Structs 和 Tags
		{
			name: "Simple Struct",
			input: struct {
				Name string
				Age  int
			}{Name: "Bob", Age: 30},
			wantJSON: `{"Name":"Bob","Age":30}`,
		},
		{
			name: "Struct with Tags",
			input: struct {
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Password  string `json:"-"`               // 忽略
				Email     string `json:"email,omitempty"` // omitempty - 空
				Phone     string `json:"phone,omitempty"` // omitempty - 非空
			}{
				FirstName: "Alice",
				LastName:  "Smith",
				Password:  "secret",
				Email:     "",
				Phone:     "123-456",
			},
			wantJSON: `{"first_name":"Alice","last_name":"Smith","phone":"123-456"}`,
		},
		{
			name: "Struct with Unexported Field",
			input: struct {
				Exported   string
				unexported int // 小写，非导出
			}{
				Exported:   "visible",
				unexported: 42,
			},
			wantJSON: `{"Exported":"visible"}`,
		},
		{
			name: "Nested Struct",
			input: struct {
				User struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}
				Active bool `json:"active"`
			}{
				User: struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				}{
					ID:   1,
					Name: "User1",
				},
				Active: true,
			},
			wantJSON: `{"User":{"id":1,"name":"User1"},"active":true}`,
		},
		{
			name: "Struct with omitempty (zero values)",
			input: struct {
				A int    `json:",omitempty"`
				B string `json:",omitempty"`
				C bool   `json:",omitempty"`
			}{},
			wantJSON: `{}`,
		},
		{
			name: "Struct with omitempty (non-zero values)",
			input: struct {
				A int    `json:",omitempty"`
				B string `json:",omitempty"`
				C bool   `json:",omitempty"`
			}{A: 1, B: "s", C: true},
			wantJSON: `{"A":1,"B":"s","C":true}`,
		},

		// 指针
		{"Pointer to Int", newInt(5), "5", false},
		{"Nil Pointer", (*int)(nil), "null", false},
		{
			name:     "Pointer to Struct",
			input:    &struct{ Name string }{Name: "Ptr"},
			wantJSON: `{"Name":"Ptr"}`,
		},

		// 不支持的类型
		{"Function", func() {}, "", true},
		{"Channel", make(chan int), "", true},
		{"Map with non-string key", map[int]int{1: 1}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJSON, err := Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 对于 map，键的顺序不确定，需要比较解析后的结果
				if strings.Contains(tt.name, "Map") {
					gotVal := &Value{}
					wantVal := &Value{}
					if parseErr := Parse(gotVal, gotJSON); parseErr != PARSE_OK {
						t.Errorf("Marshal() produced invalid JSON: %s, error: %s", gotJSON, parseErr.Error())
						return
					}
					if parseErr := Parse(wantVal, tt.wantJSON); parseErr != PARSE_OK {
						t.Fatalf("Test case has invalid wantJSON: %s, error: %s", tt.wantJSON, parseErr.Error())
					}
					if !Equal(gotVal, wantVal) {
						t.Errorf("Marshal() Map = %s, want %s", gotJSON, tt.wantJSON)
					}
				} else if gotJSON != tt.wantJSON {
					t.Errorf("Marshal() = %v, want %v", gotJSON, tt.wantJSON)
				}
			}
		})
	}
}

// newInt 是一个辅助函数，用于创建 int 指针
func newInt(i int) *int {
	return &i
}

// ... (现有测试函数 TestParse, TestStringify, etc.) ...
