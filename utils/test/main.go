package main

import (
	"fmt"
	"github.com/hellgate75/rebind/utils"
)

func main() {
	text1 := "myhost.mail.google.com"
	text2 := "myhost.myshortdomain"
	fmt.Printf("text1=%v\n", utils.SplitDomainsFromHostname(text1))
	fmt.Printf("text2=%v\n", utils.SplitDomainsFromHostname(text2))
}
