package components

import (
    "github.com/charmbracelet/bubbles/viewport"
)

// NewMessageViewport returns a viewport configured to allow horizontal and vertical
// scrolling and with wrapping disabled so long lines can scroll horizontally.
func NewMessageViewport(width, height int, content string) viewport.Model {
    vp := viewport.New(width, height)
    vp.SetContent(content)
    vp.YOffset = 0
    vp.XOffset = 0
    vp.HighPerformanceRendering = true
    vp.MouseWheelEnabled = true
    vp.WrapText = false // disable wrapping for horizontal scroll
    return vp
}
