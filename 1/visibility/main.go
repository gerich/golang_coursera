package main

import (
	"coursera/1/visibility/person"
	"fmt"
)

func main() {
	p := person.NewPerson(1, "foo")
	fmt.Println(p.GetSecret())
}
