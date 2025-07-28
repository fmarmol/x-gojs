//go:build js

package main

import (
	. "github.com/fmarmol/gojs"
)

func main() {
	stop := make(chan struct{})
	text1 := Div().Text(String("1"))
	text2 := Div().Text(String("2"))

	div := Div()
	div.C(
		text1,
		text2,
		Button().
			OnClick(func() {
				div.SwapChildren(0, 1)
			}).
			C(Text(String("swap"))),
	)
	Init(div)

	<-stop
}
