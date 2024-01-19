package Define

import "regexp"

func IsHttpsURL(s string) bool {

	p := `^(https?:\/\/)([0-9a-z.]+)(:[0-9]+)?([/0-9a-z.]+)?(\?[0-9a-z&=]+)?(#[0-9-a-z]+)?`

	if m, err := regexp.Match(p, []byte(s)); err != nil {
		return false
	} else {
		return m
	}
}
