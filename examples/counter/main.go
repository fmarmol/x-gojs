// go:build js

package main

import (
	"fmt"

	. "github.com/fmarmol/x-gojs"
)

type Counter struct {
	count int
}

func (c *Counter) View() *Val {
	div := Div()
	div.C(
		Button().C(
			Text(String("inc"))).
			OnClick(func() {
				c.count++
				div.Render()
			}),
		Text(func() string {
			return fmt.Sprint(c.count)
		}),
		Button().C(Text(String("dec"))).
			OnClick(func() {
				c.count--
				div.Render()
			}),
	)
	return div
}

func main() {

	stop := make(chan struct{})
	c := new(Counter)
	Init(c.View())
	<-stop

}
