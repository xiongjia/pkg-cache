package cache

import "regexp"

func matchPattern(url, pattern string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(url), nil
}
