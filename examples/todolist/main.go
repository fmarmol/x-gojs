//go:build js

package main

import (
	"strings"

	. "github.com/fmarmol/x-gojs"
)

type Task struct {
	Title string
	Done  bool
}

func (t *Task) View() *Val {
	input := Input().Attr("type", String("checkbox"))
	if t.Done {
		input.Attr("checked", String("true"))
	}
	return Div().Class("flex flex-row items-center w-full justify-between p-1").
		ClassOnRevCond(func() bool { return t.Done }, "bg-green-300", "bg-slate-300").
		C(
			Text(String(t.Title)),
			input.
				OnClick(func() {
					t.Done = !t.Done
					input.Parent.Render()
				}),
		)
}

type TodoList struct {
	Tasks []Task
}

func (t *TodoList) HasTitle(title string) bool {
	for _, task := range t.Tasks {
		if task.Title == title {
			return true
		}
	}
	return false
}

func (c *TodoList) View() *Val {
	div := Div().Class("flex flex-col justify-center w-fit mx-auto my-auto min-h-screen")
	input := Input().Attr("type", String("text")).Class("border border-slate-300 h-[41px] pl-2")
	div.C(
		Div().Class("flex flex-row items-center border border-black p-1 gap-2").C(
			input,
			Button().Class("bg-blue-300 p-1 border border-black rounded h-[41px] hover:cursor-pointer hover:bg-blue-400").
				C(Text(String("new task"))).
				OnClick(func() {
					text := input.Value.Get("value").String()
					if strings.TrimSpace(text) == "" {
						return
					}
					if !c.HasTitle(text) {
						newTask := &Task{Title: text}
						c.Tasks = append(c.Tasks, *newTask)
						div.C(newTask.View())
						div.Render()
						input.Value.Set("value", "")
					}

				}),
		),
	)
	for _, t := range c.Tasks {
		div.C(t.View())
	}
	return div
}

func main() {
	stop := make(chan struct{})
	c := &TodoList{
		Tasks: []Task{
			{Title: "first"},
		},
	}
	Init(c.View())
	<-stop

}
