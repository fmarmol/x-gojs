//go:build js

package gojs

import (
	"fmt"
	"log"
	"math/rand/v2"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall/js"
	"time"
	"unsafe"
)

type Elem struct {
	value any
	val   *Val
}

var nodes map[unsafe.Pointer][]*Elem

func init() {
	nodes = make(map[unsafe.Pointer][]*Elem)
}

type EventKind int

type Event2 struct {
	Val          *Val
	Args         any
	eventKind    int
	response     chan struct{}
	NeedResponse bool
}

var eventChan = make(chan Event2)

func Send(eventKind int, event Event2) <-chan struct{} {
	event.eventKind = eventKind
	event.response = make(chan struct{}, 1)
	eventChan <- event
	return event.response
}

func (v *Val) SubScribe(eventKind int, f func(e Event2)) (cancel func()) {
	newC := make(chan Event2)
	hub.register(eventKind, newC)
	go func() {
		for event := range newC {
			f(event)
			if event.NeedResponse { // WARN: if multiple subscribers can lock goroutine, Only use if only one subcriber !!
				timer := time.NewTimer(10 * time.Millisecond)
				select {
				case event.response <- struct{}{}:
				case <-timer.C:
					panic(fmt.Errorf("timeout in response to event: %v", event))
				}
			}
		}
	}()
	return func() {
		close(newC)
	}
}

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

func String(s any) func() string {
	return func() string {
		return fmt.Sprint(s)
	}
}

func StringFmt(format string, values ...any) func() string {
	return func() string {
		return fmt.Sprintf(format, values...)
	}
}

type ClassCond struct {
	Class string
	Cond  func() bool
}

type AttrCond struct {
	Attr  string
	Value func() string
	Cond  func() bool
}

type ClassRevCond struct {
	class1 string
	class2 string
	cond   func() bool
	ptr    *bool
}

type Val struct {
	Value           js.Value
	attrs           []Attr
	attrsOnCond     []AttrCond
	styles          []Attr
	classes         []string
	classesOnCond   []ClassCond
	classesOnRevCon []ClassRevCond
	onclick         any
	id              string
	children        []*Val
	textfn          func() string
	Parent          *Val
	IdxInParent     int
	eventListeners  map[string]struct{}
	eventChan       chan Event2
	mux             sync.Mutex
}

type Imgui struct{ val *Val }

func (v *Val) Imgui() *Imgui {
	return &Imgui{val: v}
}

func (v *Val) ID(id string) *Val {
	v.id = id
	return v
}

func (v *Val) Remove() {
	v.Value.Call("remove")
	// v = nil ??
}

func (v *Val) GetID() string {
	return v.id
}

func (v *Val) Children() []*Val {
	return v.children
}

func (v *Val) SwapChildren(i, j int) *Val {
	children := v.children
	for _, child := range children {
		v.RemoveChild(child)
	}
	children[i], children[j] = children[j], children[i]
	for _, child := range children {
		v.C(child)
	}
	return v
}

func (i *Imgui) RemoveChild(child *Val) *Val {
	i.val.Call("removeChild", child.Value)
	return i.val
}

func (v *Val) RemoveChild(child *Val) *Val {
	v.mux.Lock()
	defer v.mux.Unlock()
	if child.Parent != v {
		panic(fmt.Errorf("cannot remove child parent %v different from caller %v", child.Parent.GetID(), v.GetID()))
	}
	if !child.Parent.Value.Equal(v.Value) {
		panic(fmt.Errorf("cannot remove child parent %v different from caller %v", child.Parent.GetID(), v.GetID()))
	}

	newChildren := make([]*Val, 0, len(v.children))
	for _, c := range v.children {
		if c.id == child.id {
			continue
		}
		newChildren = append(newChildren, c)
		c.IdxInParent = len(newChildren)
	}
	v.children = newChildren
	v.Value.Call("removeChild", child.Value)
	return v
}
func (v *Val) AttrOnCond(attr string, value func() string, cond func() bool) *Val {
	v.attrsOnCond = append(v.attrsOnCond, AttrCond{Attr: attr, Value: value, Cond: cond})
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

func (v *Val) ClassOnRevCond2(b *bool, c1, c2 string) *Val {
	v.classesOnRevCon = append(v.classesOnRevCon, ClassRevCond{ptr: b, class1: c1, class2: c2})
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

func State[T any](val *Val, v *T) {
	if v == nil {
		panic("pointer nil")
	}
	ptr := unsafe.Pointer(v)
	nodes[ptr] = append(nodes[ptr], &Elem{val: val, value: *v})
}

func Update[T any](v *T) {
	ptr := unsafe.Pointer(v)
	nodes, ok := nodes[ptr]
	if !ok {
		_, file, line, _ := runtime.Caller(1)
		log.Printf("Warning not found from: file=%s, line=%d\n", file, line)
	}
	for _, node := range nodes {
		new_value := *v
		prev_value := node.value

		if !reflect.DeepEqual(new_value, prev_value) {
			node.value = new_value
			node.val.Render()
		}
	}
}

func (v *Val) OnClick(f any) *Val {
	v.onclick = f
	return v
}

func (v *Val) OnChange(f JsFunc) *Val {
	return v.f("change", f)
}
func (v *Val) OnInput(f JsFunc) *Val {
	return v.f("input", f)
}

func (imgui *Imgui) C(others ...*Val) *Val {
	for _, other := range others {
		imgui.val.c(other)
	}
	return imgui.val
}

func (v *Val) C(others ...*Val) *Val {
	for _, other := range others {
		v.children = append(v.children, other)
		other.IdxInParent = len(v.children) - 1
		other.Parent = v
		v.c(other)
	}
	return v
}

func (v *Val) P(others ...*Val) *Val {
	for _, other := range others {
		other.Parent = v
		v.p(other)
		v.children = append([]*Val{other}, v.children...)
	}
	return v
}

func (v *Val) MouseMove(f JsFunc) *Val {
	return v.f("mousemove", f)
}

func (v *Val) MouseEnter(f JsFunc) *Val {
	return v.f("mouseenter", f)
}

func (v *Val) MouseLeave(f JsFunc) *Val {
	return v.f("mouseleave", f)
}

func (v *Val) MouseRightClick(f JsFunc) *Val {
	return v.f("contextmenu", f)
}

func (v *Val) MouseLeftClick(f JsFunc) *Val {
	return v.f("click", f)
}

func (v *Val) MouseDblClick(f JsFunc) *Val {
	return v.f("dblclick", f)
}

func (v *Val) DragEnd(f JsFunc) *Val {
	return v.f("dragend", f)
}

func (v *Val) Draggable(f JsFunc) *Val {
	v.Attr("draggable", String("true"))
	return v.f("dragstart", f)
}

func (v *Val) OnDrop(f JsFunc) *Val {
	return v.f("drop", f)
}

func (v *Val) OnDragOver(f JsFunc) *Val {
	return v.f("dragover", f)
}

func (v *Val) Attr(key string, value func() string) *Val {
	v.attrs = append(v.attrs, Attr{Key: key, Value: value})
	return v
}

func (v *Val) Style(key string, value func() string) *Val {
	v.styles = append(v.styles, Attr{Key: key, Value: value})
	return v
}

func (i *Imgui) AddClass(c string) *Val {
	i.val.Value.Get("classList").Call("add", c)
	return i.val
}
func (i *Imgui) DelClass(c string) *Val {
	i.val.Value.Get("classList").Call("remove", c)
	return i.val
}

func (v *Val) AddClass(c string) *Val {
	v.Value.Get("classList").Call("add", c)
	return v
}

func (v *Val) DelClass(c string) *Val {
	v.Value.Get("classList").Call("remove", c)
	return v
}

func (imgui *Imgui) SetStyle(key string, value string) *Val {
	imgui.val.Value.Get("style").Set(key, func() string { return value })
	return imgui.val
}

func (v *Val) SetStyle(key string, value func() string) *Val {
	v.Value.Get("style").Set(key, value())
	return v
}

func (v *Val) Render() *Val {
	// fmt.Println("render:", v.id)
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
	for _, attr := range v.attrsOnCond {
		if attr.Cond() {
			v.a(attr.Attr, attr.Value)
		} else {
			v.rma(attr.Attr)
		}
	}
	for _, class := range v.classesOnRevCon {
		classesOk := strings.Fields(class.class1)
		classesKO := strings.Fields(class.class2)
		if class.ptr != nil {
			if *class.ptr {
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
			continue
		}
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
	id := "id_" + fmt.Sprint(rand.Int32())
	// id := uuid.NewString()

	n := &Val{Value: v.Value.Call("createElement", elem), id: id}
	return n
}

func (v *Val) Call(funcname string, args ...any) js.Value {
	return v.Value.Call(funcname, args...)
}

// func (i *Imgui) C(child *Val) *Val {
// 	return i.val.c(child)
// }

func (v *Val) c(child *Val) *Val {
	v.Call("appendChild", child.Value)
	return v
}
func (v *Val) p(child *Val) *Val {
	if len(v.children) == 0 {
		v.c(child)
	} else {
		if child.Parent != v.children[0].Parent {
			panic("not possible")
		}
		v.Call("insertBefore", child.Value, v.children[0].Value)
	}
	return v
}

func (i *Imgui) A(attrName string, value func() string) *Val {
	return i.val.a(attrName, value)
}

func (v *Val) a(attrName string, value func() string) *Val {
	v.Value.Set(attrName, value())
	return v
}

func (v *Val) rma(attrName string) *Val {
	v.Value.Call("removeAttribute", attrName)
	return v
}

func (v *Val) AddEventListener(event string, fn any) {
	v.f(event, fn)
}

func (v *Val) f(event string, value any) *Val {
	_type := reflect.TypeOf(value)
	if _type.Kind() != reflect.Func {
		panic(fmt.Errorf("cannot only use function, but received a %v", _type.Kind()))

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

	switch _type.NumIn() {
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
	case 2:
		fn := value.(func(this js.Value, args []js.Value) any)
		resFn := reflect.MakeFunc(fnType, func(args []reflect.Value) (results []reflect.Value) {
			arg0 := args[0]
			arg1 := args[1]
			fn(arg0.Interface().(js.Value), arg1.Interface().([]js.Value))
			return []reflect.Value{reflect.ValueOf(1)}
		})
		goF := resFn.Convert(fnType).Interface().(func(js.Value, []js.Value) any)
		jsF := js.FuncOf(goF)
		v.Value.Call("addEventListener", event, jsF)
		return v
	default:
		panic(fmt.Errorf("func with %d args not supported", _type.NumIn()))
	}
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

type Func struct {
	js.Func
	Name string
}

func GetElementById(id string) *Val {
	child := doc.Val.Value.Call("getElementById", id)
	return &Val{Value: child}
}

func RegisterFunc(name string, f js.Func) *Func {
	js.Global().Set(name, f)
	return &Func{
		Name: name,
		Func: f,
	}
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

func Details() *Val {
	return n("DETAILS")
}

func Summary() *Val {
	return n("SUMMARY")
}

func Div() *Val {
	return n("DIV")
}

func Svg() *Val {
	return n("SVG")
}

func Style() *Val {
	return n("style")
}

func Delete(v *Val) {
	child := doc.Val.Value.Call("getElementById", v.id)
	parent := v.Parent
	parent.Value.Call("removeChild", child)
}

type Hub struct {
	sync.Mutex
	subscribers map[int][]chan Event2
}

var hub = Hub{
	subscribers: make(map[int][]chan Event2),
}

func (h *Hub) register(event int, ch chan Event2) {
	h.Lock()
	h.subscribers[event] = append(h.subscribers[event], ch)
	h.Unlock()
}

func (h *Hub) run() {
	for event := range eventChan {
		for _, ch := range h.subscribers[event.eventKind] {
			ch <- event
		}
	}
}

func Init(v *Val) {
	go func() {
		hub.run()
	}()
	body := doc.Body()
	v.Parent = body
	body.
		C(v)
	body.Render()

}

type JsFunc = func(js.Value, []js.Value) any
