package main

import "testing"

func TestHello(t *testing.T) {
	want := "Hello, CI!"
	if got := Hello(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
