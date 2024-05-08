// Package format and console layout utilities
package format

import (
	"github.com/gosuri/uitable"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TitleCase returns the string with the first letter of each word to the capital case.
// this function initializes a caser with the language und, and performs the task.
// the strings.Title function is deprecated, and recommends using the package
// golang.org/x/text/cases for returning a string in title case.
func TitleCase(str string) string {
	c := cases.Title(language.English)
	return c.String(str)
}

// NewTable returns a table for use to display information.
func NewTable() *uitable.Table {
	t := uitable.New()
	t.Wrap = true
	return t
}
