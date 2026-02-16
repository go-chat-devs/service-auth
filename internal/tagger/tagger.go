package tagger

import "fmt"

func Tagger(tag string) func(string, ...any) string {
	tag += " "
	return func(format string, args ...any) string {
		return tag + fmt.Sprintf(format, args)
	}
}
