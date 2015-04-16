package main

import "testing"

func TestBuilding(t *testing.T) {
	err := NewAmazonStore().build()
	if err != nil {
		t.Fail()
	}
}
