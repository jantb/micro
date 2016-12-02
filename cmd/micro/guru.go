package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// What defines the what command from guru
type What struct {
	Enclosing []struct {
		Desc  string `json:"desc"`
		Start int    `json:"start"`
		End   int    `json:"end"`
	} `json:"enclosing"`
	Modes      []string `json:"modes"`
	Srcdir     string   `json:"srcdir"`
	Importpath string   `json:"importpath"`
	Object     string   `json:"object"`
	Sameids    []string `json:"sameids"`
}

func getWhat(v *View) What {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")
	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "what", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var what = What{}
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &what)
	if err != nil {
		messenger.Message(string(data))
	}
	return what

}
