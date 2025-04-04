// json_path_test.go - JSON Path 测试
package leptjson

import (
	"testing"
)

// 测试 NewJSONPath 基本功能
func TestNewJSONPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "空路径",
			path:    "",
			wantErr: true,
		},
		{
			name:    "不以$开头",
			path:    ".store.book",
			wantErr: true,
		},
		{
			name:    "有效路径",
			path:    "$",
			wantErr: false,
		},
		{
			name:    "有效路径带属性",
			path:    "$.store.book",
			wantErr: false,
		},
		{
			name:    "有效路径带数组索引",
			path:    "$.store.book[0]",
			wantErr: false,
		},
		{
			name:    "有效路径带通配符",
			path:    "$.store.book[*]",
			wantErr: false,
		},
		{
			name:    "有效路径带切片",
			path:    "$.store.book[1:3]",
			wantErr: false,
		},
		{
			name:    "有效路径带引号属性",
			path:    "$.store.book[0]['title']",
			wantErr: false,
		},
		{
			name:    "有效路径带递归下降",
			path:    "$..author",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jp, err := NewJSONPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJSONPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				// 检查解析后的令牌
				if len(jp.Tokens) == 0 {
					t.Errorf("解析结果不应为空")
				}

				// 第一个令牌必须是根节点
				if jp.Tokens[0].Type != ROOT {
					t.Errorf("第一个令牌必须是 ROOT，实际是 %v", jp.Tokens[0].Type)
				}
			}
		})
	}
}

// 创建测试用的 JSON 数据
func createTestData() *Value {
	// 创建示例数据结构:
	// {
	//   "store": {
	//     "book": [
	//       {
	//         "category": "reference",
	//         "author": "Nigel Rees",
	//         "title": "Sayings of the Century",
	//         "price": 8.95
	//       },
	//       {
	//         "category": "fiction",
	//         "author": "Evelyn Waugh",
	//         "title": "Sword of Honour",
	//         "price": 12.99
	//       }
	//     ],
	//     "bicycle": {
	//       "color": "red",
	//       "price": 19.95
	//     }
	//   }
	// }

	// 创建根对象
	root := &Value{}
	SetObject(root)

	// 创建 store 对象
	store := SetObjectValue(root, "store")
	SetObject(store)

	// 创建 book 数组
	book := SetObjectValue(store, "book")
	SetArray(book, 2)

	// 第一本书
	book1 := PushBackArrayElement(book)
	SetObject(book1)

	category1 := SetObjectValue(book1, "category")
	SetString(category1, "reference")

	author1 := SetObjectValue(book1, "author")
	SetString(author1, "Nigel Rees")

	title1 := SetObjectValue(book1, "title")
	SetString(title1, "Sayings of the Century")

	price1 := SetObjectValue(book1, "price")
	SetNumber(price1, 8.95)

	// 第二本书
	book2 := PushBackArrayElement(book)
	SetObject(book2)

	category2 := SetObjectValue(book2, "category")
	SetString(category2, "fiction")

	author2 := SetObjectValue(book2, "author")
	SetString(author2, "Evelyn Waugh")

	title2 := SetObjectValue(book2, "title")
	SetString(title2, "Sword of Honour")

	price2 := SetObjectValue(book2, "price")
	SetNumber(price2, 12.99)

	// 自行车
	bicycle := SetObjectValue(store, "bicycle")
	SetObject(bicycle)

	color := SetObjectValue(bicycle, "color")
	SetString(color, "red")

	price := SetObjectValue(bicycle, "price")
	SetNumber(price, 19.95)

	return root
}

// 测试基本的 JSON Path 查询功能
func TestBasicJSONPathQuery(t *testing.T) {
	doc := createTestData()

	tests := []struct {
		name      string
		path      string
		wantLen   int
		wantType  ValueType
		wantValue interface{} // 可能是字符串、数字或 nil (如果只关心类型)
	}{
		{
			name:      "根路径",
			path:      "$",
			wantLen:   1,
			wantType:  OBJECT,
			wantValue: nil,
		},
		{
			name:      "访问store对象",
			path:      "$.store",
			wantLen:   1,
			wantType:  OBJECT,
			wantValue: nil,
		},
		{
			name:      "访问book数组",
			path:      "$.store.book",
			wantLen:   1,
			wantType:  ARRAY,
			wantValue: nil,
		},
		{
			name:      "访问第一本书",
			path:      "$.store.book[0]",
			wantLen:   1,
			wantType:  OBJECT,
			wantValue: nil,
		},
		{
			name:      "访问第一本书的作者",
			path:      "$.store.book[0].author",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "Nigel Rees",
		},
		{
			name:      "访问第二本书的价格",
			path:      "$.store.book[1].price",
			wantLen:   1,
			wantType:  NUMBER,
			wantValue: 12.99,
		},
		{
			name:      "使用点符号访问自行车颜色",
			path:      "$.store.bicycle.color",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "red",
		},
		{
			name:      "使用括号符号访问自行车颜色",
			path:      "$.store.bicycle['color']",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "red",
		},
		{
			name:      "访问所有书的作者",
			path:      "$.store.book[*].author",
			wantLen:   2,
			wantType:  STRING,
			wantValue: nil, // 有多个结果
		},
		{
			name:      "使用负索引访问最后一本书",
			path:      "$.store.book[-1]",
			wantLen:   1,
			wantType:  OBJECT,
			wantValue: nil,
		},
		{
			name:      "使用切片访问前两本书",
			path:      "$.store.book[0:2]",
			wantLen:   2,
			wantType:  OBJECT,
			wantValue: nil,
		},
		{
			name:      "使用递归下降查找所有作者",
			path:      "$..author",
			wantLen:   2,
			wantType:  STRING,
			wantValue: nil,
		},
		{
			name:      "访问所有价格",
			path:      "$..price",
			wantLen:   3,
			wantType:  NUMBER,
			wantValue: nil,
		},
		{
			name:      "递归查找所有的类别",
			path:      "$..category",
			wantLen:   2,
			wantType:  STRING,
			wantValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := QueryString(doc, tt.path)
			if err != nil {
				t.Fatalf("QueryString() error = %v", err)
			}

			if len(results) != tt.wantLen {
				t.Errorf("QueryString() result length = %v, want %v", len(results), tt.wantLen)
			}

			if len(results) > 0 && results[0].Type != tt.wantType {
				t.Errorf("QueryString() first result type = %v, want %v", results[0].Type, tt.wantType)
			}

			// 检查具体值（如果提供了期望值）
			if tt.wantValue != nil && len(results) == 1 {
				switch tt.wantType {
				case STRING:
					if GetString(results[0]) != tt.wantValue.(string) {
						t.Errorf("QueryString() value = %v, want %v", GetString(results[0]), tt.wantValue)
					}
				case NUMBER:
					if GetNumber(results[0]) != tt.wantValue.(float64) {
						t.Errorf("QueryString() value = %v, want %v", GetNumber(results[0]), tt.wantValue)
					}
				}
			}
		})
	}
}

// 测试切片操作
func TestSliceOperations(t *testing.T) {
	// 创建一个有5个元素的数组
	array := &Value{}
	SetArray(array, 5)

	for i := 0; i < 5; i++ {
		elem := PushBackArrayElement(array)
		SetNumber(elem, float64(i+1))
	}

	tests := []struct {
		name     string
		path     string
		wantLen  int
		wantNums []float64
	}{
		{
			name:     "所有元素",
			path:     "$[*]",
			wantLen:  5,
			wantNums: []float64{1, 2, 3, 4, 5},
		},
		{
			name:     "前两个元素",
			path:     "$[0:2]",
			wantLen:  2,
			wantNums: []float64{1, 2},
		},
		{
			name:     "最后两个元素(使用负索引)",
			path:     "$[-2:]",
			wantLen:  2,
			wantNums: []float64{4, 5},
		},
		{
			name:     "每隔一个元素",
			path:     "$[0:5:2]",
			wantLen:  3,
			wantNums: []float64{1, 3, 5},
		},
		{
			name:     "反向顺序",
			path:     "$[4:0:-1]",
			wantLen:  4,
			wantNums: []float64{5, 4, 3, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := QueryString(array, tt.path)
			if err != nil {
				t.Fatalf("QueryString() error = %v", err)
			}

			if len(results) != tt.wantLen {
				t.Errorf("QueryString() result length = %v, want %v", len(results), tt.wantLen)
			}

			// 检查具体数值
			for i, result := range results {
				if i < len(tt.wantNums) && GetNumber(result) != tt.wantNums[i] {
					t.Errorf("QueryString() result[%d] = %v, want %v", i, GetNumber(result), tt.wantNums[i])
				}
			}
		})
	}
}

// 测试复杂嵌套结构的查询
func TestComplexStructure(t *testing.T) {
	// 创建一个更复杂的嵌套结构
	// {
	//   "users": [
	//     {
	//       "id": 1,
	//       "name": "John",
	//       "address": {
	//         "city": "New York",
	//         "country": "USA"
	//       },
	//       "phones": ["123-456-7890", "098-765-4321"]
	//     },
	//     {
	//       "id": 2,
	//       "name": "Alice",
	//       "address": {
	//         "city": "London",
	//         "country": "UK"
	//       },
	//       "phones": ["111-222-3333"]
	//     }
	//   ],
	//   "company": {
	//     "name": "Example Corp",
	//     "country": "USA"
	//   }
	// }

	root := &Value{}
	SetObject(root)

	// 用户数组
	users := SetObjectValue(root, "users")
	SetArray(users, 2)

	// 第一个用户
	user1 := PushBackArrayElement(users)
	SetObject(user1)

	id1 := SetObjectValue(user1, "id")
	SetNumber(id1, 1)

	name1 := SetObjectValue(user1, "name")
	SetString(name1, "John")

	address1 := SetObjectValue(user1, "address")
	SetObject(address1)

	city1 := SetObjectValue(address1, "city")
	SetString(city1, "New York")

	country1 := SetObjectValue(address1, "country")
	SetString(country1, "USA")

	phones1 := SetObjectValue(user1, "phones")
	SetArray(phones1, 2)

	phone1_1 := PushBackArrayElement(phones1)
	SetString(phone1_1, "123-456-7890")

	phone1_2 := PushBackArrayElement(phones1)
	SetString(phone1_2, "098-765-4321")

	// 第二个用户
	user2 := PushBackArrayElement(users)
	SetObject(user2)

	id2 := SetObjectValue(user2, "id")
	SetNumber(id2, 2)

	name2 := SetObjectValue(user2, "name")
	SetString(name2, "Alice")

	address2 := SetObjectValue(user2, "address")
	SetObject(address2)

	city2 := SetObjectValue(address2, "city")
	SetString(city2, "London")

	country2 := SetObjectValue(address2, "country")
	SetString(country2, "UK")

	phones2 := SetObjectValue(user2, "phones")
	SetArray(phones2, 1)

	phone2_1 := PushBackArrayElement(phones2)
	SetString(phone2_1, "111-222-3333")

	// 公司
	company := SetObjectValue(root, "company")
	SetObject(company)

	companyName := SetObjectValue(company, "name")
	SetString(companyName, "Example Corp")

	companyCountry := SetObjectValue(company, "country")
	SetString(companyCountry, "USA")

	tests := []struct {
		name      string
		path      string
		wantLen   int
		wantType  ValueType
		wantValue interface{}
	}{
		{
			name:      "所有用户",
			path:      "$.users",
			wantLen:   1,
			wantType:  ARRAY,
			wantValue: nil,
		},
		{
			name:      "第一个用户名",
			path:      "$.users[0].name",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "John",
		},
		{
			name:      "所有用户名",
			path:      "$.users[*].name",
			wantLen:   2,
			wantType:  STRING,
			wantValue: nil,
		},
		{
			name:      "第二个用户的城市",
			path:      "$.users[1].address.city",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "London",
		},
		{
			name:      "所有电话号码",
			path:      "$.users[*].phones[*]",
			wantLen:   3,
			wantType:  STRING,
			wantValue: nil,
		},
		{
			name:      "所有国家",
			path:      "$..country",
			wantLen:   3,
			wantType:  STRING,
			wantValue: nil,
		},
		{
			name:      "所有城市",
			path:      "$..city",
			wantLen:   2,
			wantType:  STRING,
			wantValue: nil,
		},
		{
			name:      "公司名称",
			path:      "$.company.name",
			wantLen:   1,
			wantType:  STRING,
			wantValue: "Example Corp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := QueryString(root, tt.path)
			if err != nil {
				t.Fatalf("QueryString() error = %v", err)
			}

			if len(results) != tt.wantLen {
				t.Errorf("QueryString() result length = %v, want %v", len(results), tt.wantLen)
			}

			if len(results) > 0 && results[0].Type != tt.wantType {
				t.Errorf("QueryString() first result type = %v, want %v", results[0].Type, tt.wantType)
			}

			// 检查具体值（如果提供了期望值）
			if tt.wantValue != nil && len(results) == 1 {
				switch tt.wantType {
				case STRING:
					if GetString(results[0]) != tt.wantValue.(string) {
						t.Errorf("QueryString() value = %v, want %v", GetString(results[0]), tt.wantValue)
					}
				case NUMBER:
					if GetNumber(results[0]) != tt.wantValue.(float64) {
						t.Errorf("QueryString() value = %v, want %v", GetNumber(results[0]), tt.wantValue)
					}
				}
			}
		})
	}
}

// 测试边缘情况
func TestEdgeCases(t *testing.T) {
	doc := createTestData()

	tests := []struct {
		name    string
		path    string
		wantErr bool
		wantLen int
	}{
		{
			name:    "无效路径",
			path:    "$.[",
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "不存在的属性",
			path:    "$.nonexistent",
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "不存在的数组索引",
			path:    "$.store.book[99]",
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "非数组使用索引",
			path:    "$.store.bicycle[0]",
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "非对象使用属性",
			path:    "$.store.book[0].price.value",
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := QueryString(doc, tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("QueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(results) != tt.wantLen {
				t.Errorf("QueryString() result length = %v, want %v", len(results), tt.wantLen)
			}
		})
	}
}

// 测试 QueryOne 函数
func TestQueryOne(t *testing.T) {
	doc := createTestData()

	tests := []struct {
		name      string
		path      string
		wantType  ValueType
		wantValue interface{}
		wantNil   bool
	}{
		{
			name:      "获取单个对象",
			path:      "$.store",
			wantType:  OBJECT,
			wantValue: nil,
			wantNil:   false,
		},
		{
			name:      "获取单个字符串",
			path:      "$.store.book[0].author",
			wantType:  STRING,
			wantValue: "Nigel Rees",
			wantNil:   false,
		},
		{
			name:      "获取单个数字",
			path:      "$.store.bicycle.price",
			wantType:  NUMBER,
			wantValue: 19.95,
			wantNil:   false,
		},
		{
			name:      "获取不存在的属性",
			path:      "$.nonexistent",
			wantType:  NULL,
			wantValue: nil,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := QueryOneString(doc, tt.path)
			if err != nil {
				t.Fatalf("QueryOneString() error = %v", err)
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("QueryOneString() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatalf("QueryOneString() = nil, want non-nil")
			}

			if result.Type != tt.wantType {
				t.Errorf("QueryOneString() type = %v, want %v", result.Type, tt.wantType)
			}

			// 检查具体值（如果提供了期望值）
			if tt.wantValue != nil {
				switch tt.wantType {
				case STRING:
					if GetString(result) != tt.wantValue.(string) {
						t.Errorf("QueryOneString() value = %v, want %v", GetString(result), tt.wantValue)
					}
				case NUMBER:
					if GetNumber(result) != tt.wantValue.(float64) {
						t.Errorf("QueryOneString() value = %v, want %v", GetNumber(result), tt.wantValue)
					}
				}
			}
		})
	}
}
