// json_schema.go - JSON Schema 验证实现（基于部分 JSON Schema Draft 7）
package leptjson

import (
	"fmt"
	"math"
	"regexp"
)

// SchemaValidationError 表示 JSON Schema 验证错误
type SchemaValidationError struct {
	Path    string // 导致错误的 JSON 路径
	Message string // 错误描述
}

// 实现 Error 接口
func (e SchemaValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("位于 '%s': %s", e.Path, e.Message)
}

// SchemaValidationResult 存储验证结果
type SchemaValidationResult struct {
	Valid  bool                    // 是否验证通过
	Errors []SchemaValidationError // 验证错误列表
}

// AddError 添加验证错误
func (r *SchemaValidationResult) AddError(path, message string) {
	r.Valid = false
	r.Errors = append(r.Errors, SchemaValidationError{Path: path, Message: message})
}

// JSONSchema 表示一个 JSON Schema 对象
type JSONSchema struct {
	Schema *Value // 存储 JSON Schema 的 Value 对象
}

// NewJSONSchema 创建一个新的 JSON Schema
func NewJSONSchema(schemaJSON string) (*JSONSchema, error) {
	schema := &Value{}
	if err := Parse(schema, schemaJSON); err != PARSE_OK {
		return nil, fmt.Errorf("无法解析 JSON Schema: %v", err)
	}

	// 验证输入是否是一个对象
	if schema.Type != OBJECT {
		return nil, fmt.Errorf("JSON Schema 必须是一个对象")
	}

	return &JSONSchema{Schema: schema}, nil
}

// NewJSONSchemaFromValue 从 Value 对象创建 JSON Schema
func NewJSONSchemaFromValue(schema *Value) (*JSONSchema, error) {
	if schema.Type != OBJECT {
		return nil, fmt.Errorf("JSON Schema 必须是一个对象")
	}

	return &JSONSchema{Schema: schema}, nil
}

// Validate 根据 Schema 验证 JSON 数据
func (js *JSONSchema) Validate(data *Value) *SchemaValidationResult {
	result := &SchemaValidationResult{Valid: true}
	js.validateValue(js.Schema, data, "", result)
	return result
}

// validateValue 是验证的核心递归函数
func (js *JSONSchema) validateValue(schema, data *Value, path string, result *SchemaValidationResult) {
	// 类型验证
	if typeValue, found := FindObjectKey(schema, "type"); found {
		js.validateType(typeValue, data, path, result)
	}

	// 根据数据类型执行不同的验证
	switch data.Type {
	case NUMBER:
		js.validateNumber(schema, data, path, result)
	case STRING:
		js.validateString(schema, data, path, result)
	case ARRAY:
		js.validateArray(schema, data, path, result)
	case OBJECT:
		js.validateObject(schema, data, path, result)
	}

	// 枚举验证
	if enumValue, found := FindObjectKey(schema, "enum"); found && enumValue.Type == ARRAY {
		matched := false
		for i := 0; i < len(enumValue.A); i++ {
			if Equal(data, enumValue.A[i]) {
				matched = true
				break
			}
		}
		if !matched {
			result.AddError(path, "值不在枚举列表中")
		}
	}

	// 常量验证
	if constValue, found := FindObjectKey(schema, "const"); found {
		if !Equal(data, constValue) {
			result.AddError(path, "值不等于常量")
		}
	}

	// allOf 验证 (所有子模式都要匹配)
	if allOf, found := FindObjectKey(schema, "allOf"); found && allOf.Type == ARRAY {
		for i := 0; i < len(allOf.A); i++ {
			subSchema := allOf.A[i]
			if subSchema.Type == OBJECT {
				js.validateValue(subSchema, data, path, result)
			} else {
				result.AddError(path, "allOf 中的子模式必须是对象")
			}
		}
	}

	// anyOf 验证 (至少一个子模式要匹配)
	if anyOf, found := FindObjectKey(schema, "anyOf"); found && anyOf.Type == ARRAY {
		if len(anyOf.A) == 0 {
			result.AddError(path, "anyOf 不能为空数组")
			return
		}

		anyMatched := false

		for i := 0; i < len(anyOf.A); i++ {
			subSchema := anyOf.A[i]
			if subSchema.Type != OBJECT {
				continue
			}

			// 为每个子模式创建新的结果
			tempResult := &SchemaValidationResult{Valid: true}
			js.validateValue(subSchema, data, path, tempResult)

			if tempResult.Valid {
				anyMatched = true
				break
			}
		}

		if !anyMatched {
			result.AddError(path, "值不符合 anyOf 中的任何模式")
		}
	}

	// oneOf 验证 (恰好一个子模式要匹配)
	if oneOf, found := FindObjectKey(schema, "oneOf"); found && oneOf.Type == ARRAY {
		if len(oneOf.A) == 0 {
			result.AddError(path, "oneOf 不能为空数组")
			return
		}

		matchCount := 0

		for i := 0; i < len(oneOf.A); i++ {
			subSchema := oneOf.A[i]
			if subSchema.Type != OBJECT {
				continue
			}

			tempResult := &SchemaValidationResult{Valid: true}
			js.validateValue(subSchema, data, path, tempResult)

			if tempResult.Valid {
				matchCount++
			}
		}

		if matchCount != 1 {
			result.AddError(path, fmt.Sprintf("值匹配了 %d 个 oneOf 模式，但应该恰好匹配 1 个", matchCount))
		}
	}

	// not 验证 (子模式不能匹配)
	if notSchema, found := FindObjectKey(schema, "not"); found && notSchema.Type == OBJECT {
		tempResult := &SchemaValidationResult{Valid: true}
		js.validateValue(notSchema, data, path, tempResult)

		if tempResult.Valid {
			result.AddError(path, "值不应该匹配 not 模式")
		}
	}
}

// validateType 验证值的类型
func (js *JSONSchema) validateType(typeSchema, data *Value, path string, result *SchemaValidationResult) {
	// 处理单一类型
	if typeSchema.Type == STRING {
		typeName := GetString(typeSchema)
		if !js.matchType(typeName, data) {
			result.AddError(path, fmt.Sprintf("类型不匹配，期望 '%s' 类型，得到 '%s' 类型", typeName, getTypeName(data)))
		}
		return
	}

	// 处理类型数组（多个可能的类型）
	if typeSchema.Type == ARRAY {
		// 检查数组中是否有任何类型匹配
		matched := false
		for i := 0; i < len(typeSchema.A); i++ {
			typeValue := typeSchema.A[i]
			if typeValue.Type == STRING {
				typeName := GetString(typeValue)
				if js.matchType(typeName, data) {
					matched = true
					break
				}
			}
		}

		if !matched {
			result.AddError(path, fmt.Sprintf("类型不匹配，值类型 '%s' 不在允许的类型列表中", getTypeName(data)))
		}
	}
}

// matchType 检查值是否匹配指定的类型名
func (js *JSONSchema) matchType(typeName string, value *Value) bool {
	switch typeName {
	case "null":
		return value.Type == NULL
	case "boolean":
		return value.Type == TRUE || value.Type == FALSE
	case "number":
		return value.Type == NUMBER
	case "integer":
		return value.Type == NUMBER && math.Floor(GetNumber(value)) == GetNumber(value)
	case "string":
		return value.Type == STRING
	case "array":
		return value.Type == ARRAY
	case "object":
		return value.Type == OBJECT
	default:
		return false
	}
}

// getTypeName 返回值的类型名称
func getTypeName(value *Value) string {
	switch value.Type {
	case NULL:
		return "null"
	case TRUE, FALSE:
		return "boolean"
	case NUMBER:
		// 检查是否为整数
		if math.Floor(GetNumber(value)) == GetNumber(value) {
			return "integer"
		}
		return "number"
	case STRING:
		return "string"
	case ARRAY:
		return "array"
	case OBJECT:
		return "object"
	default:
		return "unknown"
	}
}

// validateNumber 验证数字
func (js *JSONSchema) validateNumber(schema, data *Value, path string, result *SchemaValidationResult) {
	num := GetNumber(data)

	// 最小值验证
	if minValue, found := FindObjectKey(schema, "minimum"); found && minValue.Type == NUMBER {
		min := GetNumber(minValue)
		if num < min {
			result.AddError(path, fmt.Sprintf("值 %g 小于最小值 %g", num, min))
		}
	}

	// 独占最小值验证
	if minValue, found := FindObjectKey(schema, "exclusiveMinimum"); found && minValue.Type == NUMBER {
		min := GetNumber(minValue)
		if num <= min {
			result.AddError(path, fmt.Sprintf("值 %g 应该大于独占最小值 %g", num, min))
		}
	}

	// 最大值验证
	if maxValue, found := FindObjectKey(schema, "maximum"); found && maxValue.Type == NUMBER {
		max := GetNumber(maxValue)
		if num > max {
			result.AddError(path, fmt.Sprintf("值 %g 大于最大值 %g", num, max))
		}
	}

	// 独占最大值验证
	if maxValue, found := FindObjectKey(schema, "exclusiveMaximum"); found && maxValue.Type == NUMBER {
		max := GetNumber(maxValue)
		if num >= max {
			result.AddError(path, fmt.Sprintf("值 %g 应该小于独占最大值 %g", num, max))
		}
	}

	// 倍数验证
	if multipleOf, found := FindObjectKey(schema, "multipleOf"); found && multipleOf.Type == NUMBER {
		divisor := GetNumber(multipleOf)
		if divisor <= 0 {
			result.AddError(path, "multipleOf 必须为正数")
		} else {
			// 检查是否为倍数（考虑浮点数精度问题）
			quotient := num / divisor
			if math.Abs(math.Round(quotient)-quotient) > 1e-10 {
				result.AddError(path, fmt.Sprintf("值 %g 不是 %g 的倍数", num, divisor))
			}
		}
	}
}

// validateString 验证字符串
func (js *JSONSchema) validateString(schema, data *Value, path string, result *SchemaValidationResult) {
	str := GetString(data)

	// 最小长度验证
	if minLength, found := FindObjectKey(schema, "minLength"); found && minLength.Type == NUMBER {
		min := int(GetNumber(minLength))
		if min < 0 {
			result.AddError(path, "minLength 必须为非负整数")
		} else if len([]rune(str)) < min { // 使用 []rune 计算 Unicode 字符数
			result.AddError(path, fmt.Sprintf("字符串长度 %d 小于最小长度 %d", len([]rune(str)), min))
		}
	}

	// 最大长度验证
	if maxLength, found := FindObjectKey(schema, "maxLength"); found && maxLength.Type == NUMBER {
		max := int(GetNumber(maxLength))
		if max < 0 {
			result.AddError(path, "maxLength 必须为非负整数")
		} else if len([]rune(str)) > max {
			result.AddError(path, fmt.Sprintf("字符串长度 %d 大于最大长度 %d", len([]rune(str)), max))
		}
	}

	// 模式验证 (正则表达式)
	if pattern, found := FindObjectKey(schema, "pattern"); found && pattern.Type == STRING {
		patternStr := GetString(pattern)
		re, err := regexp.Compile(patternStr)
		if err != nil {
			result.AddError(path, fmt.Sprintf("无效的正则表达式模式: %s", patternStr))
		} else if !re.MatchString(str) {
			result.AddError(path, fmt.Sprintf("字符串不匹配模式: %s", patternStr))
		}
	}

	// format 验证 (部分常见格式)
	if format, found := FindObjectKey(schema, "format"); found && format.Type == STRING {
		formatName := GetString(format)
		switch formatName {
		case "email":
			// 简单的电子邮件验证
			emailRegex := regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)
			if !emailRegex.MatchString(str) {
				result.AddError(path, "字符串不是有效的电子邮件格式")
			}
		case "date-time":
			// 简化的 ISO8601 日期时间验证
			dateTimeRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(Z|[+-]\d{2}:\d{2})$`)
			if !dateTimeRegex.MatchString(str) {
				result.AddError(path, "字符串不是有效的 ISO8601 日期时间格式")
			}
		case "uri":
			// 简单的 URI 验证
			uriRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:.*$`)
			if !uriRegex.MatchString(str) {
				result.AddError(path, "字符串不是有效的 URI 格式")
			}
		}
		// 其他格式可以根据需要添加
	}
}

// validateArray 验证数组
func (js *JSONSchema) validateArray(schema, data *Value, path string, result *SchemaValidationResult) {
	arrLen := len(data.A)

	// 最小元素数量验证
	if minItems, found := FindObjectKey(schema, "minItems"); found && minItems.Type == NUMBER {
		min := int(GetNumber(minItems))
		if min < 0 {
			result.AddError(path, "minItems 必须为非负整数")
		} else if arrLen < min {
			result.AddError(path, fmt.Sprintf("数组长度小于最小长度 %d", min))
		}
	}

	// 最大元素数量验证
	if maxItems, found := FindObjectKey(schema, "maxItems"); found && maxItems.Type == NUMBER {
		max := int(GetNumber(maxItems))
		if max < 0 {
			result.AddError(path, "maxItems 必须为非负整数")
		} else if arrLen > max {
			result.AddError(path, fmt.Sprintf("数组长度大于最大长度 %d", max))
		}
	}

	// 元素唯一性验证
	if uniqueItems, found := FindObjectKey(schema, "uniqueItems"); found && (uniqueItems.Type == TRUE || uniqueItems.Type == FALSE) {
		if uniqueItems.Type == TRUE && arrLen > 1 {
			// 检查数组元素的唯一性
			for i := 0; i < arrLen-1; i++ {
				for j := i + 1; j < arrLen; j++ {
					if Equal(data.A[i], data.A[j]) {
						result.AddError(path, fmt.Sprintf("数组包含重复项（在索引 %d 和 %d 处）", i, j))
						break
					}
				}
			}
		}
	}

	// items 验证 - 单一模式
	if items, found := FindObjectKey(schema, "items"); found && items.Type == OBJECT {
		// 验证数组中的每个元素
		for i := 0; i < arrLen; i++ {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			js.validateValue(items, data.A[i], elemPath, result)
		}
	}

	// items 验证 - 模式数组
	if items, found := FindObjectKey(schema, "items"); found && items.Type == ARRAY {
		itemsArr := items.A
		itemsLen := len(itemsArr)

		// 验证每个元素与对应的模式
		for i := 0; i < arrLen && i < itemsLen; i++ {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			js.validateValue(itemsArr[i], data.A[i], elemPath, result)
		}

		// 处理额外的元素
		if additionalItems, found := FindObjectKey(schema, "additionalItems"); found {
			// 如果 additionalItems 是 false，且数组长度超过了 items 中的模式数量
			if additionalItems.Type == FALSE && arrLen > itemsLen {
				result.AddError(path, fmt.Sprintf("数组包含 %d 个元素，但只允许 %d 个元素", arrLen, itemsLen))
			}

			// 如果 additionalItems 是对象，验证额外的元素
			if additionalItems.Type == OBJECT {
				for i := itemsLen; i < arrLen; i++ {
					elemPath := fmt.Sprintf("%s[%d]", path, i)
					js.validateValue(additionalItems, data.A[i], elemPath, result)
				}
			}
		}
	}

	// contains 验证 (至少有一个元素匹配指定的模式)
	if contains, found := FindObjectKey(schema, "contains"); found && contains.Type == OBJECT {
		if arrLen == 0 {
			result.AddError(path, "空数组不满足 contains 要求")
			return
		}

		matched := false
		for i := 0; i < arrLen; i++ {
			tempResult := &SchemaValidationResult{Valid: true}
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			js.validateValue(contains, data.A[i], elemPath, tempResult)

			if tempResult.Valid {
				matched = true
				break
			}
		}

		if !matched {
			result.AddError(path, "数组中没有元素满足 contains 模式")
		}
	}
}

// validateObject 验证对象
func (js *JSONSchema) validateObject(schema, data *Value, path string, result *SchemaValidationResult) {
	objSize := len(data.O)

	// 最小属性数验证
	if minProps, found := FindObjectKey(schema, "minProperties"); found && minProps.Type == NUMBER {
		min := int(GetNumber(minProps))
		if min < 0 {
			result.AddError(path, "minProperties 必须为非负整数")
		} else if objSize < min {
			result.AddError(path, fmt.Sprintf("对象属性数量 %d 小于最小数量 %d", objSize, min))
		}
	}

	// 最大属性数验证
	if maxProps, found := FindObjectKey(schema, "maxProperties"); found && maxProps.Type == NUMBER {
		max := int(GetNumber(maxProps))
		if max < 0 {
			result.AddError(path, "maxProperties 必须为非负整数")
		} else if objSize > max {
			result.AddError(path, fmt.Sprintf("对象属性数量 %d 大于最大数量 %d", objSize, max))
		}
	}

	// 获取数据对象的所有属性名
	properties := make(map[string]bool)
	for _, member := range data.O {
		properties[member.K] = true
	}

	// required 验证 (必需属性)
	if required, found := FindObjectKey(schema, "required"); found && required.Type == ARRAY {
		for i := 0; i < len(required.A); i++ {
			reqProp := required.A[i]
			if reqProp.Type == STRING {
				propName := GetString(reqProp)
				if _, exists := properties[propName]; !exists {
					result.AddError(path, fmt.Sprintf("缺少必需属性 '%s'", propName))
				}
			}
		}
	}

	// 属性验证
	schemaProps, hasProps := FindObjectKey(schema, "properties")
	patternProps, hasPatternProps := FindObjectKey(schema, "patternProperties")
	additionalProps, hasAdditionalProps := FindObjectKey(schema, "additionalProperties")

	// 跟踪已验证的属性
	validatedProps := make(map[string]bool)

	// 通过 properties 验证
	if hasProps && schemaProps.Type == OBJECT {
		for _, member := range data.O {
			propName := member.K
			if propSchema, found := FindObjectKey(schemaProps, propName); found {
				propPath := path
				if path != "" {
					propPath += "."
				}
				propPath += propName

				js.validateValue(propSchema, member.V, propPath, result)
				validatedProps[propName] = true
			}
		}
	}

	// 通过 patternProperties 验证
	if hasPatternProps && patternProps.Type == OBJECT {
		for _, member := range data.O {
			propName := member.K

			// 检查属性名是否匹配任何模式
			for _, patternProp := range patternProps.O {
				pattern := patternProp.K
				re, err := regexp.Compile(pattern)
				if err != nil {
					result.AddError(path, fmt.Sprintf("无效的属性名模式: %s", pattern))
					continue
				}

				if re.MatchString(propName) {
					propPath := path
					if path != "" {
						propPath += "."
					}
					propPath += propName

					js.validateValue(patternProp.V, member.V, propPath, result)
					validatedProps[propName] = true
				}
			}
		}
	}

	// 验证额外属性
	if hasAdditionalProps {
		// 如果 additionalProperties 是 false，则不允许未经验证的属性
		if additionalProps.Type == FALSE {
			for _, member := range data.O {
				propName := member.K
				if !validatedProps[propName] {
					result.AddError(path, fmt.Sprintf("属性 '%s' 不允许存在", propName))
				}
			}
		}

		// 如果 additionalProperties 是对象，使用它验证未经验证的属性
		if additionalProps.Type == OBJECT {
			for _, member := range data.O {
				propName := member.K
				if !validatedProps[propName] {
					propPath := path
					if path != "" {
						propPath += "."
					}
					propPath += propName

					js.validateValue(additionalProps, member.V, propPath, result)
				}
			}
		}
	}

	// propertyNames 验证 (属性名的模式)
	if propNames, found := FindObjectKey(schema, "propertyNames"); found && propNames.Type == OBJECT {
		// 创建一个只有 type 和 pattern 的临时 schema
		tempSchema := &Value{}
		SetObject(tempSchema)

		if pattern, found := FindObjectKey(propNames, "pattern"); found && pattern.Type == STRING {
			patternValue := SetObjectValue(tempSchema, "pattern")
			Copy(patternValue, pattern)
		}

		// 验证每个属性名
		for _, member := range data.O {
			propName := member.K

			// 创建临时的字符串值用于验证
			tempValue := &Value{}
			SetString(tempValue, propName)

			propPath := path
			if path != "" {
				propPath += "."
			}
			propPath += "[property name '" + propName + "']"

			js.validateString(propNames, tempValue, propPath, result)
		}
	}

	// dependencies 验证 (属性依赖)
	if deps, found := FindObjectKey(schema, "dependencies"); found && deps.Type == OBJECT {
		for _, dep := range deps.O {
			propName := dep.K
			// 只有当属性存在时才验证依赖
			if _, exists := properties[propName]; exists {
				// 属性依赖 (数组依赖)
				if dep.V.Type == ARRAY {
					for i := 0; i < len(dep.V.A); i++ {
						depProp := dep.V.A[i]
						if depProp.Type == STRING {
							depPropName := GetString(depProp)
							if _, depExists := properties[depPropName]; !depExists {
								result.AddError(path, fmt.Sprintf("属性 '%s' 依赖于属性 '%s'，但后者不存在", propName, depPropName))
							}
						}
					}
				}

				// schema 依赖
				if dep.V.Type == OBJECT {
					js.validateValue(dep.V, data, path, result)
				}
			}
		}
	}
}
