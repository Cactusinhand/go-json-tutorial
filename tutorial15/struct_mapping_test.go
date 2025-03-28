package tutorial15

import (
	"testing"
	"time"
)

// Person 测试用结构体
type Person struct {
	Name     string                 `json:"name"`
	Age      int                    `json:"age,omitempty"`
	Email    string                 `json:"email"`
	Address  string                 `json:"-"` // 不序列化
	Birthday JSONTime               `json:"birthday"`
	Tags     []string               `json:"tags,omitempty"`
	Scores   []int                  `json:"scores"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// 测试结构体转 JSON
func TestStructToJSON(t *testing.T) {
	// 创建一个测试结构体
	person := Person{
		Name:     "John Doe",
		Age:      30,
		Email:    "john@example.com",
		Address:  "123 Main St", // 不会被序列化
		Birthday: JSONTime(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)),
		Tags:     []string{"developer", "go"},
		Scores:   []int{95, 85, 90},
		Metadata: map[string]interface{}{
			"role":        "admin",
			"department":  "engineering",
			"employee_id": 12345,
		},
	}

	// 转换为 JSON
	jsonVal, err := StructToJSON(person)
	if err != nil {
		t.Fatalf("StructToJSON failed: %v", err)
	}

	// 验证类型和字段
	if jsonVal.Type != OBJECT {
		t.Errorf("Expected OBJECT type, got %s", GetTypeStr(jsonVal))
	}

	// 检查特定字段
	nameValue := FindObjectValue(jsonVal, "name")
	if nameValue == nil || nameValue.Type != STRING || nameValue.S != "John Doe" {
		t.Errorf("Name field not correctly serialized")
	}

	// 检查 Age 字段
	ageValue := FindObjectValue(jsonVal, "age")
	if ageValue == nil || ageValue.Type != NUMBER || ageValue.N != 30 {
		t.Errorf("Age field not correctly serialized")
	}

	// 确认 Address 字段没有被序列化
	addressValue := FindObjectValue(jsonVal, "Address")
	if addressValue != nil {
		t.Errorf("Address field should not be serialized")
	}

	// 将 JSON 转换为字符串并打印
	jsonStr, err := Stringify(jsonVal)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	t.Logf("JSON result: %s", jsonStr)
}

// 测试 omitempty 标签
func TestOmitEmpty(t *testing.T) {
	// 创建一个有空字段的结构体
	person := Person{
		Name:   "Jane Doe",
		Email:  "jane@example.com",
		Scores: []int{}, // 空切片，但没有 omitempty，应该被序列化
	}

	// 转换为 JSON
	jsonVal, err := StructToJSON(person)
	if err != nil {
		t.Fatalf("StructToJSON failed: %v", err)
	}

	// 确认 Age 字段没有被序列化
	ageValue := FindObjectValue(jsonVal, "age")
	if ageValue != nil {
		t.Errorf("Age field should be omitted when empty")
	}

	// 确认 Tags 字段没有被序列化
	tagsValue := FindObjectValue(jsonVal, "tags")
	if tagsValue != nil {
		t.Errorf("Tags field should be omitted when empty")
	}

	// 确认 Scores 字段被序列化了
	scoresValue := FindObjectValue(jsonVal, "scores")
	if scoresValue == nil || scoresValue.Type != ARRAY || len(scoresValue.Elements) != 0 {
		t.Errorf("Scores field should be serialized even when empty")
	}
}

// 测试 JSON 转结构体
func TestJSONToStruct(t *testing.T) {
	// 创建一个 JSON 对象
	v := &Value{}
	SetObject(v)

	// 添加 name 字段
	nameValue := SetObjectValue(v, "name")
	SetString(nameValue, "Alice Smith")

	// 添加 age 字段
	ageValue := SetObjectValue(v, "age")
	SetNumber(ageValue, 25)

	// 添加 email 字段
	emailValue := SetObjectValue(v, "email")
	SetString(emailValue, "alice@example.com")

	// 添加生日字段
	birthdayValue := SetObjectValue(v, "birthday")
	SetString(birthdayValue, "1995-02-15T00:00:00Z")

	// 添加标签数组
	tagsValue := SetObjectValue(v, "tags")
	SetArray(tagsValue)
	tag1 := &Value{}
	SetString(tag1, "engineer")
	PushBackArrayElement(tagsValue, tag1)
	tag2 := &Value{}
	SetString(tag2, "designer")
	PushBackArrayElement(tagsValue, tag2)

	// 转换为结构体
	var person Person
	err := JSONToStruct(v, &person)
	if err != nil {
		t.Fatalf("JSONToStruct failed: %v", err)
	}

	// 验证字段
	if person.Name != "Alice Smith" {
		t.Errorf("Expected name 'Alice Smith', got '%s'", person.Name)
	}

	if person.Age != 25 {
		t.Errorf("Expected age 25, got %d", person.Age)
	}

	if person.Email != "alice@example.com" {
		t.Errorf("Expected email 'alice@example.com', got '%s'", person.Email)
	}

	if len(person.Tags) != 2 || person.Tags[0] != "engineer" || person.Tags[1] != "designer" {
		t.Errorf("Tags not correctly deserialized, got %v", person.Tags)
	}
}

// 测试空对象和 nil
func TestNilAndEmpty(t *testing.T) {
	// 测试 nil
	jsonVal, err := StructToJSON(nil)
	if err != nil {
		t.Fatalf("StructToJSON(nil) failed: %v", err)
	}
	if jsonVal.Type != NULL {
		t.Errorf("Expected NULL for nil input, got %s", GetTypeStr(jsonVal))
	}

	// 测试空结构体
	type Empty struct{}
	empty := Empty{}
	jsonVal, err = StructToJSON(empty)
	if err != nil {
		t.Fatalf("StructToJSON(empty) failed: %v", err)
	}
	if jsonVal.Type != OBJECT {
		t.Errorf("Expected empty object, got %s", GetTypeStr(jsonVal))
	}
}

// 测试嵌套结构体
func TestNestedStruct(t *testing.T) {
	type Address struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	}

	type User struct {
		Name    string  `json:"name"`
		Address Address `json:"address"`
	}

	// 创建嵌套结构体
	user := User{
		Name: "Bob Johnson",
		Address: Address{
			Street:  "456 Oak Ave",
			City:    "San Francisco",
			Country: "USA",
		},
	}

	// 转换为 JSON
	jsonVal, err := StructToJSON(user)
	if err != nil {
		t.Fatalf("StructToJSON failed: %v", err)
	}

	// 验证基本字段
	nameValue := FindObjectValue(jsonVal, "name")
	if nameValue == nil || nameValue.Type != STRING || nameValue.S != "Bob Johnson" {
		t.Errorf("Name field not correctly serialized")
	}

	// 验证嵌套结构
	addressValue := FindObjectValue(jsonVal, "address")
	if addressValue == nil || addressValue.Type != OBJECT {
		t.Errorf("Address field not correctly serialized")
	}

	if addressValue != nil {
		streetValue := FindObjectValue(addressValue, "street")
		if streetValue == nil || streetValue.Type != STRING || streetValue.S != "456 Oak Ave" {
			t.Errorf("Address.Street not correctly serialized")
		}
	}

	// 将 JSON 转换为字符串并打印
	jsonStr, err := Stringify(jsonVal)
	if err != nil {
		t.Fatalf("Stringify failed: %v", err)
	}
	t.Logf("Nested struct JSON: %s", jsonStr)
}
