package domain

import (
	"github.com/go-playground/validator/v10"
)

// validate - общий валидатор для домена
var validate = validator.New(validator.WithRequiredStructEnabled())
