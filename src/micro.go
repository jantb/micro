package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/go-errors/errors"
	"github.com/mattn/go-isatty"
	"io/ioutil"
	"os"
)

const (
	tabSize      = 4  // This should be configurable
	synLinesUp   = 75 // How many lines up to look to do syntax highlighting
	synLinesDown = 75 // How many lines down to look to do syntax highlighting
)

// The main screen
var screen tcell.Screen

// LoadInput loads the file input for the editor
func LoadInput() (string, []byte, error) {
	// There are a number of ways micro should start given its input
	// 1. If it is given a file in os.Args, it should open that

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	// These are empty by default so if we get to option 3, we can just returns the
	// default values
	var filename string
	var input []byte
	var err error

	if len(os.Args) > 1 {
		// Option 1
		filename = os.Args[1]
		// Check that the file exists
		if _, e := os.Stat(filename); e == nil {
			input, err = ioutil.ReadFile(filename)
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = ioutil.ReadAll(os.Stdin)
	}

	// Option 3, or just return whatever we got
	return filename, input, err
}

func main() {
	filename, input, err := LoadInput()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Should we enable true color?
	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	// In order to enable true color, we have to set the TERM to `xterm-truecolor` when
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	// Load the syntax files, including the colorscheme
	LoadSyntaxFiles()

	// Initilize tcell
	screen, err = tcell.NewScreen()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	// This is just so if we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	defer func() {
		if err := recover(); err != nil {
			screen.Fini()
			fmt.Println("Micro encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	// Default style
	defStyle := tcell.StyleDefault.
		Foreground(tcell.ColorDefault).
		Background(tcell.ColorDefault)

	// There may be another default style defined in the colorscheme
	if style, ok := colorscheme["default"]; ok {
		defStyle = style
	}

	screen.SetStyle(defStyle)
	screen.EnableMouse()

	messenger := new(Messenger)
	view := NewView(NewBuffer(string(input), filename), messenger)

	for {
		// Display everything
		screen.Clear()

		view.Display()
		messenger.Display()

		screen.Show()

		// Wait for the user's action
		event := screen.PollEvent()

		// Check if we should quit
		switch e := event.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyCtrlQ {
				// Make sure not to quit if there are unsaved changes
				if view.CanClose("Quit anyway? ") {
					screen.Fini()
					os.Exit(0)
				}
			}
		}

		// Send it to the view
		view.HandleEvent(event)
	}
}
