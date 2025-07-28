//go:build js

package gojs

import "fmt"

func Canvas(width, height int) *Val {
	c := n("canvas")
	c.a("width", func() string { return fmt.Sprintf("%d", width) })
	c.a("height", func() string { return fmt.Sprintf("%d", height) })
	return c
}

func CanvasCtx(canvas *Val) *Val {
	return &Val{Value: canvas.Call("getContext", "2d")}
}
