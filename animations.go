//go:build js

package gojs

import (
	"fmt"
	"syscall/js"
)

type Animation struct {
	Data   map[string]any
	Offset float32
}
type AnimationConfig struct {
	Iterations int
	Duration   int
	Infinity   bool
}

func Rotation(angle float32) Animation {
	return Animation{Data: map[string]any{"transform": fmt.Sprintf("rotate(%vdeg)", angle)}}
}
func Scale(scale float32) Animation {
	return Animation{Data: map[string]any{"transform": fmt.Sprintf("scale(%v)", scale)}}
}

func (v *Val) Animate(animations []Animation, cfg AnimationConfig) *Val {
	param1 := []map[string]any{}
	for _, animation := range animations {
		if animation.Offset != 0 {
			animation.Data["offset"] = animation.Offset
		}
		param1 = append(param1, animation.Data)
	}
	param2 := map[string]any{}
	if cfg.Infinity {
		param2["iterations"] = "Infinity"
	} else {
		param2["iterations"] = cfg.Iterations
	}

	param2["duration"] = cfg.Duration

	_p1 := []any{}
	for _, m := range param1 {
		_p1 = append(_p1, js.ValueOf(m))
	}
	_p11 := js.ValueOf(_p1)

	_p2 := js.ValueOf(param2)
	v.Value.Call("animate", _p11, _p2)
	return v
}
