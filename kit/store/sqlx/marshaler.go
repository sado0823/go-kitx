package sqlx

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

const tagName = "db"

func validPtr(v *reflect.Value) error {
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("invalid pointer type: %v", v)
	}

	return nil
}

func ofType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func ofValue(t reflect.Value) reflect.Value {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func unmarshalRows(v interface{}, rows *sql.Rows, strict bool) error {
	ofValueRaw := reflect.ValueOf(v)
	if err := validPtr(&ofValueRaw); err != nil {
		return err
	}

	ofTypeE := reflect.TypeOf(v).Elem()
	ofValueE := ofValueRaw.Elem()

	switch ofTypeE.Kind() {
	case reflect.Slice:
		if !ofValueE.CanSet() {
			return errors.WithMessagef(ErrNotSettable, "got:%s", ofTypeE.String())
		}

		// append ptr or not to slice
		isPtr := ofTypeE.Elem().Kind() == reflect.Ptr
		appendFn := func(v reflect.Value) {
			if isPtr {
				ofValueE.Set(reflect.Append(ofValueE, v))
				return
			}
			ofValueE.Set(reflect.Append(ofValueE, reflect.Indirect(v)))
		}

		base := ofType(ofTypeE.Elem())
		switch base.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			for rows.Next() {
				value := reflect.New(base)
				if err := rows.Scan(value.Interface()); err != nil {
					return err
				}
				appendFn(value)
			}
			return nil
		case reflect.Struct:
			columns, err := rows.Columns()
			if err != nil {
				return err
			}

			for rows.Next() {
				value := reflect.New(base)
				addrs, err := mappingStructFieldAddrs(value, columns, strict)
				if err != nil {
					return err
				}

				if err := rows.Scan(addrs...); err != nil {
					return err
				}

				appendFn(value)
			}
			return nil
		default:
			return errors.WithMessagef(ErrUnsupportedUnmarshalType, "need slice but got %s", base.Kind().String())
		}
	default:
		return errors.WithMessagef(ErrUnsupportedUnmarshalType, "need slice but got %s", ofTypeE.Kind().String())
	}
}

func unmarshalRow(v interface{}, rows *sql.Rows, strict bool) error {
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}

		return ErrNotFound
	}

	ofValueRaw := reflect.ValueOf(v)
	if err := validPtr(&ofValueRaw); err != nil {
		return err
	}

	ofTypeE := reflect.TypeOf(v).Elem()
	ofValueE := ofValueRaw.Elem()

	switch ofTypeE.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		if ofValueE.CanSet() {
			return rows.Scan(v)
		}

		return errors.WithMessagef(ErrNotSettable, "got:%s", ofTypeE.String())
	case reflect.Struct:
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		addrs, err := mappingStructFieldAddrs(ofValueE, columns, strict)
		if err != nil {
			return err
		}

		return rows.Scan(addrs...)
	default:
		return errors.WithMessagef(ErrUnsupportedUnmarshalType, "got:%s", ofTypeE.String())
	}
}

func mappingStructFieldAddrs(v reflect.Value, tagValues []string, strict bool) ([]interface{}, error) {
	fields := unwrapFields(v)
	lenTagValues := len(tagValues)
	lenFields := len(fields)

	if strict && lenTagValues != lenFields {
		return nil, errors.WithMessagef(ErrColumnsNotMatched, "got %d columns, but %d fields", lenTagValues, lenFields)
	}

	tagValuesAddr, err := mappingTagValuesAddr(v)
	if err != nil {
		return nil, err
	}

	fieldsAddr := make([]interface{}, lenTagValues)
	// match keys from struct field order
	if len(tagValuesAddr) == 0 {
		for i := 0; i < lenTagValues; i++ {
			valueField := fields[i]
			switch valueField.Kind() {
			case reflect.Ptr:
				if !valueField.CanInterface() {
					return nil, errors.WithMessagef(ErrNotReadable, "got:%s", valueField.Kind().String())
				}
				if valueField.IsNil() {
					valueField.Set(reflect.New(ofType(valueField.Type())))
				}
				fieldsAddr[i] = valueField.Interface()
			default:
				if !valueField.CanAddr() || !valueField.Addr().CanInterface() {
					return nil, errors.WithMessagef(ErrNotReadable, "got:%s", valueField.Kind().String())
				}
				fieldsAddr[i] = valueField.Addr().Interface()
			}
		}

		return fieldsAddr, nil
	}

	// match fields from struct tag name
	for i, tagValue := range tagValues {
		if addr, ok := tagValuesAddr[tagValue]; ok {
			fieldsAddr[i] = addr
		} else {
			// not strict mode, ignore this
			var anonymous interface{}
			fieldsAddr[i] = &anonymous
		}
	}

	return fieldsAddr, nil
}

func mappingTagValuesAddr(v reflect.Value) (map[string]interface{}, error) {
	ofTypeE := ofType(v.Type())
	size := ofTypeE.NumField()

	mapping := make(map[string]interface{}, size)
	for i := 0; i < size; i++ {
		name := parseTagName(ofTypeE.Field(i), tagName)
		if len(name) == 0 {
			return nil, nil
		}

		indirect := reflect.Indirect(v).Field(i)
		switch indirect.Kind() {
		case reflect.Ptr:
			if !indirect.CanInterface() {
				return nil, errors.WithMessagef(ErrNotReadable, "got:%s", indirect.Kind().String())
			}

			if indirect.IsNil() {
				indirect.Set(reflect.New(ofType(indirect.Type())))
			}

			mapping[name] = indirect.Interface()

		default:
			if !indirect.CanAddr() || !indirect.Addr().CanInterface() {
				return nil, errors.WithMessagef(ErrNotReadable, "got:%s", indirect.Kind().String())
			}

			mapping[name] = indirect.Addr().Interface()
		}
	}

	return mapping, nil
}

func parseTagName(field reflect.StructField, need string) string {
	key := field.Tag.Get(need)
	if len(key) == 0 {
		return ""
	}

	options := strings.Split(key, ",")
	return options[0]
}

func unwrapFields(v reflect.Value) []reflect.Value {
	var (
		fields   []reflect.Value
		indirect = reflect.Indirect(v)
	)

	for i := 0; i < indirect.NumField(); i++ {
		field := indirect.Field(i)

		// if nil ptr, set a new one
		if field.Kind() == reflect.Ptr && field.IsNil() {
			field.Set(reflect.New(ofType(field.Type())))
		}

		field = reflect.Indirect(field)
		fieldType := indirect.Type().Field(i)
		if field.Kind() == reflect.Struct && fieldType.Anonymous {
			fields = append(fields, unwrapFields(field)...)
		} else {
			fields = append(fields, field)
		}
	}

	return fields
}
