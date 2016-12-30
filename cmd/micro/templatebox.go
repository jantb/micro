package main

import (
	"github.com/zyedidia/clipboard"
	"github.com/zyedidia/tcell"
	"sort"
	"strings"
)

// TemplateBox struct
type TemplateBox struct {
	open  bool
	width int
	// Message to print
	message string
	// The user's response to a prompt
	response string
	search   string
	cursorx  int
	// style to use when drawing the message
	style tcell.Style

	// We have to keep track of the cursor for selecting
	cursory int

	messages       Messages
	messagesToshow Messages

	selected int
	template string
	selects  selects
}
type sel struct {
	text          string
	start         Loc
	end           Loc
	resultingText string
}
type selects []sel

func (s selects) Len() int {
	return len(s)
}
func (s selects) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s selects) Less(i, j int) bool {
	return s[i].text < s[j].text
}

//Open opens a box with prompt
func (a *TemplateBox) Open(v *View, template string) {
	a.open = true
	a.template = template
	v.Cursor.Relocate()
	v.Buf.Insert(v.Cursor.Loc, a.template)

	text := a.template
	a.selects = selects{}
	rem := 0
	for strings.Index(text, "$") != -1 {
		startSelect := ToCharPos(v.Cursor.Loc, v.Buf)
		start := strings.Index(text, "$")
		startSelect += start + rem
		sel := sel{}
		sel.start = FromCharPos(startSelect, v.Buf)
		text = text[start+1:]
		rem += start + 1
		end := strings.Index(text, "$")
		endSelect := startSelect + end + 2
		sel.end = FromCharPos(endSelect, v.Buf)
		sel.text = text[:end]
		text = text[end+1:]
		rem += end + 1
		a.selects = append(a.selects, sel)
	}

	startSelect := ToCharPos(v.Cursor.Loc, v.Buf)
	startSelect += strings.Index(a.template, "$0")

	sort.Sort(a.selects)
	v.Cursor.SetSelectionStart(a.selects[0].start)
	v.Cursor.SetSelectionEnd(a.selects[0].end)
	v.Cursor.Loc = a.selects[0].start
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()
}

func (a *TemplateBox) selectNext(v *View) {
	if a.selected+1 < len(a.selects) {
		if len(a.selects[a.selected].resultingText) == 0 {
			a.selects[a.selected].resultingText = v.Cursor.GetSelection()
		}
		a.selected++

		a.cursorx = 0
		if a.selected > 0 {
			for index, sele := range a.selects[a.selected:] {
				prev := a.selects[a.selected-1]
				if prev.start.X < sele.start.X {
					sele.start.X = sele.start.X + (len(prev.resultingText) - len(prev.text)) - 2
					sele.end.X = sele.end.X + (len(prev.resultingText) - len(prev.text)) - 2
					a.selects[a.selected:][index] = sele
				}
			}
		}
		a.response = ""
		sel := a.selects[a.selected]
		v.Cursor.SetSelectionStart(sel.start)
		v.Cursor.SetSelectionEnd(sel.end)
		v.Cursor.Loc = sel.start
		v.Cursor.LastVisualX = v.Cursor.GetVisualX()
	} else {
		a.Reset()
	}
}

// Reset the autocompletebox
func (a *TemplateBox) Reset() {
	a.selected = 0
	a.response = ""
	a.search = ""
	a.cursorx = 0
	a.cursory = 0
	a.open = false
	a.messages = a.messages[:0]
	a.messagesToshow = a.messagesToshow[:0]
	a.selects = a.selects[:0]
}

// HandleEvent handles an event for the prompter
func (a *TemplateBox) HandleEvent(event tcell.Event, v *View) (swallow bool) {
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyTAB:
			a.selectNext(v)
			return true
		case tcell.KeyESC:
			a.Reset()
			return true
		}
	}
	switch e := event.(type) {
	case *tcell.EventKey:
		if e.Key() != tcell.KeyRune || e.Modifiers() != 0 {
			for key, actions := range bindings {
				if e.Key() == key.keyCode {
					if e.Key() == tcell.KeyRune {
						if e.Rune() != key.r {
							continue
						}
					}
					if e.Modifiers() == key.modifiers {
						for _, action := range actions {
							funcName := FuncName(action)
							switch funcName {
							case "main.(*View).CursorLeft":
								if a.cursorx > 0 {
									a.cursorx--
								}
							case "main.(*View).CursorRight":
								if a.cursorx < Count(a.response) {
									a.cursorx++
								}
							case "main.(*View).CursorStart", "main.(*View).StartOfLine":
								a.cursorx = 0
							case "main.(*View).CursorEnd", "main.(*View).EndOfLine":
								a.cursorx = Count(a.response)
							case "main.(*View).Backspace":
								if v.Cursor.HasSelection() {
									v.Cursor.DeleteSelection()
									v.Cursor.ResetSelection()
								}
								if a.cursorx > 0 {
									a.response = string([]rune(a.response)[:a.cursorx-1]) + string([]rune(a.response)[a.cursorx:])
									a.cursorx--
								}
							case "main.(*View).Paste":
								clip, _ := clipboard.ReadAll("clipboard")
								if v.Cursor.HasSelection() {
									v.Cursor.DeleteSelection()
									v.Cursor.ResetSelection()
								}
								a.response = Insert(a.response, a.cursorx, clip)
								a.cursorx += Count(clip)
							}
						}
					}
				}
			}
		}
		switch e.Key() {
		case tcell.KeyRune:
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			a.response = Insert(a.response, a.cursorx, string(e.Rune()))
			a.cursorx++
		}
		a.selects[a.selected].resultingText = a.response
	}
	return false
}
