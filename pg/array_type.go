package pg

import (
	"database/sql/driver"
	"strconv"
	"strings"
)

type Int64Array []int64

func (a *Int64Array) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		*a = Int64Array{}
		return nil
	}

	asString := string(asBytes)
	elements := strings.Split(strings.Trim(asString, "{}"), ",")
	for _, e := range elements {
		i, err := strconv.ParseInt(e, 10, 64)
		if err != nil {
			return err
		}
		*a = append(*a, i)
	}

	return nil
}

func (a *Int64Array) Value() (driver.Value, error) {
	return *a, nil
}

type StringArray []string

func (a *StringArray) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		*a = StringArray{}
		return nil
	}

	asString := string(asBytes)
	elements := strings.Split(strings.Trim(asString, "{}"), ",")
	for _, e := range elements {
		*a = append(*a, e)
	}

	return nil
}

func (a *StringArray) Value() (driver.Value, error) {
	return *a, nil
}
