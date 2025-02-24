package schema

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// errToErrors takes an error produces by validation.ValidateStruct and
// returns, if valid, its corresponding Errors representation. If err can not
// be converted an error is returned.
func errToErrors(err error) (Errors, error) {
	if err == nil {
		return nil, nil //nolint:nilnil // nil is a valid value
	}
	errMap, ok := err.(validation.Errors) //nolint:errorlint // Not comparing
	if !ok {
		return nil, fmt.Errorf("failed to validate form: %w", err)
	}
	errs := make(Errors, len(errMap))
	for k, v := range errMap {
		errs[k] = v.Error()
	}
	return errs, nil
}

// matchingFieldsRule returns a validation.RuleFunc forcing that the value
// passed matches field. If the rule is not met then an error specifying the
// target field's name is returned.
func matchingFieldsRule[V comparable](
	field V,
	name string,
) validation.RuleFunc {
	return func(v interface{}) error {
		s, ok := v.(V)
		if !ok {
			return fmt.Errorf("must be equal to %s", name)
		}
		if s != field {
			return fmt.Errorf("must be equal to %s", name)
		}
		return nil
	}
}
