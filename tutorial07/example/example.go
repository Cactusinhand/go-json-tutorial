package main

import (
	"fmt"
	leptjson "github.com/Cactusinhand/go-json-tutorial/tutorial07"
)

func main() {
	// 创建一个复杂的JSON值
	obj := &leptjson.Value{
		Type: leptjson.OBJECT,
		O: []leptjson.Member{
			{
				K: "name",
				V: &leptjson.Value{Type: leptjson.STRING, S: "张三"},
			},
			{
				K: "age",
				V: &leptjson.Value{Type: leptjson.NUMBER, N: 30},
			},
			{
				K: "is_student",
				V: &leptjson.Value{Type: leptjson.FALSE},
			},
			{
				K: "hobbies",
				V: &leptjson.Value{
					Type: leptjson.ARRAY,
					A: []*leptjson.Value{
						{Type: leptjson.STRING, S: "读书"},
						{Type: leptjson.STRING, S: "写作"},
						{Type: leptjson.STRING, S: "编程"},
					},
				},
			},
			{
				K: "address",
				V: &leptjson.Value{
					Type: leptjson.OBJECT,
					O: []leptjson.Member{
						{
							K: "city",
							V: &leptjson.Value{Type: leptjson.STRING, S: "北京"},
						},
						{
							K: "street",
							V: &leptjson.Value{Type: leptjson.STRING, S: "朝阳区"},
						},
					},
				},
			},
		},
	}

	// 生成JSON字符串
	json, err := leptjson.Stringify(obj)
	if err != leptjson.STRINGIFY_OK {
		fmt.Println("生成JSON出错:", err)
		return
	}

	fmt.Println("生成的JSON:")
	fmt.Println(json)

	// 解析JSON字符串
	var parsedValue leptjson.Value
	if err := leptjson.Parse(&parsedValue, json); err != leptjson.PARSE_OK {
		fmt.Println("解析JSON出错:", err)
		return
	}

	// 获取值
	fmt.Println("\n从解析的JSON中获取值:")
	fmt.Println("姓名:", leptjson.GetString(leptjson.GetObjectValueByKey(&parsedValue, "name")))
	fmt.Println("年龄:", leptjson.GetNumber(leptjson.GetObjectValueByKey(&parsedValue, "age")))
	fmt.Println("是否学生:", leptjson.GetType(leptjson.GetObjectValueByKey(&parsedValue, "is_student")) == leptjson.FALSE)

	hobbies := leptjson.GetObjectValueByKey(&parsedValue, "hobbies")
	fmt.Println("爱好数量:", leptjson.GetArraySize(hobbies))
	for i := 0; i < leptjson.GetArraySize(hobbies); i++ {
		fmt.Printf("爱好 %d: %s\n", i+1, leptjson.GetString(leptjson.GetArrayElement(hobbies, i)))
	}

	address := leptjson.GetObjectValueByKey(&parsedValue, "address")
	fmt.Println("城市:", leptjson.GetString(leptjson.GetObjectValueByKey(address, "city")))
	fmt.Println("街道:", leptjson.GetString(leptjson.GetObjectValueByKey(address, "street")))
}
