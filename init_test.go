package main

import "testing"

func Test(t *testing.T) {
	if loadInitParams().OutputDir == "" {
		t.Errorf("Oops")
	}
}
