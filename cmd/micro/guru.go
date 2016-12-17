package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/tools/cmd/guru/serial"
	"os/exec"
)

func getWhat(v *View) serial.What {
	v.Cursor.Relocate()
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

func getPointsto(v *View) []serial.PointsTo {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")

	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "-scope", getWhat(v).ImportPath, "pointsto", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var pointsto = make([]serial.PointsTo, 0)
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &pointsto)
	if err != nil {
		messenger.Message(string(data))
	}
	return pointsto
}
func getCallers(v *View) []serial.Caller {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")

	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "-scope", getWhat(v).ImportPath, "callers", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var caller = make([]serial.Caller, 0)
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &caller)
	if err != nil {
		messenger.Message(string(data))
	}
	return caller
}

func getCallStack(v *View) serial.CallStack {
	offset := ByteOffset(v.Cursor.Loc, v.Buf)
	_, err := exec.LookPath("guru")

	if err != nil {
		_, _ = exec.Command("go", "get", "-u", "golang.org/x/tools/cmd/...").CombinedOutput()
	}
	cmd := exec.Command("guru", "-modified", "-json", "-scope", getWhat(v).ImportPath, "callstack", fmt.Sprintf("%s:#%d", v.Buf.Path, offset))
	in, _ := cmd.StdinPipe()
	fmt.Fprint(in, v.Buf.GetName()+"\n")
	fmt.Fprintf(in, "%d\n", len(v.Buf.String()))
	fmt.Fprint(in, v.Buf.String())
	in.Close()
	data, err := cmd.CombinedOutput()
	var callstack = serial.CallStack{}
	if err != nil {
		messenger.Message(fmt.Sprintf("%s %s", data, err))
	}
	err = json.Unmarshal(data, &callstack)
	if err != nil {
		messenger.Message(string(data))
	}
	return callstack
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
