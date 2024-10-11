package timeplus

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var datePrefix = "date"
var dateTimePrefix = "datetime"
var dateTime64Prefix = "datetime64"

var dateLayout = "2006-01-02"
var dateTimeLayout = dateLayout + " 15:04:05"
var dateTime64Layout3 = dateTimeLayout + ".000"
var dateTime64Layout6 = dateTimeLayout + ".000000"

// this is from https://github.com/Vertamedia/clickhouse-grafana/blob/master/pkg/parser.go
type Value interface{}

func NewDataFieldByType(fieldName, fieldType string) *data.Field {
	if strings.HasPrefix(fieldType, "low_cardinality") {
		fieldType = strings.TrimSuffix(strings.TrimPrefix(fieldType, "low_cardinality("), ")")
	}

	isNullable := strings.Contains(fieldType, "nullable")
	fieldType = strings.TrimSuffix(strings.TrimPrefix(fieldType, "nullable("), ")")

	switch fieldType {
	case "string", "uuid", "ipv6", "ipv4":
		return newStringField(fieldName, isNullable)
	case "float32", "float64":
		//log.DefaultLogger.Info("[plugin.go] NewDateFieldByType", "fileldType", fieldType, "newFieldType", "float64")
		return newFloat64Field(fieldName, isNullable)
	case "uint64":
		// This can be a time or uint64 value
		// Assume that t is the field name used for timestamp
		if fieldName == "t" && !isNullable {
			return data.NewField(fieldName, nil, []time.Time{})
		}

		if isNullable {
			return data.NewField(fieldName, nil, []*uint64{})
		} else {
			return data.NewField(fieldName, nil, []uint64{})
		}
	case "uint8", "uint16", "uint32", "int8", "int16", "int32", "int64":
		if isNullable {
			return data.NewField(fieldName, nil, []*int64{})
		} else {
			return data.NewField(fieldName, nil, []int64{})
		}
	default:
		if strings.HasPrefix(fieldType, "decimal") {
			return newFloat64Field(fieldName, isNullable)
		} else if strings.HasPrefix(fieldType, "fixed_string") || strings.HasPrefix(fieldType, "enum") {
			return newStringField(fieldName, isNullable)
		} else if strings.HasPrefix(fieldType, dateTime64Prefix) || strings.HasPrefix(fieldType, dateTimePrefix) || strings.HasPrefix(fieldType, datePrefix) {
			return NewTimeField(fieldName, isNullable)
		} else {
			return newStringField(fieldName, isNullable)
		}
	}
}

func NewTimeField(fieldName string, isNullable bool) *data.Field {
	if isNullable {
		return data.NewField(fieldName, nil, []*time.Time{})
	} else {
		return data.NewField(fieldName, nil, []time.Time{})
	}
}

func newStringField(fieldName string, isNullable bool) *data.Field {
	if isNullable {
		return data.NewField(fieldName, nil, []*string{})
	} else {
		return data.NewField(fieldName, nil, []string{})
	}
}

func newFloat64Field(fieldName string, isNullable bool) *data.Field {
	if isNullable {
		return data.NewField(fieldName, nil, []*float64{})
	} else {
		return data.NewField(fieldName, nil, []float64{})
	}
}

func ParseValue(fieldName string, fieldType string, tz *time.Location, value interface{}, isNullable bool) Value {
	defer func() {
		if err := recover(); err != nil {
			log.DefaultLogger.Error("panic when paring value", "fieldName", fieldName, "fieldType", fieldType, "value", value, "valueType", fmt.Sprintf("%T", value), "isNullable", isNullable)
		}
	}()

	if strings.HasPrefix(fieldType, "nullable") {
		return ParseValue(fieldName, strings.TrimSuffix(strings.TrimPrefix(fieldType, "nullable("), ")"), tz, value, true)
	} else if strings.HasPrefix(fieldType, "low_cardinality") {
		return ParseValue(fieldName, strings.TrimSuffix(strings.TrimPrefix(fieldType, "low_cardinality("), ")"), tz, value, isNullable)
	} else {
		switch fieldType {
		case "string", "uuid", "ipv4", "ipv6":
			return parseStringValue(value, isNullable)
		case "float32", "float64":
			rv := parseFloatValue(value, isNullable)
			return rv
		case "uint8", "uint16", "uint32", "int8", "int16", "int32":
			rv := parseInt64Value(value, isNullable)
			return rv
		case "uint64":
			// Plugin specific corner case
			// This can be a time or uint64 value Assume that t is the field name used for timestamp in milliseconds
			if fieldName == "t" {
				return parseTimestampValue(value, isNullable)
			}
			rv := parseUInt64Value(value, isNullable)
			return rv
		case "int64":
			if fieldName == "t" {
				return parseTimestampValue(value, isNullable)
			}
			return parseInt64Value(value, isNullable)
		case "bool":
			return value
		default:
			if strings.HasPrefix(fieldType, "decimal") {
				return parseDecimalValue(value, isNullable)
			} else if strings.HasPrefix(fieldType, "fixed_string") || strings.HasPrefix(fieldType, "enum") {
				return parseStringValue(value, isNullable)
			} else if strings.HasPrefix(fieldType, dateTime64Prefix) && strings.Contains(fieldType, "3") {
				return parseDateTimeValue(value, dateTime64Layout3, tz, isNullable)
			} else if strings.HasPrefix(fieldType, dateTime64Prefix) && strings.Contains(fieldType, "6") {
				return parseDateTimeValue(value, dateTime64Layout6, tz, isNullable)
			} else if strings.HasPrefix(fieldType, dateTimePrefix) {
				return parseDateTimeValue(value, dateTimeLayout, tz, isNullable)
			} else if strings.HasPrefix(fieldType, datePrefix) {
				return parseDateTimeValue(value, dateLayout, tz, isNullable)
			} else {
				backend.Logger.Warn(fmt.Sprintf(
					"Value [%v] has compound type [%v] and will be returned as string", value, fieldType,
				))

				byteValue, err := json.Marshal(value)
				if err != nil {
					backend.Logger.Warn(fmt.Sprintf(
						"Unable to append value of unknown type %v because of json encoding problem: %s",
						reflect.TypeOf(value), err,
					))
					return nil
				}

				return parseStringValue(string(byteValue), isNullable)
			}
		}
	}
}

func parseDecimalValue(value interface{}, isNullable bool) Value {
	if value != nil {
		v := reflect.ValueOf(value)

		fv, err := strconv.ParseFloat(v.String(), 32)
		if err != nil {
			panic("failed to parse decimal")
		}

		if isNullable {
			return &fv
		} else {
			return fv
		}
	}

	if isNullable {
		return nil
	} else {
		return 0.0
	}
}

func parseFloatValue(value interface{}, isNullable bool) Value {
	if value != nil {
		fv := reflect.ValueOf(value).Float()
		if isNullable {
			return &fv
		} else {
			return fv
		}
	}

	if isNullable {
		return nil
	} else {
		return 0.0
	}
}

func parseStringValue(value interface{}, isNullable bool) Value {
	if value != nil {
		str := reflect.ValueOf(value).String()
		if isNullable {
			return &str
		} else {
			return str
		}
	}

	if isNullable {
		return nil
	} else {
		return ""
	}
}

func parseUInt64Value(value interface{}, isNullable bool) Value {
	if value != nil {
		ui64v, err := strconv.ParseUint(fmt.Sprintf("%v", value), 10, 64)

		if err == nil {
			if isNullable {
				return &ui64v
			} else {
				return ui64v
			}
		}
	}
	if isNullable {
		return nil
	} else {
		return uint64(0)
	}
}

func parseInt64Value(value interface{}, isNullable bool) Value {
	if value != nil {
		i64v, err := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)

		if err == nil {
			if isNullable {
				return &i64v
			} else {
				return i64v
			}
		}
	}

	if isNullable {
		return nil
	} else {
		return int64(0)
	}
}

func parseTimestampValue(value interface{}, isNullable bool) Value {
	if value != nil {
		strValue := fmt.Sprintf("%v", value)
		i64v, err := strconv.ParseInt(strValue, 10, 64)

		if err == nil {
			// Convert millisecond timestamp to nanosecond timestamp for parsing
			timeValue := time.Unix(0, i64v*int64(time.Millisecond))
			if isNullable {
				return &timeValue
			} else {
				return timeValue
			}
		}
	}

	if isNullable {
		return nil
	} else {
		return time.Unix(0, 0)
	}
}

func parseDateTimeValue(value interface{}, layout string, timezone *time.Location, isNullable bool) Value {
	if value != nil {
		t := value
		/*
				strValue := fmt.Sprintf("%v", value)
				t, err := time.ParseInLocation(layout, strValue, timezone)
				log.DefaultLogger.Info("parseDateTimeValue", "value", value, "strValue", strValue, "t", t)
			if err == nil {
		*/

		if isNullable {
			return &t
		} else {
			return t
		}
	}
	if isNullable {
		return nil
	} else {
		return time.Unix(0, 0)
	}
}
