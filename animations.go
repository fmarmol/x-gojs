//go:build js

package gojs

import (
	"fmt"
	"syscall/js"
	"time"
)

type Animation struct {
	Data   map[string]any
	Offset float32
}
type AnimationConfig struct {
	Iterations int
	Duration   time.Duration
	Infinity   bool
	CallBack   func()
}

func Translate(x, y int) Animation {
	return Animation{Data: map[string]any{"transform": fmt.Sprintf("translate(%dpx, %dpx)", x, y)}}
}

// func AnimCombine(animations []Animation) Animation {
// 	res := make(map[string]any)
// 	for _, anim := range animations {
// 		for key, value := range anim.Data {
// 			str, ok := value.(string)
// 			if !ok {
// 				panic("not managed")
// 			}
// 			_, ok = res[key]
// 			initStr = ""
// 			if !ok {
// 				res[key] = initStr
// 			}

// 		}

// 	}
// }

func Rotation(angle float32) Animation {
	return Animation{Data: map[string]any{"transform": fmt.Sprintf("rotate(%vdeg)", angle)}}
}
func Scale(scale float32) Animation {
	return Animation{Data: map[string]any{"transform": fmt.Sprintf("scale(%v)", scale)}}
}

// recv if provided will be set to the animation object
func (v *Val) Animate(animations []Animation, cfg AnimationConfig, recv ...*js.Value) *Val {
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

	param2["duration"] = cfg.Duration.Milliseconds()

	_p1 := []any{}
	for _, m := range param1 {
		_p1 = append(_p1, js.ValueOf(m))
	}
	_p11 := js.ValueOf(_p1)

	_p2 := js.ValueOf(param2)
	if len(recv) > 0 && recv[0] != nil {
		*recv[0] = v.Value.Call("animate", _p11, _p2)
	} else {
		v.Value.Call("animate", _p11, _p2)
	}
	if cfg.CallBack != nil {
		cfg.CallBack()
	}
	return v
}
