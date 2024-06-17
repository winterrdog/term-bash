package main

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/creack/pty"
)

// max buffer length of output history
const MaxBufferSize = 32

func main() {
	// create a new app
	app := app.New()
	window := app.NewWindow("term-bash")

	// make a new text grid
	ui := widget.NewTextGrid()

	// set up the call to bash
	os.Setenv("TERM", "dumb")
	sh := exec.Command("/bin/bash")

	// start the pty
	tty, err := pty.Start(sh)
	if err != nil {
		fyne.LogError("failed to start pty", err)
		os.Exit(1)
	}

	defer sh.Process.Kill()

	onTypedKey := func(e *fyne.KeyEvent) {
		if e.Name == fyne.KeyEnter || e.Name == fyne.KeyReturn {
			_, _ = tty.Write([]byte{'\r'})
		}
	}
	onTypedRune := func(r rune) {
		_, _ = tty.WriteString(string(r))
	}

	window.Canvas().SetOnTypedKey(onTypedKey)
	window.Canvas().SetOnTypedRune(onTypedRune)

	buffer := [][]rune{}
	reader := bufio.NewReader(tty)

	// go routine that reads from the pty
	go func() {
		line := []rune{}
		buffer = append(buffer, line)

		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				if err == io.EOF {
					return
				}

				os.Exit(1)
			}

			line = append(line, r)
			buffer[len(buffer)-1] = line
			if r == '\n' {
				// is buffer at capacity
				if len(buffer) > MaxBufferSize {
					buffer = buffer[1:] // pop first element
				}

				line = []rune{}
				buffer = append(buffer, line)
			}
		}
	}()

	// go routine to render UI
	go func() {
		for {
			time.Sleep(1 * time.Second)
			ui.SetText("")

			// iterate over the buffer's lines
			var lines string
			for _, line := range buffer {
				lines = lines + string(line)
			}

			ui.SetText(lines)
		}
	}()

	// setup the window
	windowDimensions := fyne.NewSize(420, 200)
	windowLayout := layout.NewGridWrapLayout(windowDimensions)
	content := container.New(windowLayout, ui)

	// set the window content
	window.SetContent(content)
	window.ShowAndRun()
}
