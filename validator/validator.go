package validator

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func ShouldBindWithBody(s any, c *gin.Context) error {
	if err := c.ShouldBindWith(s, binding.JSON); err != nil {
		errMsg := ""
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if ok {
			for _, e := range errs {
				nameSpace := e.Namespace()
				ns := strings.Split(nameSpace, ".")
				var field reflect.StructField
				t := reflect.TypeOf(s).Elem()
				v := reflect.ValueOf(s).Elem()
				field, _ = t.FieldByName(ns[1])
				fieldVal := v.FieldByName(ns[1])
				if len(ns) > 2 {
					for i := 2; i < len(ns); i++ {
						fn := ns[i]
						if fieldVal.Kind() == reflect.Ptr {
							field, _ = reflect.TypeOf(fieldVal.Interface()).Elem().FieldByName(fn)
						} else {
							field, _ = reflect.TypeOf(fieldVal.Interface()).FieldByName(fn)
						}
					}
				}
				name := field.Tag.Get("label")
				typeName := field.Type.String()
				if strings.TrimSpace(name) == "" {
					name = field.Tag.Get("json")
					if strings.TrimSpace(name) == "" {
						name = e.Field()
					}
				}
				if "required" == e.Tag() {
					requiredMsg := field.Tag.Get("requiredMsg")
					if strings.TrimSpace(requiredMsg) != "" {
						errMsg = requiredMsg
					} else {
						errMsg = "参数" + name + "必填"
					}
					return &BadRequestError{Msg: errMsg}
				} else if "regexp" == e.Tag() {
					// TODO 正则表达式？？？
				} else if "min" == e.Tag() {
					if "string" == typeName || "*string" == typeName {
						return &BadRequestError{Msg: "参数" + name + "长度过短"}
					} else {
						return &BadRequestError{Msg: "参数" + name + "不能小于最小值"}
					}
				} else if "max" == e.Tag() {
					if "string" == typeName || "*string" == typeName {
						return &BadRequestError{Msg: "参数" + name + "长度超长"}
					} else {
						return &BadRequestError{Msg: "参数" + name + "不能大于最大值"}
					}
				}
			}
		} else {
			errMsg = err.Error()
		}
		return &BadRequestError{Msg: errMsg}
	}
	return nil
}
func ShouldBindWithQuery(s any, c *gin.Context) error {
	return shouldBindWith(s, c, "Query")
}

func ShouldBindWithForm(s any, c *gin.Context) error {
	return shouldBindWith(s, c, "Form")
}

func shouldBindWith(s any, c *gin.Context, paramType string) error {
	t := reflect.TypeOf(s).Elem()
	m := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fileType := field.Type.String()
		fieldName := field.Tag.Get("json")
		if "" == fieldName {
			fieldName = field.Name
		}
		name := field.Tag.Get("label")
		if "" == name {
			name = fieldName
		}
		var v string
		if "Query" == paramType {
			v = c.Query(fieldName)
		} else {
			v = c.PostForm(fieldName)
		}
		tagBinding := field.Tag.Get("binding")
		bindingMap := ConvertString2Map(tagBinding, ",", "=")
		if "" == strings.TrimSpace(v) {
			requiredMsg := field.Tag.Get("requiredMsg")
			required := bindingMap["required"]
			if required != nil && "required" == *required {
				if "" == requiredMsg {
					rm := "参数" + name + "必填"
					requiredMsg = rm
				}
				return &BadRequestError{Msg: requiredMsg}
			}
		} else {
			regularMsg := field.Tag.Get("regularMsg")
			regular := bindingMap["regexp"]
			if nil != regular && strings.TrimSpace(*regular) != "" {
				re := regexp.MustCompile(*regular)
				if !re.MatchString(v) {
					if "" == regularMsg && strings.TrimSpace(regularMsg) != "" {
						rm := "参数" + name + "格式非法"
						regularMsg = rm
					}
					return &BadRequestError{Msg: regularMsg}
				}
			}

			val, err := ConvertType(v, fileType)
			if nil != err {
				return &BadRequestError{Msg: "参数" + name + "类型非法"}
			}

			minLength := bindingMap["min"]
			if nil != minLength && strings.TrimSpace(*minLength) != "" {
				ml, _ := strconv.ParseInt(*minLength, 10, 64)
				if "string" == fileType || "*string" == fileType {
					if int64(len(v)) < ml {
						return &BadRequestError{Msg: "参数" + name + "长度过短"}
					}
				} else {
					minErr := ValidNumberByMin(val, ml, name)
					if nil != minErr {
						return minErr
					}
				}
			}

			maxLength := bindingMap["max"]
			if nil != maxLength && strings.TrimSpace(*maxLength) != "" {
				ml, _ := strconv.ParseInt(*maxLength, 10, 64)
				if "string" == fileType || "*string" == fileType {
					if int64(len(v)) > ml {
						rm := "参数" + name + "长度超长"
						regularMsg = rm
						return &BadRequestError{Msg: regularMsg}
					}
				} else {
					maxErr := ValidNumberByMax(val, ml, name)
					if nil != maxErr {
						return maxErr
					}
				}
			}
			m[fieldName] = val
		}
	}
	bytes, _ := json.Marshal(m)
	err := json.Unmarshal(bytes, s)
	if err != nil {
		return &BadRequestError{Msg: "请求参数非法，转换错误"}
	}
	return nil
}

func ValidNumberByMin(val any, ml int64, name string) error {
	return validNumber(val, ml, name, "min")
}

func ValidNumberByMax(val any, ml int64, name string) error {
	return validNumber(val, ml, name, "max")
}

func validNumber(val any, ml int64, name string, t string) error {
	flag := true
	i, ok := val.(int64)
	if ok {
		flag = false
		if "min" == t {
			if i < ml {
				return &BadRequestError{Msg: "参数" + name + "不能小于最小值"}
			}
		} else {
			if i > ml {
				return &BadRequestError{Msg: "参数" + name + "不能大于最大值"}
			}
		}
	}
	if flag {
		f, ok := val.(float64)
		if ok {
			flag = false
			if "min" == t {
				if f < float64(ml) {
					return &BadRequestError{Msg: "参数" + name + "不能小于最小值"}
				}
			} else {
				if f > float64(ml) {
					return &BadRequestError{Msg: "参数" + name + "不能大于最大值"}
				}
			}
		}
	}
	if flag {
		ui, ok := val.(uint64)
		if ok {
			if "min" == t {
				if ui < uint64(ml) {
					return &BadRequestError{Msg: "参数" + name + "不能小于最小值"}
				}
			} else {
				if ui > uint64(ml) {
					return &BadRequestError{Msg: "参数" + name + "不能大于最大值"}
				}
			}
		}
	}
	return nil
}

func ConvertSliceType(source string, typeName string) (interface{}, error) {
	var v []any
	if strings.HasPrefix(source, "[") {
		source = source[1:]
	}
	if strings.HasSuffix(source, "]") {
		source = source[:len(source)-1]
	}
	pairs := strings.Split(source, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		nv, err := ConvertType(pair, typeName[2:])
		if nil != err {
			return nil, err
		}
		v = append(v, nv)
	}
	return v, nil
}

func ConvertType(source string, typeName string) (interface{}, error) {
	if strings.HasPrefix(typeName, "[]") {
		return ConvertSliceType(source, typeName)
	}
	switch typeName {
	case "int8":
		return strconv.ParseInt(source, 10, 8)
	case "*int8":
		return strconv.ParseInt(source, 10, 8)
	case "int16":
		return strconv.ParseInt(source, 10, 16)
	case "*int16":
		return strconv.ParseInt(source, 10, 16)
	case "int":
		return strconv.Atoi(source)
	case "*int":
		return strconv.Atoi(source)
	case "int64":
		return strconv.ParseInt(source, 10, 64)
	case "*int64":
		return strconv.ParseInt(source, 10, 64)
	case "float32":
		return strconv.ParseFloat(source, 32)
	case "*float32":
		return strconv.ParseFloat(source, 32)
	case "float64":
		return strconv.ParseFloat(source, 64)
	case "*float64":
		return strconv.ParseFloat(source, 64)
	case "uint":
		return strconv.ParseUint(source, 10, 0)
	case "*uint":
		return strconv.ParseUint(source, 10, 0)
	case "uintptr":
		return strconv.ParseUint(source, 10, 0)
	case "*uintptr":
		return strconv.ParseUint(source, 10, 0)
	case "uint8":
		return strconv.ParseUint(source, 10, 8)
	case "*uint8":
		return strconv.ParseUint(source, 10, 8)
	case "uint16":
		return strconv.ParseUint(source, 10, 16)
	case "*uint16":
		return strconv.ParseUint(source, 10, 16)
	case "uint64":
		return strconv.ParseUint(source, 10, 64)
	case "*uint64":
		return strconv.ParseUint(source, 10, 64)
	case "bool":
		return strconv.ParseBool(source)
	case "*bool":
		return strconv.ParseBool(source)
	case "string":
		return source, nil
	case "*string":
		return source, nil
	}
	return nil, &ConvertTypeError{Msg: "当前类型暂不支持转换"}
}

func ConvertString2Map(tag string, split string, keySplit string) map[string]*string {
	m := make(map[string]*string)
	if "" == strings.TrimSpace(tag) {
		return m
	}
	pairs := strings.Split(tag, split)
	for _, pair := range pairs {
		kv := strings.Split(pair, keySplit)
		if len(kv) == 2 {
			key := kv[0]
			value := strings.Trim(kv[1], `"`)
			m[key] = &value
		} else if len(kv) == 1 {
			key := kv[0]
			value := kv[0]
			m[key] = &value
		}
	}
	return m
}

type BadRequestError struct {
	Msg string
}

func (e *BadRequestError) Error() string {
	return e.Msg
}

type ConvertTypeError struct {
	Msg string
}

func (e *ConvertTypeError) Error() string {
	return e.Msg
}
