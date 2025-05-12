//go:build js

package gojs

import (
	"fmt"
	"reflect"
	"strings"
	"syscall/js"

	"github.com/google/uuid"
)

type Component[T any] struct {
	*Val
	S T
}

func NewComponent[T any](s T) *Component[T] {
	return &Component[T]{
		Val: new(Val),
		S:   s,
	}
}

type Attr struct {
	Key   string
	Value func() string
}

func String(s string) func() string {
	return func() string {
		return s
	}
}

type ClassCond struct {
	Class string
	Cond  func() bool
}
type ClassRevCond struct {
	class1 string
	class2 string
	cond   func() bool
}

type Val struct {
	Value           js.Value
	attrs           []Attr
	styles          []Attr
	classes         []string
	classesOnCond   []ClassCond
	classesOnRevCon []ClassRevCond
	onclick         any
	id              string
	children        []*Val
	textfn          func() string
	Parent          *Val
	eventListeners  map[string]struct{}
}

func (v *Val) RemoveChild(child *Val) *Val {
	newChildren := []*Val{}
	for _, c := range v.children {
		if c.id == child.id {
			continue
		}
		newChildren = append(newChildren, c)
	}
	v.children = newChildren
	v.Value.Call("removeChild", child.Value)
	return v
}

func (v *Val) ClassOnCond(value string, f func() bool) *Val {
	values := strings.Fields(value)
	for _, value := range values {
		v.classesOnCond = append(v.classesOnCond, ClassCond{Class: value, Cond: f})
	}

	return v
}

func (v *Val) ClassOnRevCond(f func() bool, c1, c2 string) *Val {
	v.classesOnRevCon = append(v.classesOnRevCon, ClassRevCond{cond: f, class1: c1, class2: c2})
	return v
}

func (v *Val) Class(values ...string) *Val {
	for _, va := range values {
		subv := strings.Fields(va)
		v.classes = append(v.classes, subv...)
	}
	return v
}

func (v *Val) Text(f func() string) *Val {
	v.textfn = f
	return v
}

// func (v *Val) OnClick2(f any) *Val {
// 	v.onclick2 = f
// 	return v
// }

func (v *Val) OnClick(f any) *Val {
	v.onclick = f
	return v
}

func (v *Val) C(others ...*Val) *Val {
	for _, other := range others {
		v.children = append(v.children, other)
		other.Parent = v
		v.c(other)
	}
	return v
}

func (v *Val) Attr(key string, value func() string) *Val {
	v.attrs = append(v.attrs, Attr{Key: key, Value: value})
	return v
}

func (v *Val) Style(key string, value func() string) *Val {
	v.styles = append(v.styles, Attr{Key: key, Value: value})
	return v
}

func (v *Val) AddClass(c string) *Val {
	v.Value.Get("classList").Call("add", c)
	return v
}

func (v *Val) DelClass(c string) *Val {
	v.Value.Get("classList").Call("remove", c)
	return v
}

func (v *Val) SetStyle(key string, value func() string) *Val {
	v.Value.Get("style").Set(key, value())
	return v
}

func (v *Val) Render() *Val {
	for _, child := range v.children {
		child.Render()
	}
	for _, attr := range v.attrs {
		v.a(attr.Key, attr.Value)
	}
	for _, style := range v.styles {
		v.Value.Get("style").Set(style.Key, style.Value())
	}
	for _, class := range v.classes {
		v.AddClass(class)
	}
	for _, class := range v.classesOnCond {
		if class.Cond() {
			v.AddClass(class.Class)
		} else {
			v.DelClass(class.Class)
		}
	}
	for _, class := range v.classesOnRevCon {
		classesOk := strings.Fields(class.class1)
		classesKO := strings.Fields(class.class2)
		if class.cond() {
			for _, c := range classesOk {
				v.Value.Get("classList").Call("add", c)
			}
			for _, c := range classesKO {
				v.Value.Get("classList").Call("remove", c)
			}
		} else {
			for _, c := range classesKO {
				v.Value.Get("classList").Call("add", c)
			}
			for _, c := range classesOk {
				v.Value.Get("classList").Call("remove", c)
			}
		}
	}

	if v.onclick != nil {
		v.f("click", v.onclick)
	}
	v.a("id", String(v.id))
	if v.textfn != nil {
		v.Value.Set("innerHTML", v.textfn())
	}
	return v
}

func (v *Val) CreateElement(elem string) *Val {
	id := uuid.NewString()

	n := &Val{Value: v.Value.Call("createElement", elem), id: id}
	return n
}

func (v *Val) c(child *Val) *Val {
	v.Value.Call("appendChild", child.Value)
	return v
}

func (v *Val) a(attrName string, value func() string) *Val {
	v.Value.Set(attrName, value())
	return v
}

func (v *Val) f(event string, value any) *Val {
	_value := reflect.TypeOf(value)
	if _value.Kind() != reflect.Func {
		panic(fmt.Errorf("cannot only use function, but received a %v", _value.Kind()))

	}
	if v.eventListeners == nil {
		v.eventListeners = map[string]struct{}{}
	}
	_, ok := v.eventListeners[event]
	if ok {
		return v
	}

	v.eventListeners[event] = struct{}{}
	fnType := reflect.TypeOf(func(js.Value, []js.Value) any { return nil })

	switch _value.NumIn() {
	case 0:
		fn := value.(func())
		resFn := reflect.MakeFunc(fnType, func(args []reflect.Value) (results []reflect.Value) {
			fn()
			return []reflect.Value{reflect.ValueOf(1)}
		})
		goF := resFn.Convert(fnType).Interface().(func(js.Value, []js.Value) any)
		jsF := js.FuncOf(goF)
		v.Value.Call("addEventListener", event, jsF)
		return v
	}
	return v

}

type Doc struct{ Val }

func NewDoc() Doc {
	v := js.Global().Get("document")
	return Doc{Val{Value: v}}
}

func (d *Doc) Body() *Val {
	b := d.Val.Value.Get("body")
	return &Val{Value: b}
}

var doc = NewDoc()

type Html struct{}

func n(kind string) *Val {
	v := doc.CreateElement(kind)
	if !v.Value.Truthy() {
		panic(fmt.Errorf("create Element %s is not valid", kind))

	}
	return v
}

func Input() *Val {
	input := n("INPUT")
	if !input.Value.Truthy() {
		panic("input not valid")
	}
	return input
}

func Text(t func() string) *Val {
	text := n("TEXT")
	if !text.Value.Truthy() {
		panic("text not valid")
	}
	text.textfn = t
	return text

}

func Button() *Val {
	return n("BUTTON")
}

func Img() *Val {
	return n("IMG")
}

func Div() *Val {
	return n("DIV")
}

func Delete(v *Val) {
	child := doc.Val.Value.Call("getElementById", v.id)
	parent := v.Parent
	parent.Value.Call("removeChild", child)
}

func Init(v *Val) {
	body := doc.Body()
	v.Parent = body
	body.
		C(v)
	body.Render()
}
