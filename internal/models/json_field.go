package models

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONField 通用JSON字段类型
type JSONField map[string]interface{}

// Scan 实现 sql.Scanner 接口
func (jf *JSONField) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, jf)
}

// Value 实现 driver.Valuer 接口
func (jf JSONField) Value() (driver.Value, error) {
	return json.Marshal(jf)
}

// JSONArray 通用JSON数组类型
type JSONArray []interface{}

// Scan 实现 sql.Scanner 接口
func (ja *JSONArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, ja)
}

// Value 实现 driver.Valuer 接口
func (ja JSONArray) Value() (driver.Value, error) {
	return json.Marshal(ja)
}

// StringArray 字符串数组类型
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sa)
}

// Value 实现 driver.Valuer 接口
func (sa StringArray) Value() (driver.Value, error) {
	return json.Marshal(sa)
}