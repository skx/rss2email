package main

import (
	"testing"
)

func TestCronNoArguments(t *testing.T) {

	c := cronCmd{}

	out := c.Execute([]string{})
	if out != 1 {
		t.Fatalf("Expected error when called with no arguments")
	}
}

func TestCronNotEmails(t *testing.T) {

	d := daemonCmd{}

	out := d.Execute([]string{"foo@example.com", "bar"})
	if out != 1 {
		t.Fatalf("Expected error when called with non-email addresses")
	}
}
