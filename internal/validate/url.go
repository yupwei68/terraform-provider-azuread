package validate

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func URLIsHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return URLWithScheme([]string{"https"}, false)(i, k)
}

func URLIsHTTPOrHTTPS(i interface{}, k string) (_ []string, errors []error) {
	return URLWithScheme([]string{"http", "https"}, false)(i, k)
}

func URLIsHTTPOrHTTPSorEmpty(i interface{}, k string) (_ []string, errors []error) {
	return URLWithScheme([]string{"http", "https"}, true)(i, k)
}

func URLIsAppURI(i interface{}, k string) (_ []string, errors []error) {
	return URLWithScheme([]string{"http", "https", "api", "urn", "ms-appx"}, false)(i, k)
}

func URLWithScheme(validSchemes []string, allowEmpty bool) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (_ []string, errors []error) {
		v, ok := i.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %q to be string", k))
			return
		}

		if v == "" {
			if allowEmpty {
				return
			}
			errors = append(errors, fmt.Errorf("expected %q url to not be empty", k))
			return
		}

		u, err := url.Parse(v)
		if err != nil {
			errors = append(errors, fmt.Errorf("%q url is in an invalid format: %q (%+v)", k, v, err))
			return
		}

		if u.Host == "" {
			errors = append(errors, fmt.Errorf("%q url has no host: %q", k, v))
			return
		}

		for _, s := range validSchemes {
			if u.Scheme == s {
				return //last check so just return
			}
		}

		errors = append(errors, fmt.Errorf("expected %q url %q to have a schema of: %q", k, v, strings.Join(validSchemes, ",")))
		return
	}
}
