package main

import (
	"log"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestHelloName(t *testing.T) {
	s := downTxt("")
	for _, e := range s {
		log.Println(e)
	}
	log.Println(len(s))
}

// TestHelloEmpty calls greetings.Hello with an empty string,
// checking for an error.
func TestDown(t *testing.T) {

}
