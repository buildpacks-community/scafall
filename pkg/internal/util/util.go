package util

import "github.com/coveooss/gotemplate/v3/collections"

func Contains(strings []string, element string) bool {
	for _, s := range strings {
		if s == element {
			return true
		}
	}
	return false
}

func ToIDictionary(xs map[string]string) collections.IDictionary {
	ys := collections.CreateDictionary()
	for k, v := range xs {
		ys.Add(k, v)
	}
	return ys
}
