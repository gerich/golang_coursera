package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	foo, _ := ioutil.ReadDir("hw1_tree/testdata")
	fmt.Printf("%#v\n", foo)
}
