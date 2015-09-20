package migration

import (
	"strconv"
)

func ExtractId(record Record, fieldName string) (int64, error) {
	return strconv.ParseFloat(record[fieldName].(string), 64)
}
