package validate

import "fmt"

// NoEmptyStrings validates that the string is not just whitespace characters (equal to [\r\n\t\f\v ])
func ValidInt32(i interface{}, k string) ([]string, []error) {
	v, ok := i.(int)
	if !ok {
		return nil, []error{fmt.Errorf("expected type of %q to be integer", k)}
	}

	if v < -2147483648 || v > 2147483647 {
		return nil, []error{fmt.Errorf("value of %q is too large", k)}
	}

	return nil, nil
}
