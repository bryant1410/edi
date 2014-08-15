package main

import (
	"fmt"
	"strconv"
	"strings"

	_ "net" // For some reason adding this fixes the non-rendering bug in MacOS

	"github.com/nsf/gothic"
)

const (
	cmdname   string = ".command" // Name used for the Tcl/Tk text view for command prompt
)

// App defines the user interface for Edi.
type App struct {
	ir *gothic.Interpreter
}

// NewApp creates a new Tcl/Tk command and editor windows.
func NewApp(name string) *App {
	title := fmt.Sprintf("package require Tk; wm title . %s", name)
	ir := gothic.NewInterpreter(title)
	app := App{ir: ir}
	app.ir.RegisterCommands("edi", &app)
	err := app.newCommand(nil)
	if err != nil {
		fmt.Println(err)
	}
	return &app
}

// Wait executes the Tcl/Tk event loop and waits for it to close.
func (a *App) Wait() {
	<-a.ir.Done
}

// newCommand creates the command prompt view for a given Tcl/Tk Interpreter.
// Extra options to the command can be passed as arguments.
func (a *App) newCommand(options gothic.ArgMap) error {
	options = updateCommandDefaults(options)
	err := a.ir.Eval(`
		grid [text %{cname}] -row 0 -column 0 -sticky nwes
		grid columnconfigure . 0 -weight 1
		grid rowconfigure . 0 -weight 1
		%{cname} configure -highlightthickness 0 
		%{cname} configure -wrap %{wrap}
		%{cname} configure -insertofftime 0
		%{cname} configure -bg %{bg} -fg %{fg}
		%{cname} configure -font %{font%q}
		%{cname} configure -insertbackground %{cursor-color}
		bind %{cname} <KeyPress-Return> edi::Exec 
	`, options)

	return err
}

// CurrentLine returns the current line number.
func (a *App) CurrentLineNo() int {
	r, err := a.ir.EvalAsString("%{0} index", cmdname)
	if err != nil {
		return 1
	}
	line, err := strconv.Atoi(strings.Split(r, ".")[0])
	if err != nil {
		return 1
	}
	return line
}

// CurrentLine returns the current line.
func (a *App) CurrentLine() string {
	currline := a.CurrentLineNo()
	err := a.ir.Eval(`%{0} mark set insert "%{1} lineend"`, cmdname, currline)
	if err != nil {
		fmt.Println(err)
	}
	line, err := a.ir.EvalAsString(`%{0} get "insert linestart" "insert lineend"`, cmdname)
	if err != nil {
		return ""
	}
	return line
}

func (a *App) TCL_Exec() {
	line := a.CurrentLine()
	fmt.Println(line)
}

// Returns the default settings for the Command view.
func updateCommandDefaults(options gothic.ArgMap) gothic.ArgMap {
	// These are the default settings of the command.
	d := gothic.ArgMap{
		"cname":  cmdname,
		"wrap":   "word",
		"bg":     "#2d2d2d",
		"fg":     "#ffffff",
		"cursor-color": "red",
		"font":   "Menlo 12",
	}
	updateArgs(d, options)
	return d
}

func updateArgs(orig, options gothic.ArgMap) {
	for key, value := range options {
		orig[key] = value
	}
}

func main() {
	app := NewApp("Edi")
	app.Wait()
}
