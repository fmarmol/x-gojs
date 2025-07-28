//go:build js

package gojs

import (
	"encoding/base64"
	"fmt"
)

func Audio(autoplay bool) *Val {
	audio := n("audio")
	if autoplay {
		audio.Imgui().A("autoplay", String("true"))
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
