package main

import (
	"testing"
)

func TestDaemonNoArguments(t *testing.T) {

	d := daemonCmd{}

	out := d.Execute([]string{})
	if out != 1 {
		t.Fatalf("Expected error when called with no arguments")
	}
}

func TestDaemonNotEmails(t *testing.T) {

	d := daemonCmd{}

	out := d.Execute([]string{"foo@example.com", "bart"})
	if out != 1 {
		t.Fatalf("Expected error when called with non-email addresses")
	}
}
