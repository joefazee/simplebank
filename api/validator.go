package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/joefazee/simplebank/util"
)

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsSupprtedCurrency(currency)
	}
	return false
}
