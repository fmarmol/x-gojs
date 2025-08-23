//go:build js

package gojs

import (
	"encoding/base64"
	"fmt"
)

func Audio(autoplay bool, controls bool, muted bool, loop bool) *Val {
	audio := n("audio")
	if autoplay {
		audio.Imgui().A("autoplay", String("true"))
	}
	if controls {
		audio.Imgui().A("controls", String("true"))
	}
	if muted {
		audio.Imgui().A("muted", String("true"))
	}
	if loop {
		audio.Imgui().A("loop", String("true"))
	}
	return audio
}

// Source ext can be wav, ...
func Source(music []byte, ext string) *Val {
	source := n("source")

	b64 := base64.StdEncoding.EncodeToString(music)
	source.Imgui().A("src", func() string {
		return fmt.Sprintf("data:audio/%v;base64,%v", ext, b64)

	})
	return source
}
