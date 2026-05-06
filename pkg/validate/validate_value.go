package validate

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/suzuki-shunsuke/cmdx/pkg/domain"
)

const (
	validateTypeEmail = "email"
	validateTypeURL   = "url"
	validateTypeInt   = "int"
)

func ValueWithValidates(val string, validates []domain.Validate) error {
	for _, validateParam := range validates {
		if err := value(val, validateParam); err != nil {
			return err
		}
	}
	return nil
}

func value(val string, validate domain.Validate) error {
	switch validate.Type {
	case validateTypeEmail:
		if !govalidator.IsEmail(val) {
			return errors.New("must be email: " + val)
		}
	case validateTypeURL:
		if !govalidator.IsURL(val) {
			return errors.New("must be url: " + val)
		}
	case validateTypeInt:
		if !govalidator.IsInt(val) {
			return errors.New("must be int: " + val)
		}
	}
	if validate.Contain != "" {
		if !strings.Contains(val, validate.Contain) {
			return errors.New("must contain " + validate.Contain + ": " + val)
		}
	}
	if validate.Prefix != "" {
		if !strings.HasPrefix(val, validate.Prefix) {
			return errors.New("must start with " + validate.Prefix + ": " + val)
		}
	}
	if validate.Suffix != "" {
		if !strings.HasSuffix(val, validate.Suffix) {
			return errors.New("must end with " + validate.Suffix + ": " + val)
		}
	}
	if validate.MinLength != 0 {
		if len(val) < validate.MinLength {
			return errors.New("the length must be greater equal than " + strconv.Itoa(validate.MinLength) + ": " + val)
		}
	}
	if validate.MaxLength != 0 {
		if len(val) > validate.MaxLength {
			return errors.New("the length must be less equal than " + strconv.Itoa(validate.MaxLength) + ": " + val)
		}
	}
	if len(validate.Enum) != 0 {
		if !slices.Contains(validate.Enum, val) {
			return errors.New("enum (" + strings.Join(validate.Enum, ", ") + "): " + val)
		}
	}
	if validate.RegExp != "" {
		f, err := regexp.MatchString(validate.RegExp, val)
		if err != nil {
			return fmt.Errorf("invalid regular expression: "+validate.RegExp+": %w", err)
		}
		if !f {
			return errors.New("must be matched to the regular expression " + validate.RegExp + ": " + val)
		}
	}
	return nil
}
