// Package template just holds our email-template.
//
// This is abstracted because we want to refer to it from our
// processor-package, which is not in package-main, and also
// the template-listing command.
package template

import (
	_ "embed" // embedded-resource magic
)

//go:embed template.txt
var message string

// EmailTemplate returns the embedded email template.
func EmailTemplate() []byte {
	return []byte(message)
}
