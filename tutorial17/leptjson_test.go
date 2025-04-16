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

// 测试安全特性
func TestSecurity(t *testing.T) {
	// 测试用例
	tests := []struct {
		name      string
		json      string
		options   ParseOptions
		expectErr ParseError
	}{
		{
			name: "大小超限",
			json: strings.Repeat("x", 2000),
			options: ParseOptions{
				MaxTotalSize:    1000,
				MaxDepth:        1000, // 足够大
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_TOTAL_SIZE_EXCEEDED,
		},
		{
			name: "字符串长度超限",
			json: `"` + strings.Repeat("x", 1000) + `"`,
			options: ParseOptions{
				MaxStringLength: 500,
				MaxTotalSize:    10000, // 足够大
				MaxDepth:        1000,  // 足够大
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_STRING_LENGTH_EXCEEDED,
		},
		{
			name: "数组元素超限",
			json: `[1,2,3,4,5,6,7,8,9,10]`,
			options: ParseOptions{
				MaxArraySize:    5,
				MaxTotalSize:    10000,  // 足够大
				MaxDepth:        1000,   // 足够大
				MaxNumberValue:  1e308,  // 确保数值检查不会干扰
				MinNumberValue:  -1e308, // 确保数值检查不会干扰
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_ARRAY_SIZE_EXCEEDED,
		},
		{
			name: "对象成员超限",
			json: `{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6}`,
			options: ParseOptions{
				MaxObjectSize:   5,
				MaxTotalSize:    10000,  // 足够大
				MaxDepth:        1000,   // 足够大
				MaxNumberValue:  1e308,  // 确保数值检查不会干扰
				MinNumberValue:  -1e308, // 确保数值检查不会干扰
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_OBJECT_SIZE_EXCEEDED,
		},
		{
			name: "数值范围上限超限",
			json: `1e100`,
			options: ParseOptions{
				MaxNumberValue:  1e10,
				MinNumberValue:  -1e308, // 设置一个非常大的最小值，避免干扰
				MaxTotalSize:    10000,  // 足够大
				MaxDepth:        1000,   // 足够大
				EnabledSecurity: true,
			},
			expectErr: PARSE_NUMBER_RANGE_EXCEEDED,
		},
		{
			name: "数值范围下限超限",
			json: `-1e100`,
			options: ParseOptions{
				MinNumberValue:  -1e10,
				MaxNumberValue:  1e308, // 设置一个非常大的最大值，避免干扰
				MaxTotalSize:    10000, // 足够大
				MaxDepth:        1000,  // 足够大
				EnabledSecurity: true,
			},
			expectErr: PARSE_NUMBER_RANGE_EXCEEDED,
		},
		{
			name: "递归深度超限",
			json: `[[[[[[[[[[1]]]]]]]]]]`,
			options: ParseOptions{
				MaxDepth:        5,
				MaxTotalSize:    10000, // 足够大
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_DEPTH_EXCEEDED,
		},
		{
			name: "安全检查禁用时无错误",
			json: `{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6}`,
			options: ParseOptions{
				MaxObjectSize:   5,
				MaxTotalSize:    10000, // 足够大
				MaxDepth:        1000,  // 足够大
				EnabledSecurity: false,
			},
			expectErr: PARSE_OK,
		},
		{
			name: "嵌套对象检查",
			json: `{"obj":{"a":1,"b":2,"c":3,"d":4,"e":5,"f":6}}`,
			options: ParseOptions{
				MaxObjectSize:   5,
				MaxTotalSize:    10000,  // 足够大
				MaxDepth:        1000,   // 足够大
				MaxNumberValue:  1e308,  // 确保数值检查不会干扰
				MinNumberValue:  -1e308, // 确保数值检查不会干扰
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_OBJECT_SIZE_EXCEEDED,
		},
		{
			name: "嵌套数组检查",
			json: `[1,[1,2,3,4,5,6,7,8,9,10]]`,
			options: ParseOptions{
				MaxArraySize:    5,
				MaxTotalSize:    10000,  // 足够大
				MaxDepth:        1000,   // 足够大
				MaxNumberValue:  1e308,  // 确保数值检查不会干扰
				MinNumberValue:  -1e308, // 确保数值检查不会干扰
				EnabledSecurity: true,
			},
			expectErr: PARSE_MAX_ARRAY_SIZE_EXCEEDED,
		},
	}

	// 执行测试
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Value{}
			err := ParseWithOptions(&v, tt.json, tt.options)
			if err != tt.expectErr {
				t.Errorf("Parse(%q) error = %v (code=%d), wantErr %v (code=%d)",
					tt.json, GetErrorMessage(err), err, GetErrorMessage(tt.expectErr), tt.expectErr)
			}
		})
	}
}

// 性能测试：验证安全检查开启与关闭的性能差异
func BenchmarkParseSecurity(b *testing.B) {
	// 生成测试用的大型JSON
	jsonData := `{"array":[` + strings.Repeat(`1,`, 1000) + `1],"string":"` + strings.Repeat("x", 1000) + `"}`

	// 不开启安全检查
	b.Run("SecurityDisabled", func(b *testing.B) {
		options := DefaultParseOptions()
		options.EnabledSecurity = false

		for i := 0; i < b.N; i++ {
			v := Value{}
			ParseWithOptions(&v, jsonData, options)
		}
	})

	// 开启安全检查
	b.Run("SecurityEnabled", func(b *testing.B) {
		options := DefaultParseOptions()
		options.EnabledSecurity = true

		for i := 0; i < b.N; i++ {
			v := Value{}
			ParseWithOptions(&v, jsonData, options)
		}
	})
}
