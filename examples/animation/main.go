// go:build js

package main

import (
	"time"

	. "github.com/fmarmol/gojs"
)

func View() *Val {

	div := Div().C(
		Div().Text(String("moving")).
			Style("position", String("absolute")).
			Style("left", String("300px")).
			Style("top", String("300px")),
	).Animate([]Animation{
		Scale(2),
		Scale(0),
		Rotation(0),
		Rotation(30),
		Rotation(60),
		Rotation(90),
		Rotation(120),
		Rotation(150),
		Rotation(180),
	}, AnimationConfig{Infinity: true, Duration: 10 * time.Second})

	return div
}

func main() {

	stop := make(chan struct{})
	Init(View())

	<-stop

}
