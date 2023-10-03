package main

import (
	"fmt"
	"testing"
)

func TestCleanBody(t *testing.T) {
	cases := []struct{ body, cleaned string }{
		{
			body:    "Today is such a kerfuffle day",
			cleaned: "Today is such a **** day",
		},
		{
			body:    "I like turtles",
			cleaned: "I like turtles",
		},
		{
			body:    "Everybody keeps telling me to say sharbert and fornax today, but I won't",
			cleaned: "Everybody keeps telling me to say **** and **** today, but I won't",
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cleanedBody := cleanBody(c.body)
			if cleanedBody != c.cleaned {
				t.Errorf("Expected to be cleaned correctly")
				return
			}
		})
	}
}
