package module

import (
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

func DatabaseModelToStruct(src interface{}, to interface{}) error {
	srcMap := make(map[string]interface{})
	{
		data, err := json.Marshal(src)
		if err != nil {
			return fmt.Errorf("struct convert failed err=%s\n", err)
		}

		if err := json.Unmarshal(data, &srcMap); err != nil {
			return err
		}
	}

	srcV := reflect.ValueOf(src).Elem()
	toT := reflect.TypeOf(to).Elem()
	toV := reflect.ValueOf(to).Elem()

	for i := 0; i < toT.NumField(); i++ {
		fieldName := toT.Field(i).Name
		srcField := srcV.FieldByName(fieldName)
		toField := toV.FieldByName(fieldName)

		//当src不存在字段，跳过
		if srcField.IsValid() == false {
			continue
		}

		//当src字段为string，目标to字段不为string，使用json解析
		//当to字段为string，src字段不为string，使用json压缩
		t := valueIsString(toField)
		s := valueIsString(srcField)

		if s == true && t == false {
			var str string
			switch srcMap[fieldName].(type) {
			case string:
				str = srcMap[fieldName].(string)
				break
			case *string:
				str = *srcMap[fieldName].(*string)
				break
			default:
				return fmt.Errorf("reflect srcMap field name=%s to string failed", fieldName)
			}

			temp := reflect.New(toField.Type()).Interface()
			err := jsoniter.UnmarshalFromString(str, temp)
			if err != nil {
				return fmt.Errorf("database model %s to struct %s : field name=%s unmarshal failed err=%s", srcV.Type(), toV.Type(), fieldName, err)
			}
			srcMap[fieldName] = temp
			continue
		}
	}

	{
		data, err := jsoniter.Marshal(&srcMap)
		if err != nil {
			return err
		}
		if err := jsoniter.Unmarshal(data, to); err != nil {
			return fmt.Errorf("database model to struct failed err=%s", err)
		}
	}

	return nil
}

func StructToDatabaseModel(src interface{}, to interface{}) error {
	toMap := make(map[string]interface{})
	{
		data, err := jsoniter.Marshal(src)
		if err != nil {
			return fmt.Errorf("struct convert failed err=%s\n", err)
		}
		if err := jsoniter.Unmarshal(data, &toMap); err != nil {
			fmt.Printf("StructToDatabaseModel json.Unmarshal 出错， err: %v \n", err)
			return err
		}
	}

	var srcT reflect.Type
	if reflect.ValueOf(src).Kind() == reflect.Ptr {
		srcT = reflect.TypeOf(src).Elem()
	} else {
		srcT = reflect.TypeOf(src)
	}
	var srcV reflect.Value
	if reflect.ValueOf(src).Kind() == reflect.Ptr {
		srcV = reflect.ValueOf(src).Elem()
	} else {
		srcV = reflect.ValueOf(src)
	}

	toV := reflect.ValueOf(to)
	if toV.Kind() == reflect.Ptr {
		toV = toV.Elem()
	} else {
		return fmt.Errorf("the params 'to' is not a ptr")
	}

	for i := 0; i < srcT.NumField(); i++ {
		fieldName := srcT.Field(i).Name
		srcField := srcV.FieldByName(fieldName)
		toField := toV.FieldByName(fieldName)

		//当to不存在字段，跳过
		if toField.IsValid() == false {
			continue
		}

		//当to字段为string，src字段不为string，使用json压缩
		t := valueIsString(toField)
		s := valueIsString(srcField)

		if t == true && s == false {
			str, err := jsoniter.MarshalToString(srcField.Interface())
			if err != nil {
				return fmt.Errorf("marshal to string field name=%s failed", fieldName)
			}
			toMap[fieldName] = str
		}
	}

	{
		data, err := jsoniter.Marshal(&toMap)
		if err != nil {
			return err
		}
		if err := jsoniter.Unmarshal(data, to); err != nil {
			return fmt.Errorf("struct to model unmarshal faield err=%s", err)
		}
	}

	return nil
}

func valueIsString(v reflect.Value) bool {
	switch v.Interface().(type) {
	case string:
		return true
	case *string:
		return true
	}
	return false
}
