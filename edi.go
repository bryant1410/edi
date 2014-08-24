package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/nsf/gothic"
)

const (
	ROOT string = ".c" // The base frame where all widgets reside.
)

var SHELL_DEFAULTS = gothic.ArgMap{
	"wrap":          "word",
	"shell-bg":      "#2d2d2d",
	"shell-fg":      "#ffffff",
	"shell-cursor":  "red",
	"font":          "Menlo 12",
	"shell-padding": 5,
	"prompt-bg":     "#415F69",
	"prompt-fg":     "#ffffff",
}

// Regex for splitting command and arguments.
var cmdRegex = regexp.MustCompile("'.+'|\".+\"|\\S+")

// App defines the user interface for Edi.
type App struct {
	ir       *gothic.Interpreter
	opts     gothic.ArgMap
	shells   map[int]*Shell
}

// NewApp creates a new Tcl/Tk command and editor windows.
func NewApp(name string) (*App, error) {
	ini := fmt.Sprintf("package require Tk; wm title . %s", name)
	ir := gothic.NewInterpreter(ini)
	app := App{
		ir: ir,
		shells: make(map[int]*Shell),
	}
	app.ir.RegisterCommands("edi", &app)

	app.opts = make(gothic.ArgMap)
	updateArgs(app.opts, SHELL_DEFAULTS)

	shell, err := NewShell(ir, SHELL_DEFAULTS)
	if err != nil {
		fmt.Println("Error:", err)
	}
	app.shells[shell.Id] = shell

	return &app, nil
}

// Wait executes the Tcl/Tk event loop and waits for it to close.
func (a *App) Wait() {
	<-a.ir.Done
}

func (a *App) TCLExec(id int) {
	shell, ok := a.shells[id]
	if !ok {
		log.Println("Cant find the shell.")
		return
	}
	shell.Exec()
}

func (a *App) TCLToggle(id, cmdid int) {
	shell, ok := a.shells[id]
	if !ok {
		log.Println("Cant find the shell.")
		return
	}
	shell.Toggle(cmdid)
}

// updateArgs updates the user defined values with the default values.
func updateArgs(orig, options gothic.ArgMap) {
	for key, value := range options {
		orig[key] = value
	}
}

func main() {
	app, _ := NewApp("Edi")
	app.Wait()
}
