package main

import (
	"fmt"
	"testing"
)

func TestGetCodeComplete(t *testing.T) {
	for _, value := range GetCodeComplete("fmt") {
		fmt.Println(value)
	}

}
