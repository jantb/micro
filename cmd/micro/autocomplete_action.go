package main

import (
	"fmt"
	"os/exec"
	"strings"
)

func AutocompleteGocode(v *View) {
	_, err := exec.LookPath("gocode")
	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "github.com/nsf/gocode").CombinedOutput()
	}
	getMessages := func(v *View) (messages Messages) {
		offset := ByteOffset(v.Cursor.Loc, v.Buf)
		cmd := exec.Command("gocode", "-f=csv", "autocomplete", fmt.Sprintf("%d", offset))
		in, _ := cmd.StdinPipe()
		fmt.Fprint(in, v.Buf.String())
		in.Close()
		b, err := cmd.CombinedOutput()
		if err != nil {
			messenger.Message(fmt.Sprintf("%s %s", b, err))
			return Messages{}
		}

		messages = Messages{}
		for _, value := range strings.Split(string(b), "\n") {
			split := strings.Split(value, ",,")
			if len(split) != 3 {
				continue
			}
			messages = append(messages, Message{Searchable: split[1], MessageToDisplay: fmt.Sprintf("%s %s %s", split[0], split[1], split[2]), Value2: []byte(value)})
		}
		return messages
	}
	acceptTab := func(message Message) {
		split := strings.Split(string(message.Value2), ",,")
		t := split[0]
		val := split[1]
		def := split[2]
		if t == "func" {
			v.Cursor.Left()
			if IsWordChar(string(v.Cursor.RuneUnder(v.Cursor.X))) {
				v.Cursor.SelectWord()
				v.Cursor.DeleteSelection()
			} else {
				v.Cursor.Right()
			}

			def = def[5:strings.Index(def, ")")]
			d := []string{}
			for i, value := range strings.Split(def, ", ") {
				d = append(d, fmt.Sprintf("$%d_%s$", i, value))
			}
			text := val + "(" + strings.Join(d, ", ") + ")"

			template.Open(v, text)

			return
		}
		v.Cursor.Left()
		if IsWordChar(string(v.Cursor.RuneUnder(v.Cursor.X))) {
			v.Cursor.SelectWord()
			v.Cursor.DeleteSelection()
		} else {
			v.Cursor.Right()
		}

		v.Buf.Insert(v.Cursor.Loc, val)
		for range val {
			v.Cursor.Right()
		}
		v.Vet()
		v.Lint()
	}
	autocomplete.OpenNoPrompt(getMessages, nil, acceptTab, v)
}
func AutocompleteGlobal(v *View) {

	getMessages := func(v *View) (messages Messages) {

		word := []rune{}
		for i := 1; !IsWhitespace(v.Cursor.RuneUnder(v.Cursor.X - i)); i++ {
			word = append([]rune{v.Cursor.RuneUnder(v.Cursor.X - i)}, word...)
		}

		completions := GetCodeComplete(string(word))
		messages = Messages{}
		for _, value := range completions {
			split := strings.Split(value, ",,")
			if len(split) != 3 {
				continue
			}
			messages = append(messages, Message{Searchable: split[1], MessageToDisplay: fmt.Sprintf("%s %s %s", split[0], split[1], split[2]), Value2: []byte(value)})
		}
		return messages
	}
	acceptTab := func(message Message) {
		split := strings.Split(string(message.Value2), ",,")
		t := split[0]
		val := split[1]
		if t == "func" {
			v.Cursor.Left()
			if IsWordChar(string(v.Cursor.RuneUnder(v.Cursor.X))) {
				v.Cursor.SelectWord()
				v.Cursor.DeleteSelection()
			} else {
				v.Cursor.Right()
			}

			template.Open(v, string(val))

			return
		}
		v.Cursor.Left()
		if IsWordChar(string(v.Cursor.RuneUnder(v.Cursor.X))) {
			v.Cursor.SelectWord()
			v.Cursor.DeleteSelection()
		} else {
			v.Cursor.Right()
		}

		v.Buf.Insert(v.Cursor.Loc, val)
		for range val {
			v.Cursor.Right()
		}
		v.Vet()
		v.Lint()
	}
	autocomplete.OpenNoPrompt(getMessages, nil, acceptTab, v)
}
