package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/yuin/gopher-lua"
	"github.com/zyedidia/clipboard"
	"go/format"
	"io"
	"io/ioutil"
	"os/exec"
	"sort"
)

// PreActionCall executes the lua pre callback if possible
func PreActionCall(funcName string, view *View) bool {
	executeAction := true
	for pl := range loadedPlugins {
		ret, err := Call(pl+".pre"+funcName, view)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			executeAction = false
		}
	}
	return executeAction
}

// PostActionCall executes the lua plugin callback if possible
func PostActionCall(funcName string, view *View) bool {
	relocate := true
	for pl := range loadedPlugins {
		ret, err := Call(pl+".on"+funcName, view)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			relocate = false
		}
	}
	return relocate
}

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		return true
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Center", v) {
		return false
	}

	v.Topline = v.Cursor.Y - v.Height/2
	if v.Topline+v.Height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.Height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("Center", v)
	}
	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorUp", v) {
		return false
	}

	v.deselect(0)
	v.Cursor.Up()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorUp", v)
	}
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorDown", v) {
		return false
	}

	v.deselect(1)
	v.Cursor.Down()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorDown", v)
	}
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorLeft", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Left()
	}

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorLeft", v)
	}
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorRight", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1].Move(-1, v.Buf)
		v.Cursor.ResetSelection()
	} else {
		v.Cursor.Right()
	}

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorRight", v)
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordRight", v) {
		return false
	}

	v.Cursor.WordRight()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("WordRight", v)
	}
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordLeft", v) {
		return false
	}

	v.Cursor.WordLeft()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("WordLeft", v)
	}
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectUp", v)
	}
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectDown", v)
	}
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLeft", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectLeft", v)
	}
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectRight", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End().Move(-1, v.Buf)
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectRight", v)
	}
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordRight", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectWordRight", v)
	}
	return true
}

//Template completion
func (v *View) Template(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Template", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		template.Open(v, "")
	}

	if usePlugin {
		return PostActionCall("Template", v)
	}
	return true
}

//ExtractVariable creates a local variable
func (v *View) ExtractVariable(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Template", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		what := getWhat(v)
		if what.Enclosing[0].Description == "identifier" {
			desc := getDescription(v)
			start := FromByteOffset(what.Enclosing[0].Start, v.Buf)
			end := FromByteOffset(what.Enclosing[0].End, v.Buf)
			identifier := v.Buf.Substr(start, end)
			if desc.Detail != "type" && strings.HasPrefix(desc.Value.Type, "func") {
				for _, value := range what.Enclosing {
					if strings.HasPrefix(value.Description, "function") {
						start = FromByteOffset(value.Start, v.Buf)
						end = FromByteOffset(value.End, v.Buf)
						identifier = v.Buf.Substr(start, end)
						v.Buf.Remove(start, end)
						break
					}
				}
				split := strings.Split(desc.Value.Type, ") (")
				if len(split) == 2 {
					ret := split[1][:len(split[1])-1]
					split = strings.Split(ret, ",")
					vars := []string{}
					for i, value := range split {
						vars = append(vars, fmt.Sprintf("$%d_%s$", i, value))
					}
					template.Open(v, fmt.Sprintf("%s := %s", strings.Join(vars, ", "), identifier))
					return usePlugin
				}
			}

			v.Buf.Remove(start, end)
			template.Open(v, fmt.Sprintf("$0_identifier$ := %s", identifier))
		} else if what.Enclosing[0].Description == "basic literal" {
			start := FromByteOffset(what.Enclosing[0].Start, v.Buf)
			end := FromByteOffset(what.Enclosing[0].End, v.Buf)
			identifier := v.Buf.Substr(start, end)
			v.Buf.Remove(start, end)
			template.Open(v, fmt.Sprintf("$0_identifier$ := %s", identifier))
		} else {
			TermMessage("@" + what.Enclosing[0].Description + "@")
		}

	}

	if usePlugin {
		return PostActionCall("Template", v)
	}
	return true
}

// SelectWordRight selects the word under the cursor
func (v *View) SelectWord(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWord", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		what := getWhat(v)
		selection := v.Cursor.CurSelection

		expandToNext := false
		for _, enclosing := range what.Enclosing {
			locStart := FromByteOffset(enclosing.Start, v.Buf)
			locEnd := FromByteOffset(enclosing.End, v.Buf)
			if !v.Cursor.HasSelection() {
				//TermMessage(fmt.Sprintf("%s %s %s", locStart, locEnd, v.Cursor.Loc))
				v.Cursor.SetSelectionStart(locStart)
				v.Cursor.SetSelectionEnd(locEnd)
				continue
			}
			if expandToNext {
				v.Cursor.SetSelectionStart(locStart)
				v.Cursor.SetSelectionEnd(locEnd)
				break
			}
			if selection[0] == locStart && selection[1] == locEnd {
				expandToNext = true
			}
		}
	}

	if usePlugin {
		return PostActionCall("SelectWord", v)
	}
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordLeft", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectWordLeft", v)
	}
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("StartOfLine", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Start()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("StartOfLine", v)
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("EndOfLine", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.End()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("EndOfLine", v)
	}
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStartOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectToStartOfLine", v)
	}
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEndOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectToEndOfLine", v)
	}
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorStart", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0
	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorStart", v)
	}
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorEnd", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()
	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("CursorEnd", v)
	}
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStart", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart(false)
	v.Cursor.SelectTo(v.Buf.Start())

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectToStart", v)
	}
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEnd", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd(false)
	v.Cursor.SelectTo(v.Buf.End())

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectToEnd", v)
	}
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertSpace", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")
	v.Cursor.Right()

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("InsertSpace", v)
	}
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertNewline", v) {
		return false
	}

	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	v.Buf.Insert(v.Cursor.Loc, "\n")
	ws := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
	v.Cursor.Right()

	if v.Buf.Settings["autoindent"].(bool) {
		v.Buf.Insert(v.Cursor.Loc, ws)
		for i := 0; i < len(ws); i++ {
			v.Cursor.Right()
		}

		// Remove the whitespaces if keepautoindent setting is off
		if IsSpacesOrTabs(v.Buf.Line(v.Cursor.Y-1)) && !v.Buf.Settings["keepautoindent"].(bool) {
			line := v.Buf.Line(v.Cursor.Y - 1)
			v.Buf.Remove(Loc{0, v.Cursor.Y - 1}, Loc{Count(line), v.Cursor.Y - 1})
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("InsertNewline", v)
	}
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Backspace", v) {
		return false
	}

	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc.GreaterThan(v.Buf.Start()) {
		// We have to do something a bit hacky here because we want to
		// delete the line by first moving left and then deleting backwards
		// but the undo redo would place the cursor in the wrong place
		// So instead we move left, save the position, move back, delete
		// and restore the position

		// If the user is using spaces instead of tabs and they are deleting
		// whitespace at the start of the line, we should delete as if it's a
		// tab (tabSize number of spaces)
		lineStart := v.Buf.Line(v.Cursor.Y)[:v.Cursor.X]
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		if v.Buf.Settings["tabstospaces"].(bool) && IsSpaces(lineStart) && len(lineStart) != 0 && len(lineStart)%tabSize == 0 {
			loc := v.Cursor.Loc
			v.Cursor.Loc = loc.Move(-tabSize, v.Buf)
			cx, cy := v.Cursor.X, v.Cursor.Y
			v.Cursor.Loc = loc
			v.Buf.Remove(loc.Move(-tabSize, v.Buf), loc)
			v.Cursor.X, v.Cursor.Y = cx, cy
		} else {
			v.Cursor.Left()
			cx, cy := v.Cursor.X, v.Cursor.Y
			v.Cursor.Right()
			loc := v.Cursor.Loc
			v.Buf.Remove(loc.Move(-1, v.Buf), loc)
			v.Cursor.X, v.Cursor.Y = cx, cy
		}
	}
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("Backspace", v)
	}
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordRight", v) {
		return false
	}

	v.SelectWordRight(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("DeleteWordRight", v)
	}
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordLeft", v) {
		return false
	}

	v.SelectWordLeft(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("DeleteWordLeft", v)
	}
	return true
}

// Delete deletes the next character
func (v *View) Delete(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Delete", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		loc := v.Cursor.Loc
		if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("Delete", v)
	}
	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("IndentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		startY := v.Cursor.CurSelection[0].Y
		endY := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for y := startY; y <= endY; y++ {
			tabsize := len(v.Buf.IndentString())
			v.Buf.Insert(Loc{0, y}, v.Buf.IndentString())
			if y == startY && v.Cursor.CurSelection[0].X > 0 {
				v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(tabsize, v.Buf))
			}
			if y == endY {
				v.Cursor.SetSelectionEnd(Loc{endX + tabsize + 1, endY})
			}
		}
		v.Cursor.Relocate()

		if usePlugin {
			return PostActionCall("IndentSelection", v)
		}
		return true
	}

	go v.What(usePlugin)

	return false
}

// OutdentLine moves the current line back one indentation
func (v *View) OutdentLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	for x := 0; x < len(v.Buf.IndentString()); x++ {
		if len(GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))) == 0 {
			break
		}
		v.Buf.Remove(Loc{0, v.Cursor.Y}, Loc{1, v.Cursor.Y})
		v.Cursor.X--
	}
	v.Cursor.Relocate()

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("OutdentLine", v)
	}
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		startY := v.Cursor.CurSelection[0].Y
		endY := v.Cursor.CurSelection[1].Move(-1, v.Buf).Y
		endX := v.Cursor.CurSelection[1].Move(-1, v.Buf).X
		for y := startY; y <= endY; y++ {
			for x := 0; x < len(v.Buf.IndentString()); x++ {
				if len(GetLeadingWhitespace(v.Buf.Line(y))) == 0 {
					break
				}
				v.Buf.Remove(Loc{0, y}, Loc{1, y})
				if y == startY && v.Cursor.CurSelection[0].X > 0 {
					v.Cursor.SetSelectionStart(v.Cursor.CurSelection[0].Move(-1, v.Buf))
				}
				if y == endY {
					v.Cursor.SetSelectionEnd(Loc{endX - x, endY})
				}
			}
		}
		v.Cursor.Relocate()

		go v.What(usePlugin)

		if usePlugin {
			return PostActionCall("OutdentSelection", v)
		}
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertTab", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	for i := 0; i < bytesUntilIndent; i++ {
		v.Cursor.Right()
	}

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("InsertTab", v)
	}
	return true
}

// Save the buffer to disk
func (v *View) Save(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Save", v) {
		return false
	}

	if v.Type == vtHelp {
		// We can't save the help text
		return false
	}
	// If this is an empty buffer, ask for a filename
	if v.Buf.Path == "" {
		v.SaveAs(false)
	}
	err := v.Buf.Save()
	if err != nil {
		if strings.HasSuffix(err.Error(), "permission denied") {
			choice, _ := messenger.YesNoPrompt("Permission denied. Do you want to save this file using sudo? (y,n)")
			if choice {
				err = v.Buf.SaveWithSudo()
				if err != nil {
					messenger.Error(err.Error())
					return false
				}
				messenger.Message("Saved " + v.Buf.Path)
			}
			messenger.Reset()
			messenger.Clear()
		} else {
			messenger.Error(err.Error())
		}
	} else {
		messenger.Message("Saved " + v.Buf.Path)
	}

	if usePlugin {
		return PostActionCall("Save", v)
	}
	return false
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs(usePlugin bool) bool {
	filename, canceled := messenger.Prompt("Filename: ", "", "Save", NoCompletion)
	if !canceled {
		// the filename might or might not be quoted, so unquote first then join the strings.
		filename = strings.Join(SplitCommandArgs(filename), " ")
		v.Buf.Path = filename
		v.Buf.name = filename

		v.Save(true)
	}

	return false
}

// Find opens a prompt and searches forward for the input
func (v *View) Find(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Find", v) {
		return false
	}

	searchStr := ""
	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		searchStr = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	BeginSearch(searchStr)

	if usePlugin {
		return PostActionCall("Find", v)
	}
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindNext", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[1], v.Buf)
		lastSearch = v.Cursor.GetSelection()
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	if lastSearch == "" {
		return true
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	if usePlugin {
		return PostActionCall("FindNext", v)
	}
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious(usePlugin bool) bool {
	if usePlugin && !PreActionCall("FindPrevious", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = ToCharPos(v.Cursor.CurSelection[0], v.Buf)
	} else {
		searchStart = ToCharPos(v.Cursor.Loc, v.Buf)
	}
	messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	if usePlugin {
		return PostActionCall("FindPrevious", v)
	}
	return true
}

// Undo undoes the last action
func (v *View) Undo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Undo", v) {
		return false
	}

	v.Buf.Undo()
	messenger.Message("Undid action")
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("Undo", v)
	}
	return true
}

// Redo redoes the last action
func (v *View) Redo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Redo", v) {
		return false
	}

	v.Buf.Redo()
	messenger.Message("Redid action")
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("Redo", v)
	}
	return true
}

// Copy the selection to the system clipboard
func (v *View) Copy(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Copy", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.CopySelection("clipboard")
		v.freshClip = true
		messenger.Message("Copied selection")
	}

	if usePlugin {
		return PostActionCall("Copy", v)
	}
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CutLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	if v.freshClip == true {
		if v.Cursor.HasSelection() {
			if clip, err := clipboard.ReadAll("clipboard"); err != nil {
				messenger.Error(err)
			} else {
				clipboard.WriteAll(clip+v.Cursor.GetSelection(), "clipboard")
			}
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || v.freshClip == false {
		v.Copy(true)
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Cut line")
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("CutLine", v)
	}
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Cut", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.CopySelection("clipboard")
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true
		messenger.Message("Cut selection")

		if usePlugin {
			return PostActionCall("Cut", v)
		}
		return true
	}
	v.Vet()
	v.Lint()
	return false
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DuplicateLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		v.Cursor.Right()
	}

	messenger.Message("Duplicated line")
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("DuplicateLine", v)
	}
	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	messenger.Message("Deleted line")

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("DeleteLine", v)
	}
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (v *View) MoveLinesUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesUp", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[0].Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.CurSelection[0].Y,
			v.Cursor.CurSelection[1].Y,
		)
		v.Cursor.UpN(1)
		v.Cursor.CurSelection[0].Y--
		v.Cursor.CurSelection[1].Y--
		messenger.Message("Moved up selected line(s)")
	} else {
		if v.Cursor.Loc.Y == 0 {
			messenger.Message("Can not move further up")
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.UpN(1)
		messenger.Message("Moved up current line")
	}
	v.Buf.IsModified = true

	go v.What(usePlugin)
	v.Vet()
	v.Lint()
	if usePlugin {
		return PostActionCall("MoveLinesUp", v)
	}
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (v *View) MoveLinesDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesDown", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[1].Y >= len(v.Buf.lines) {
			messenger.Message("Can not move further down")
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.CurSelection[0].Y,
			v.Cursor.CurSelection[1].Y,
		)
		v.Cursor.DownN(1)
		v.Cursor.CurSelection[0].Y++
		v.Cursor.CurSelection[1].Y++
		messenger.Message("Moved down selected line(s)")
	} else {
		if v.Cursor.Loc.Y >= len(v.Buf.lines)-1 {
			messenger.Message("Can not move further down")
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
		v.Cursor.DownN(1)
		messenger.Message("Moved down current line")
	}
	v.Buf.IsModified = true

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("MoveLinesDown", v)
	}
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("clipboard")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// PastePrimary pastes from the primary clipboard (only use on linux)
func (v *View) PastePrimary(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip, _ := clipboard.ReadAll("primary")
	v.paste(clip)

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectAll", v) {
		return false
	}

	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	go v.What(usePlugin)

	if usePlugin {
		return PostActionCall("SelectAll", v)
	}
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OpenFile", v) {
		return false
	}

	if v.CanClose() {
		input, canceled := messenger.Prompt("> ", "open ", "Open", CommandCompletion)
		if !canceled {
			HandleCommand(input)

			if usePlugin {
				return PostActionCall("OpenFile", v)
			}
		}
	}
	return false
}

// GotoFile open autocomplete for file
func (v *View) GotoFile(usePlugin bool) bool {
	if usePlugin && !PreActionCall("GotoFile", v) {
		return false
	}
	autocomplete.Open(func(v *View) (messages Messages) {
		files := getFilesInCurrentDir()
		for _, file := range files {
			b, _ := json.Marshal(file)
			message := Message{Searchable: file.Name, MessageToDisplay: fmt.Sprintf("%s (%s)", file.Name, file.Path), Value2: b}
			messages = append(messages, message)
		}

		sort.Sort(messages)
		return messages
	}, func(message Message) {
		var f File
		json.Unmarshal(message.Value2, &f)
		v.Buf.Save()
		v.Open(f.Path)
		cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})

		// Move cursor and view if possible.
		if message.Extra.line+1 < v.Buf.NumLines && message.Extra.line-1 >= 0 {
			v.Cursor.Y = message.Extra.line - 1
			v.Relocate()
		}
	}, nil, v)

	if usePlugin {
		return PostActionCall("GotoFile", v)
	}
	return true
}

// GotoGutterMesssage goes to guttermessage
func (v *View) GotoGutterMesssage(usePlugin bool) bool {
	if usePlugin && !PreActionCall("GotoGutterMesssage", v) {
		return false
	}
	linenum := len(v.Buf.lines) + 1
	for _, guttermessages := range v.messages {
		for _, message := range guttermessages {
			if v.Buf.Cursor.Y < message.lineNum {
				linenum = Min(linenum, message.lineNum)
			}
		}
	}
	if linenum < len(v.Buf.lines)+1 {
		v.Cursor.Y = linenum
	} else {
		v.Cursor.Y = 0
	}

	if usePlugin {
		return PostActionCall("GotoGutterMesssage", v)
	}
	return true
}

// Start moves the viewport to the start of the buffer
func (v *View) Start(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Start", v) {
		return false
	}

	v.Topline = 0

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("Start", v)
	}
	return false
}

// Vet checks for errors
func (v *View) Vet() {
	if v.Buf.FileType() == "go" {
		_, err := exec.LookPath("goimports")
		if err != nil {
			_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
		}
		cmd := exec.Command("goimports")
		in, _ := cmd.StdinPipe()
		fmt.Fprint(in, v.Buf.String())
		in.Close()
		data, err := cmd.CombinedOutput()
		if err != nil {
			v.ClearGutterMessages("VetErrors")
			errors := strings.Split(string(data), "\n")
			for _, value := range errors {
				if strings.TrimSpace(value) == "" {
					continue
				}
				split := strings.Split(value, ":")
				x, _ := strconv.Atoi(split[1])
				if len(split) > 3 {
					v.GutterMessage("VetErrors", x, split[3], GutterError)
				}
			}
		} else {
			v.ClearGutterMessages("VetErrors")
		}
	}
}

// Lint checks for errors
func (v *View) Lint() {
	if v.Buf.FileType() == "go" {
		_, err := exec.LookPath("golint")
		if err != nil {
			_, _ = exec.Command("go", "get", "-u", "github.com/golang/lint/golint").CombinedOutput()
		}
		f, _ := ioutil.TempFile("", "lint")
		defer os.Remove(f.Name())
		ioutil.WriteFile(f.Name(), []byte(v.Buf.String()), 0644)
		cmd := exec.Command("golint", "-set_exit_status", f.Name())
		data, err := cmd.CombinedOutput()
		if err != nil {
			v.ClearGutterMessages("LintErrors")
			errors := strings.Split(string(data), "\n")

			for _, value := range errors[:len(errors)-2] {
				if strings.TrimSpace(value) == "" {
					continue
				}
				x, _ := strconv.Atoi(strings.Split(value, ":")[1])
				v.GutterMessage("LintErrors", x, strings.Split(value, ":")[3], GutterWarning)
			}
		} else {
			v.ClearGutterMessages("LintErrors")
		}
	}
}

// Autocomplete box
func (v *View) Autocomplete(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Autocomplete", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		if !autocomplete.open {
			AutocompleteGocode(v)
		} else {
			AutocompleteGlobal(v)
		}
	}
	return true
}

// Rename opens a prompt and renames
func (v *View) Rename(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Rename", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		response, _ := messenger.Prompt("Rename:", "", "", NoCompletion)
		offset := ByteOffset(v.Cursor.Loc, v.Buf)
		_, err := exec.LookPath("gorename")
		if err != nil {
			_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
		}

		data, err := exec.Command("gorename", "--offset", fmt.Sprintf("%s:#%d", v.Buf.Path, offset), "--to", response).CombinedOutput()
		if err != nil {
			messenger.Message(fmt.Sprintf("%s %s", data, err))
			return true
		}
		v.ReOpen()
		messenger.Message(string(data))
	}
	if usePlugin {
		return PostActionCall("Rename", v)
	}
	return true
}

// Suggest show suggestion box
func (v *View) Suggest(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Suggest", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		what := getWhat(v)
		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			for _, mode := range what.Modes {
				messages = append(messages, Message{MessageToDisplay: strings.Title(mode), Value2: []byte(mode)})
			}
			return messages
		}, func(message Message) {
			switch string(message.Value2) {
			case "definition":
				v.Definition(false)
			case "describe":
				v.Describe(false)
			case "implements":
				v.Implements(false)
			case "pointsto":
				v.PointsTo(false)
			case "referrers":
				v.Referrers(false)
			case "callers":
				v.Callers(false)
			case "callstack":
				v.CallStack(false)
			}
		}, nil, v)
	}
	if usePlugin {
		return PostActionCall("Suggest", v)
	}
	return true
}

// CallStack show path from callgraph root to selected function
func (v *View) CallStack(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Callers", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		callstack := getCallStack(v)
		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			for _, caller := range callstack.Callers {
				messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s %s (%s)", caller.Desc, caller.Caller, caller.Pos), Value2: []byte(caller.Pos)})
			}
			return messages
		}, func(message Message) {
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
			v.Buf.Save()
			v.Open(strings.Split(string(message.Value2), ":")[0])
			x, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[2])
			y, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[1])
			v.Buf.Cursor.X = x - 1
			v.Buf.Cursor.Y = y - 1
			v.Relocate()
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
		}, nil, v)
	}
	if usePlugin {
		return PostActionCall("Callers", v)
	}
	return true
}

// Callers show possible targets of selected function call
func (v *View) Callers(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Callers", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		callers := getCallers(v)
		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			for _, p := range callers {
				messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s %s (%s)", p.Caller, p.Desc, p.Pos), Value2: []byte(p.Pos)})
			}
			return messages
		}, func(message Message) {
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
			v.Buf.Save()
			v.Open(strings.Split(string(message.Value2), ":")[0])
			x, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[2])
			y, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[1])
			v.Buf.Cursor.X = x - 1
			v.Buf.Cursor.Y = y - 1
			v.Relocate()
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
		}, nil, v)
	}
	if usePlugin {
		return PostActionCall("Callers", v)
	}
	return true
}

// PointsTo show variables the selected pointer may point to
func (v *View) PointsTo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Implements", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		pointsto := getPointsto(v)
		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			for _, p := range pointsto {
				for _, l := range p.Labels {
					messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s %s", p.Type, l.Desc), Value2: []byte(l.Pos)})
				}
			}
			return messages
		}, func(message Message) {
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
			v.Buf.Save()
			v.Open(strings.Split(string(message.Value2), ":")[0])
			x, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[2])
			y, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[1])
			v.Buf.Cursor.X = x - 1
			v.Buf.Cursor.Y = y - 1
			v.Relocate()
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
		}, nil, v)
	}
	if usePlugin {
		return PostActionCall("Implements", v)
	}
	return true
}

// Implements gives list of interfaces
func (v *View) Implements(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Implements", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		implements := getImplements(v)
		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			for _, from := range implements.AssignableFrom {
				messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s %s", from.Name, from.Kind), Value2: []byte(from.Pos)})
			}
			return messages
		}, func(message Message) {
			v.Buf.Save()
			v.Open(strings.Split(string(message.Value2), ":")[0])
			x, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[2])
			y, _ := strconv.Atoi(strings.Split(string(message.Value2), ":")[1])
			v.Buf.Cursor.X = x - 1
			v.Buf.Cursor.Y = y - 1
			v.Relocate()
			cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
		}, nil, v)
	}
	if usePlugin {
		return PostActionCall("Implements", v)
	}
	return true
}

// What highlight under cursor
func (v *View) What(usePlugin bool) bool {
	if usePlugin && !PreActionCall("What", v) {
		return false
	}

	if v.Buf.FileType() == "go" {
		what := getWhat(v)
		highlight := [][]Loc{}
		if len(what.Enclosing) == 0 {
			v.SetHighLight(&highlight)
			return true
		}
		if what.Enclosing[0].Description != "identifier" {
			v.SetHighLight(&highlight)
			return true
		}
		highlight = append(highlight, []Loc{FromByteOffset(what.Enclosing[0].Start, v.Buf), FromByteOffset(what.Enclosing[0].End, v.Buf)})
		for _, ids := range what.SameIDs {
			split := strings.Split(ids, ":")
			y, _ := strconv.Atoi(split[1])
			x, _ := strconv.Atoi(split[2])
			offset := ByteOffset(Loc{X: x - 1, Y: y - 1}, v.Buf)
			highlight = append(highlight, []Loc{FromByteOffset(offset, v.Buf), FromByteOffset(offset+ByteOffset(highlight[0][1], v.Buf)-ByteOffset(highlight[0][0], v.Buf), v.Buf)})
		}
		v.SetHighLight(&highlight)
	}
	if usePlugin {
		return PostActionCall("What", v)
	}
	return true
}

// Definition show declaration of selected identifier
func (v *View) Definition(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Definition", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		definition := getDefinition(v)
		v.Buf.Save()
		v.Open(strings.Split(definition.ObjPos, ":")[0])
		x, _ := strconv.Atoi(strings.Split(definition.ObjPos, ":")[2])
		y, _ := strconv.Atoi(strings.Split(definition.ObjPos, ":")[1])
		v.Buf.Cursor.X = x - 1
		v.Buf.Cursor.Y = y - 1
		v.Relocate()
		cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})

		go v.What(usePlugin)
	}
	if usePlugin {
		return PostActionCall("Definition", v)
	}
	return true
}

// Describe describe selected syntax: definition, methods, etc
func (v *View) Describe(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Describe", v) {
		return false
	}
	type Info struct {
		Desc   string `json:"desc"`
		Pos    string `json:"pos"`
		Detail string `json:"detail"`
		Value  struct {
			Type   string `json:"type"`
			Objpos string `json:"objpos"`
		} `json:"value"`
	}
	if v.Buf.FileType() == "go" {
		description := getDescription(v)

		autocomplete.OpenNoPrompt(func(v *View) (messages Messages) {
			messages = Messages{}
			if description.Detail == "value" {
				messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s %s", strings.TrimSpace(description.Desc), description.Value.Type)})
			}
			return messages
		}, nil, nil, v)
	}
	if usePlugin {
		return PostActionCall("Describe", v)
	}
	return true
}

// Referrers show all refs to entity denoted by selected identifier
func (v *View) Referrers(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Referrers", v) {
		return false
	}
	type Loc []struct {
		Objpos  string `json:"objpos,omitempty"`
		Desc    string `json:"desc,omitempty"`
		Package string `json:"package,omitempty"`
		Refs    []struct {
			Pos  string `json:"pos"`
			Text string `json:"text"`
		} `json:"refs,omitempty"`
	}
	if v.Buf.FileType() == "go" {
		offset := ByteOffset(v.Cursor.Loc, v.Buf)
		_, err := exec.LookPath("goimports")
		if err != nil {
			_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
		}
		cmd := exec.Command("guru", "-modified", "-json", "referrers", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
		in, _ := cmd.StdinPipe()
		fmt.Fprint(in, v.Buf.GetName()+"\n")
		fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
		fmt.Fprint(in, v.Buf.String())
		in.Close()
		data, err := cmd.CombinedOutput()
		if err != nil {
			messenger.Message(fmt.Sprintf("%s %s", data, err))
			return true
		}

		data = []byte(fmt.Sprintf("[%s]", strings.Replace(string(data), "}\n{", "},{", -1)))

		var loc = Loc{}
		err = json.Unmarshal(data, &loc)
		if err != nil {
			TermMessage(string(data))
			messenger.Message(err)
		} else {
			autocomplete.Open(func(v *View) (messages Messages) {
				messages = Messages{}
				loc = loc[1:]
				for _, refs := range loc {
					for _, ref := range refs.Refs {
						messages = append(messages, Message{MessageToDisplay: fmt.Sprintf("%s (%s)", strings.TrimSpace(ref.Text), ref.Pos), Value2: []byte(ref.Pos)})
					}
				}
				return messages
			}, func(message Message) {
				l := string(message.Value2)
				v.Buf.Save()
				v.Open(strings.Split(l, ":")[0])
				x, _ := strconv.Atoi(strings.Split(l, ":")[2])
				y, _ := strconv.Atoi(strings.Split(l, ":")[1])
				v.Buf.Cursor.X = x - 1
				v.Buf.Cursor.Y = y - 1

				v.Relocate()
				go v.What(usePlugin)
			}, nil, v)
		}
	}
	if usePlugin {
		return PostActionCall("Referrers", v)
	}
	return true
}

// Format file
func (v *View) Format(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Format", v) {
		return false
	}
	if v.Buf.FileType() == "go" {
		data, err := format.Source([]byte(v.Buf.String()))
		if err != nil {
			messenger.Message(fmt.Sprintf("%s %s", data, err))
			return true
		}
		v.Buf.ApplyDiff(string(data))
		v.Lint()
		v.Vet()
	}
	if v.Buf.FileType() == "json" {
		var j bytes.Buffer
		err := json.Indent(&j, []byte(v.Buf.String()), "", "\t")
		v.Cursor.X = 0
		v.Cursor.Y = 0
		if err != nil {
			messenger.Message(err)
		} else {
			v.Buf.ApplyDiff(j.String())
		}
	}
	if v.Buf.FileType() == "xml" {
		b := &bytes.Buffer{}
		decoder := xml.NewDecoder(bytes.NewReader([]byte(v.Buf.String())))
		encoder := xml.NewEncoder(b)
		encoder.Indent("", "  ")
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				encoder.Flush()
				break
			}
			if err != nil {
				messenger.Message(err)
				b.Reset()
				b.WriteString(v.Buf.String())
				break
			}
			err = encoder.EncodeToken(token)
			if err != nil {
				messenger.Message(err)
				b.Reset()
				b.WriteString(v.Buf.String())
				break
			}
		}
		v.Cursor.X = 0
		v.Cursor.Y = 0
		v.Buf.ApplyDiff(b.String())
	}

	if usePlugin {
		return PostActionCall("Format", v)
	}
	return true
}

// End moves the viewport to the end of the buffer
func (v *View) End(usePlugin bool) bool {
	if usePlugin && !PreActionCall("End", v) {
		return false
	}

	if v.Height > v.Buf.NumLines {
		v.Topline = 0
	} else {
		v.Topline = v.Buf.NumLines - v.Height
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("End", v)
	}
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PageUp", v) {
		return false
	}

	if v.Topline > v.Height {
		v.ScrollUp(v.Height)
	} else {
		v.Topline = 0
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("PageUp", v)
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PageDown", v) {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.Height) > v.Height {
		v.ScrollDown(v.Height)
	} else if v.Buf.NumLines >= v.Height {
		v.Topline = v.Buf.NumLines - v.Height
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("PageDown", v)
	}
	return false
}

// NextLoc moves char to next place
func (v *View) NextLoc(usePlugin bool) bool {
	if usePlugin && !PreActionCall("NextLoc", v) {
		return false
	}
	next := cursorLocations.GetNext()
	if next.Path != "" {
		v.Buf.Save()
		v.Open(next.Path)
	}
	v.Buf.Cursor.X = next.X
	v.Buf.Cursor.Y = next.Y
	v.Buf.Cursor.Relocate()
	if usePlugin {
		return PostActionCall("NextLoc", v)
	}
	return true
}

// PrevLoc moves char to prev place
func (v *View) PrevLoc(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PrevLoc", v) {
		return false
	}
	prev := cursorLocations.GetPrev()
	if prev.Path != "" {
		v.Buf.Save()
		v.Open(prev.Path)
	}
	v.Buf.Cursor.X = prev.X
	v.Buf.Cursor.Y = prev.Y
	v.Buf.Cursor.Relocate()
	if usePlugin {
		return PostActionCall("PrevLoc", v)
	}
	return true
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageUp", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
	}
	v.Cursor.UpN(v.Height)

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("CursorPageUp", v)
	}
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageDown", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
	}
	v.Cursor.DownN(v.Height)

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("CursorPageDown", v)
	}
	return true
}

// HalfPageUp scrolls the view up half a page
func (v *View) HalfPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("HalfPageUp", v) {
		return false
	}

	if v.Topline > v.Height/2 {
		v.ScrollUp(v.Height / 2)
	} else {
		v.Topline = 0
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("HalfPageUp", v)
	}
	return false
}

// HalfPageDown scrolls the view down half a page
func (v *View) HalfPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("HalfPageDown", v) {
		return false
	}

	if v.Buf.NumLines-(v.Topline+v.Height) > v.Height/2 {
		v.ScrollDown(v.Height / 2)
	} else {
		if v.Buf.NumLines >= v.Height {
			v.Topline = v.Buf.NumLines - v.Height
		}
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("HalfPageDown", v)
	}
	return false
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleRuler", v) {
		return false
	}

	if v.Buf.Settings["ruler"] == false {
		v.Buf.Settings["ruler"] = true
		messenger.Message("Enabled ruler")
	} else {
		v.Buf.Settings["ruler"] = false
		messenger.Message("Disabled ruler")
	}

	if usePlugin {
		return PostActionCall("ToggleRuler", v)
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpLine", v) {
		return false
	}

	// Prompt for line number
	linestring, canceled := messenger.Prompt("Jump to line # ", "", "LineNumber", NoCompletion)
	if canceled {
		return false
	}
	lineint, err := strconv.Atoi(linestring)
	lineint = lineint - 1 // fix offset
	if err != nil {
		messenger.Error(err) // return errors
		return false
	}
	// Move cursor and view if possible.
	if lineint < v.Buf.NumLines && lineint >= 0 {
		v.Cursor.X = 0
		v.Cursor.Y = lineint

		cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
		if usePlugin {
			return PostActionCall("JumpLine", v)
		}
		return true
	}
	messenger.Error("Only ", v.Buf.NumLines, " lines to jump")
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ClearStatus", v) {
		return false
	}

	messenger.Message("")

	if usePlugin {
		return PostActionCall("ClearStatus", v)
	}
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleHelp", v) {
		return false
	}

	if v.Type != vtHelp {
		// Open the default help
		v.openHelp("help")
	} else {
		v.Quit(true)
	}

	if usePlugin {
		return PostActionCall("ToggleHelp", v)
	}
	return true
}

// ShellMode opens a terminal to run a shell command
func (v *View) ShellMode(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ShellMode", v) {
		return false
	}

	input, canceled := messenger.Prompt("$ ", "", "Shell", NoCompletion)
	if !canceled {
		// The true here is for openTerm to make the command interactive
		HandleShellCommand(input, true, true)
		if usePlugin {
			return PostActionCall("ShellMode", v)
		}
	}
	return false
}

// CommandMode lets the user enter a command
func (v *View) CommandMode(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CommandMode", v) {
		return false
	}

	input, canceled := messenger.Prompt("> ", "", "Command", CommandCompletion)
	if !canceled {
		HandleCommand(input)
		if usePlugin {
			return PostActionCall("CommandMode", v)
		}
	}

	return false
}

// Escape leaves current mode / quits the editor
func (v *View) Escape(usePlugin bool) bool {
	// check if user is searching, or the last search is still active
	if searching || lastSearch != "" {
		ExitSearch(v)
		return true
	}
	// check if a prompt is shown, hide it and don't quit
	if messenger.hasPrompt {
		messenger.Reset() // FIXME
		return true
	}
	return v.Quit(usePlugin)
}

// Quit quits the editor
// This behavior needs to be changed and should really only quit the editor if this
// is the last view
// However, since micro only supports one view for now, it doesn't really matter
func (v *View) Quit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Quit", v) {
		return false
	}

	// Make sure not to quit if there are unsaved changes
	if v.CanClose() {
		v.CloseBuffer()
		if len(tabs[curTab].views) > 1 {
			v.splitNode.Delete()
			tabs[v.TabNum].Cleanup()
			tabs[v.TabNum].Resize()
		} else if len(tabs) > 1 {
			if len(tabs[v.TabNum].views) == 1 {
				tabs = tabs[:v.TabNum+copy(tabs[v.TabNum:], tabs[v.TabNum+1:])]
				for i, t := range tabs {
					t.SetNum(i)
				}
				if curTab >= len(tabs) {
					curTab--
				}
				if curTab == 0 {
					// CurView().Resize(screen.Size())
					CurView().ToggleTabbar()
					CurView().matches = Match(CurView())
				}
			}
		} else {
			if usePlugin {
				PostActionCall("Quit", v)
			}

			screen.Fini()
			os.Exit(0)
		}
	}

	if usePlugin {
		return PostActionCall("Quit", v)
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("QuitAll", v) {
		return false
	}

	closeAll := true
	for _, tab := range tabs {
		for _, v := range tab.views {
			if !v.CanClose() {
				closeAll = false
			}
		}
	}

	if closeAll {
		for _, tab := range tabs {
			for _, v := range tab.views {
				v.CloseBuffer()
			}
		}

		if usePlugin {
			PostActionCall("QuitAll", v)
		}

		screen.Fini()
		os.Exit(0)
	}

	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("AddTab", v) {
		return false
	}

	tab := NewTabFromView(NewView(NewBuffer(strings.NewReader(""), "")))
	tab.SetNum(len(tabs))
	tabs = append(tabs, tab)
	curTab = len(tabs) - 1
	if len(tabs) == 2 {
		for _, t := range tabs {
			for _, v := range t.views {
				v.ToggleTabbar()
			}
		}
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("AddTab", v)
	}
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PreviousTab", v) {
		return false
	}

	if curTab > 0 {
		curTab--
	} else if curTab == 0 {
		curTab = len(tabs) - 1
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("PreviousTab", v)
	}
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("NextTab", v) {
		return false
	}

	if curTab < len(tabs)-1 {
		curTab++
	} else if curTab == len(tabs)-1 {
		curTab = 0
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("NextTab", v)
	}
	return false
}

// VSplitBinding opens an empty vertical split
func (v *View) VSplitBinding(usePlugin bool) bool {
	if usePlugin && !PreActionCall("VSplit", v) {
		return false
	}

	v.VSplit(NewBuffer(strings.NewReader(""), ""))

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("VSplit", v)
	}
	return false
}

// HSplitBinding opens an empty horizontal split
func (v *View) HSplitBinding(usePlugin bool) bool {
	if usePlugin && !PreActionCall("HSplit", v) {
		return false
	}

	v.HSplit(NewBuffer(strings.NewReader(""), ""))

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("HSplit", v)
	}
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (v *View) Unsplit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Unsplit", v) {
		return false
	}

	curView := tabs[curTab].CurView
	for i := len(tabs[curTab].views) - 1; i >= 0; i-- {
		view := tabs[curTab].views[i]
		if view != nil && view.Num != curView {
			view.Quit(true)
			// messenger.Message("Quit ", view.Buf.Path)
		}
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})

	if usePlugin {
		return PostActionCall("Unsplit", v)
	}
	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("NextSplit", v) {
		return false
	}

	tab := tabs[curTab]
	if tab.CurView < len(tab.views)-1 {
		tab.CurView++
	} else {
		tab.CurView = 0
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("NextSplit", v)
	}
	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PreviousSplit", v) {
		return false
	}

	tab := tabs[curTab]
	if tab.CurView > 0 {
		tab.CurView--
	} else {
		tab.CurView = len(tab.views) - 1
	}

	cursorLocations.AddLocation(CursorLocation{X: v.Buf.Cursor.X, Y: v.Buf.Cursor.Y, Path: v.Buf.Path})
	if usePlugin {
		return PostActionCall("PreviousSplit", v)
	}
	return false
}

var curMacro []interface{}
var recordingMacro bool

// ToggleMacro toggles recording of a macro
func (v *View) ToggleMacro(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ToggleMacro", v) {
		return false
	}

	recordingMacro = !recordingMacro

	if recordingMacro {
		curMacro = []interface{}{}
		messenger.Message("Recording")
	} else {
		messenger.Message("Stopped recording")
	}

	if usePlugin {
		return PostActionCall("ToggleMacro", v)
	}
	return true
}

// PlayMacro plays back the most recently recorded macro
func (v *View) PlayMacro(usePlugin bool) bool {
	if usePlugin && !PreActionCall("PlayMacro", v) {
		return false
	}

	for _, action := range curMacro {
		switch t := action.(type) {
		case rune:
			// Insert a character
			if v.Cursor.HasSelection() {
				v.Cursor.DeleteSelection()
				v.Cursor.ResetSelection()
			}
			v.Buf.Insert(v.Cursor.Loc, string(t))
			v.Cursor.Right()

			for pl := range loadedPlugins {
				_, err := Call(pl+".onRune", string(t), v)
				if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
					TermMessage(err)
				}
			}
		case func(*View, bool) bool:
			t(v, true)
		}
	}

	if usePlugin {
		return PostActionCall("PlayMacro", v)
	}
	return true
}

// None is no action
func None() bool {
	return false
}
