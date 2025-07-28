//go:build js

package gojs

import "syscall/js"

type WebSocket struct {
	ws js.Value
}

type WebSocketBinType int

const (
	ArrayBuffer WebSocketBinType = iota
	Blob
)

func NewWebSocket(addr string, binaryType WebSocketBinType) *WebSocket {
	ws := js.Global().Get("WebSocket").New(addr)
	switch binaryType {
	case Blob:
	case ArrayBuffer:
		ws.Set("binaryType", "arraybuffer")
	}

	return &WebSocket{ws: ws}
}

type Marshaler interface {
	Marshal() ([]byte, error)
}

func (w *WebSocket) SendM(v Marshaler) error {
	data, err := v.Marshal()
	if err != nil {
		return err
	}
	w.Send(data)
	return nil
}

func (w *WebSocket) Send(payload []byte) {
	jsPayload := js.Global().Get("Uint8Array").New(len(payload))
	js.CopyBytesToJS(jsPayload, payload)
	w.ws.Call("send", jsPayload)
}

func (w *WebSocket) OnOpen(fn JsFunc) {
	w.ws.Set("onopen", js.FuncOf(fn))
}

func (w *WebSocket) OnMessage(fn JsFunc) {
	w.ws.Set("onmessage", js.FuncOf(fn))
}
