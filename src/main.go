package main

import (
	"fmt"
	"sort"
)

type A struct {
	a string
}

type B struct {
	A
	b string
}


type OrderdString []string
func (self OrderdString) insert(s string) {

	sort.Strings(self)

	fmt.Println(self)

	index := sort.SearchStrings(self, s)
	if self[index]!=s {
		self = append(self[0:index],
			append([]string{s}, self[index:]...)...)
	}
	fmt.Println(self)
}

func main() {
	s := OrderdString{"a", "c", "e", "d"}
	s.insert("d")

	fmt.Println(s)
}
