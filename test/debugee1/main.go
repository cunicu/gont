// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
)

type myStruct struct {
	A int
}

func main() {
	t := myStruct{
		A: 555,
	}
	i := 1337
	s := "Hello World"

	i = myFunction(s, i, t)
	i = myFunction(s, i, t)

	fmt.Println(i)
}

func myFunction(s string, i int, t myStruct) int {
	fmt.Println(s, i, true, "Bla", t)

	i *= 2

	return i
}
