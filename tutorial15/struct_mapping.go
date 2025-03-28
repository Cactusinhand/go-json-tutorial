// struct_mapping.go - 实现 Go 结构体与 JSON 的映射功能
package tutorial15

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// 定义特殊接口
type JSONMarshaler interface {
	MarshalJSON() ([]byte, error)
}

type JSONUnmarshaler interface {
	UnmarshalJSON([]byte) error
}

// StructToJSON 将 Go 结构体转换为 JSON 值
func StructToJSON(v interface{}) (*Value, error) {
	if v == nil {
		result := &Value{}
		SetNull(result)
		return result, nil
	}

	// 使用反射获取值
	val := reflect.ValueOf(v)

	// 处理指针
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			result := &Value{}
			SetNull(result)
			return result, nil
		}
		val = val.Elem()
	}

	// 检查自定义 JSON 编码
	if marshaler, ok := v.(JSONMarshaler); ok {
		data, err := marshaler.MarshalJSON()
		if err != nil {
			return nil, err
		}

		// 使用自己的解析而不是json.Unmarshal
		result := &Value{}
		err = Parse(result, string(data))
		if err != nil {
			return nil, fmt.Errorf("解析自定义JSON错误: %v", err)
		}
		return result, nil
	}

	// 根据类型进行处理
	switch val.Kind() {
	case reflect.Bool:
		result := &Value{}
		SetBool(result, val.Bool())
		return result, nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		result := &Value{}
		SetNumber(result, float64(val.Int()))
		return result, nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		result := &Value{}
		SetNumber(result, float64(val.Uint()))
		return result, nil

	case reflect.Float32, reflect.Float64:
		result := &Value{}
		SetNumber(result, val.Float())
		return result, nil

	case reflect.String:
		result := &Value{}
		SetString(result, val.String())
		return result, nil

	case reflect.Array, reflect.Slice:
		result := &Value{}
		SetArray(result)

		for i := 0; i < val.Len(); i++ {
			elemJSON, err := StructToJSON(val.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			PushBackArrayElement(result, elemJSON)
		}
		return result, nil

	case reflect.Map:
		// 只支持键为字符串的映射
		if val.Type().Key().Kind() != reflect.String {
			return nil, errors.New("map key must be string")
		}

		result := &Value{}
		SetObject(result)

		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key().String()
			v := iter.Value().Interface()

			valueJSON, err := StructToJSON(v)
			if err != nil {
				return nil, err
			}

			objVal := SetObjectValue(result, k)
			*objVal = *valueJSON
		}
		return result, nil

	case reflect.Struct:
		return structToJSON(val)

	default:
		return nil, fmt.Errorf("unsupported type: %v", val.Kind())
	}
}

// 处理结构体转 JSON
func structToJSON(val reflect.Value) (*Value, error) {
	result := &Value{}
	SetObject(result)

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 解析 json 标签
		tagValue := fieldType.Tag.Get("json")
		if tagValue == "-" {
			// 跳过被标记为忽略的字段
			continue
		}

		// 处理标签选项
		parts := strings.Split(tagValue, ",")
		name := fieldType.Name // 默认使用字段名
		omitEmpty := false

		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}

		for _, opt := range parts[1:] {
			if opt == "omitempty" {
				omitEmpty = true
			}
		}

		// 处理空值
		if omitEmpty {
			isZero := false
			switch field.Kind() {
			case reflect.Bool:
				isZero = !field.Bool()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				isZero = field.Int() == 0
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				isZero = field.Uint() == 0
			case reflect.Float32, reflect.Float64:
				isZero = field.Float() == 0
			case reflect.String:
				isZero = field.String() == ""
			case reflect.Map, reflect.Slice, reflect.Array:
				isZero = field.Len() == 0
			case reflect.Interface, reflect.Ptr:
				isZero = field.IsNil()
			}

			if isZero {
				continue
			}
		}

		// 递归转换字段值
		fieldJSON, err := StructToJSON(field.Interface())
		if err != nil {
			return nil, err
		}

		// 设置对象值
		objVal := SetObjectValue(result, name)
		*objVal = *fieldJSON
	}

	return result, nil
}

// JSONToStruct 将 JSON 值转换为 Go 结构体
func JSONToStruct(v *Value, target interface{}) error {
	if v == nil || target == nil {
		return errors.New("nil value or target")
	}

	// 获取目标类型和值
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}

	// 解引用
	targetValue = targetValue.Elem()

	// 检查自定义 JSON 解码
	if unmarshaler, ok := target.(JSONUnmarshaler); ok {
		jsonStr, err := Stringify(v)
		if err != nil {
			return err
		}
		return unmarshaler.UnmarshalJSON([]byte(jsonStr))
	}

	// 根据目标类型进行处理
	switch targetValue.Kind() {
	case reflect.Bool:
		if v.Type != BOOLEAN {
			return errors.New("value is not a boolean")
		}
		targetValue.SetBool(v.B)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Type != NUMBER {
			return errors.New("value is not a number")
		}
		targetValue.SetInt(int64(v.N))

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if v.Type != NUMBER {
			return errors.New("value is not a number")
		}
		targetValue.SetUint(uint64(v.N))

	case reflect.Float32, reflect.Float64:
		if v.Type != NUMBER {
			return errors.New("value is not a number")
		}
		targetValue.SetFloat(v.N)

	case reflect.String:
		if v.Type != STRING {
			return errors.New("value is not a string")
		}
		targetValue.SetString(v.S)

	case reflect.Array, reflect.Slice:
		if v.Type != ARRAY {
			return errors.New("value is not an array")
		}

		// 处理数组
		arrayLen := len(v.Elements)
		targetType := targetValue.Type()
		elemType := targetType.Elem()

		// 创建新的切片
		if targetValue.Kind() == reflect.Slice {
			targetValue.Set(reflect.MakeSlice(targetType, arrayLen, arrayLen))
		} else if targetValue.Len() < arrayLen {
			return errors.New("array length mismatch")
		}

		// 填充元素
		for i := 0; i < arrayLen; i++ {
			if i >= targetValue.Len() {
				break
			}

			elemValue := reflect.New(elemType).Elem()
			elemPtr := reflect.New(elemType)
			elemPtr.Elem().Set(elemValue)

			if err := JSONToStruct(v.Elements[i], elemPtr.Interface()); err != nil {
				return err
			}

			targetValue.Index(i).Set(elemPtr.Elem())
		}

	case reflect.Map:
		if v.Type != OBJECT {
			return errors.New("value is not an object")
		}

		// 确保键类型为字符串
		if targetValue.Type().Key().Kind() != reflect.String {
			return errors.New("map key must be string")
		}

		// 获取值类型
		valType := targetValue.Type().Elem()

		// 创建新映射
		if targetValue.IsNil() {
			targetValue.Set(reflect.MakeMap(targetValue.Type()))
		}

		// 遍历 JSON 对象成员
		for _, member := range v.Members {
			// 创建新的值
			newVal := reflect.New(valType).Elem()
			newValPtr := reflect.New(valType)
			newValPtr.Elem().Set(newVal)

			// 递归设置值
			if err := JSONToStruct(member.Value, newValPtr.Interface()); err != nil {
				return err
			}

			// 设置映射值
			targetValue.SetMapIndex(reflect.ValueOf(member.Key), newValPtr.Elem())
		}

	case reflect.Struct:
		if v.Type != OBJECT {
			return errors.New("value is not an object")
		}

		return jsonToStruct(v, targetValue)

	default:
		return fmt.Errorf("unsupported target type: %v", targetValue.Kind())
	}

	return nil
}

// 处理 JSON 转结构体
func jsonToStruct(v *Value, targetValue reflect.Value) error {
	typ := targetValue.Type()

	// 特殊处理 time.Time
	if typ == reflect.TypeOf(time.Time{}) && v.Type == STRING {
		t, err := time.Parse(time.RFC3339, v.S)
		if err != nil {
			return err
		}
		targetValue.Set(reflect.ValueOf(t))
		return nil
	}

	// 遍历结构体字段
	for i := 0; i < targetValue.NumField(); i++ {
		field := targetValue.Field(i)
		fieldType := typ.Field(i)

		// 跳过未导出字段
		if !fieldType.IsExported() {
			continue
		}

		// 解析 json 标签
		tagValue := fieldType.Tag.Get("json")
		if tagValue == "-" {
			// 跳过被标记为忽略的字段
			continue
		}

		// 处理标签名
		parts := strings.Split(tagValue, ",")
		name := fieldType.Name // 默认使用字段名

		if len(parts) > 0 && parts[0] != "" {
			name = parts[0]
		}

		// 查找 JSON 对象中的值
		jsonValue := FindObjectValue(v, name)
		if jsonValue == nil {
			continue // 字段在 JSON 中不存在
		}

		// 为字段创建新值
		fieldPtr := reflect.New(field.Type())

		// 递归填充字段值
		if err := JSONToStruct(jsonValue, fieldPtr.Interface()); err != nil {
			return err
		}

		// 设置字段值
		if field.CanSet() {
			field.Set(fieldPtr.Elem())
		}
	}

	return nil
}

// 时间类型的示例
type JSONTime time.Time

func (t JSONTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", time.Time(t).Format(time.RFC3339))), nil
}

func (t *JSONTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	s = strings.Trim(s, "\"")

	parsedTime, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*t = JSONTime(parsedTime)
	return nil
}
