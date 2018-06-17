package person

import (
	"fmt"
)

func NewPerson(id int, secret string) *Person {
	return &Person{
		ID:     1,
		Name:   "Foo Bar",
		secret: secret,
	}
}

func GetSecret(p *Person) string {
	return p.secret
}

func printSecret(p *Person) {
	fmt.Println(p.secret)
}
