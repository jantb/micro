package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/tools/cmd/guru/serial"
	"os/exec"
)

func getWhat(v *View) serial.What {
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
	var what = serial.What{}
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &what)
	if err != nil {
		messenger.Message(string(data))
	}
	return what

}

func getImplements(v *View) serial.Implements {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")
	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "implements", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var implements = serial.Implements{}
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &implements)
	if err != nil {
		messenger.Message(string(data))
	}
	return implements
}

func getDefinition(v *View) serial.Definition {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")
	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "definition", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var definition = serial.Definition{}
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &definition)
	if err != nil {
		messenger.Message(string(data))
	}
	return definition
}
