package leptjson

import (
	"strings"
	"testing"
)

// 创建循环引用的测试数据
func createCyclicData() *Value {
	// 创建对象 {"a": {"b": <循环引用到根对象>}}
	root := &Value{}
	SetObject(root)

	a := SetObjectValue(root, "a")
	SetObject(a)

	// 直接设置b为指向root的引用，而不是复制
	// 注意: 这是手动创建循环引用，通常在实际代码中应该避免
	a.O = append(a.O, Member{K: "b", V: root})

	return root
}

// 创建自引用的测试数据
func createSelfReferencingData() *Value {
	// 创建对象 {"self": <引用到自身>}
	root := &Value{}
	SetObject(root)

	// 创建自引用
	root.O = append(root.O, Member{K: "self", V: root})

	return root
}

// 创建数组循环引用的测试数据
func createCyclicArrayData() *Value {
	// 创建数组 [1, [2, <循环引用到根数组>]]
	root := &Value{}
	SetArray(root, 2)

	// 添加第一个元素
	first := PushBackArrayElement(root)
	SetNumber(first, 1)

	// 添加第二个元素（数组）
	second := PushBackArrayElement(root)
	SetArray(second, 2)

	// 添加嵌套数组的元素
	nested1 := PushBackArrayElement(second)
	SetNumber(nested1, 2)

	// 添加循环引用元素 - 直接引用而不是复制
	// 这里我们直接将root添加为second的第二个元素
	second.A = append(second.A, root)

	return root
}

// 创建无循环引用的测试数据
func createNonCyclicData() *Value {
	// 创建复杂但没有循环引用的数据
	// {"a": {"b": 42}, "c": [1, {"d": "test"}]}
	root := &Value{}
	SetObject(root)

	a := SetObjectValue(root, "a")
	SetObject(a)

	b := SetObjectValue(a, "b")
	SetNumber(b, 42)

	c := SetObjectValue(root, "c")
	SetArray(c, 2)

	c1 := PushBackArrayElement(c)
	SetNumber(c1, 1)

	c2 := PushBackArrayElement(c)
	SetObject(c2)

	d := SetObjectValue(c2, "d")
	SetString(d, "test")

	return root
}

// 测试循环引用检测
func TestDetectCycle(t *testing.T) {
	tests := []struct {
		name      string
		data      func() *Value
		wantCycle bool
	}{
		{
			name:      "对象循环引用",
			data:      createCyclicData,
			wantCycle: true,
		},
		{
			name:      "自引用",
			data:      createSelfReferencingData,
			wantCycle: true,
		},
		{
			name:      "数组循环引用",
			data:      createCyclicArrayData,
			wantCycle: true,
		},
		{
			name:      "无循环引用",
			data:      createNonCyclicData,
			wantCycle: false,
		},
		{
			name: "简单值无循环引用",
			data: func() *Value {
				v := &Value{}
				SetNumber(v, 42)
				return v
			},
			wantCycle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := tt.data()
			result := DetectCycle(data)

			hasCycle := result == CYCLE_DETECTED
			if hasCycle != tt.wantCycle {
				t.Errorf("DetectCycle() = %v, want %v", hasCycle, tt.wantCycle)
			}
		})
	}
}

// 测试安全复制
func TestSafeCopy(t *testing.T) {
	tests := []struct {
		name    string
		data    func() *Value
		wantErr bool
	}{
		{
			name:    "对象循环引用",
			data:    createCyclicData,
			wantErr: true,
		},
		{
			name:    "无循环引用",
			data:    createNonCyclicData,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := tt.data()
			dst := &Value{}

			err := SafeCopy(dst, src)

			if (err != CYCLE_OK) != tt.wantErr {
				t.Errorf("SafeCopy() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// 确保复制正确
				if !Equal(dst, src) {
					t.Errorf("SafeCopy() 复制结果不相等")
				}
			}
		})
	}
}

// 测试带替换的安全复制
func TestCopySafeWithReplacement(t *testing.T) {
	t.Run("带循环引用的复制替换", func(t *testing.T) {
		src := createCyclicData()
		dst := &Value{}

		// 使用默认替换器进行复制
		CopySafeWithReplacement(dst, src)

		// 验证结果
		// 检查是否能获取 a.b 路径
		aValue, err := GetValueByPointer(dst, "/a")
		if err != nil {
			t.Fatalf("无法获取路径 /a: %v", err)
		}

		bValue, err := GetValueByPointer(aValue, "/b")
		if err != nil {
			t.Fatalf("无法获取路径 /a/b: %v", err)
		}

		// 检查 bValue 是否为替换的字符串
		if bValue.Type != STRING {
			t.Errorf("替换值类型应为STRING，得到 %v", bValue.Type)
		}

		s := GetString(bValue)
		if !strings.Contains(s, "循环引用") {
			t.Errorf("替换文本应包含'循环引用'，得到 %q", s)
		}
	})

	t.Run("自定义替换器", func(t *testing.T) {
		src := createCyclicData()
		dst := &Value{}

		// 自定义替换器
		replacer := func(path []string) *Value {
			v := &Value{}
			SetObject(v)
			typeValue := SetObjectValue(v, "type")
			SetString(typeValue, "cyclic")

			pathValue := SetObjectValue(v, "path")
			SetString(pathValue, "/"+strings.Join(path, "/"))

			return v
		}

		// 使用自定义替换器进行复制
		CustomCopySafeWithReplacement(dst, src, replacer)

		// 验证结果
		bValue, err := GetValueByPointer(dst, "/a/b")
		if err != nil {
			t.Fatalf("无法获取路径 /a/b: %v", err)
		}

		// 检查 bValue 是否为我们的自定义对象
		if bValue.Type != OBJECT {
			t.Errorf("替换值类型应为OBJECT，得到 %v", bValue.Type)
		}

		typeValue, err := GetValueByPointer(bValue, "/type")
		if err != nil || typeValue.Type != STRING || GetString(typeValue) != "cyclic" {
			t.Errorf("替换对象缺少或错误的type字段")
		}
	})
}
