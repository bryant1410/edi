package main

import (
	"fmt"
	"log"

	"github.com/nsf/gothic"
	"github.com/satran/edi/utils"
)

const (
	windowP = "w"
	columnP = "col"
	WIDTH   = 800
	HEIGHT  = 600
)

// App defines the application. Windows, Columns and Shells are to be created
// from an instance of App.
type App struct {
	Windows map[int]*Window
	Shells  map[int]*Shell
	Cols    map[int]*Col

	ir *gothic.Interpreter

	winCount *utils.Counter
	shCount  *utils.Counter
	colCount *utils.Counter
}

// NewApp creates a new Application.
func NewApp() *App {
	ir := gothic.NewInterpreter("package require Tk")

	app := &App{
		Windows: make(map[int]*Window),
		Shells:  make(map[int]*Shell),
		Cols:    make(map[int]*Col),

		ir:       ir,
		winCount: utils.NewCounter(),
		shCount:  utils.NewCounter(),
		colCount: utils.NewCounter(),
	}

	err := ir.RegisterCommands("edi", app)
	if err != nil {
		log.Fatal("Error registering windows ", err)
	}

	return app
}

type Window struct {
	Id   int
	Name string
	ir   *gothic.Interpreter
}

// NewWindow creates the toplevel window
func (a *App) NewWindow() error {
	toplevel := "."
	var name string
	// When the first window is created we must use the existing toplevel .
	id := a.winCount.Inc()
	if id == 1 {
		name = fmt.Sprintf(".%d%s", id, windowP)
	} else {
		toplevel := fmt.Sprintf(".%d", id)
		err := a.ir.Eval(`toplevel .%{0}`, toplevel)
		if err != nil {
			return err
		}
		name = fmt.Sprintf(".%d.%s", id, windowP)
	}

	err := a.ir.Eval(`
		wm title %{0} {EDI %{1}}
		grid [ttk::panedwindow %{2} -orient horizontal] -column 0 -row 0 -sticky nwes
		grid columnconfigure %{0} 0 -weight 1
		grid rowconfigure %{0} 0 -weight 1
		wm geometry %{0} %{3}x%{4}
		`, toplevel, id, name, WIDTH, HEIGHT,
	)
	if err != nil {
		return err
	}

	window := &Window{
		Id:   id,
		Name: name,
		ir:   a.ir,
	}
	a.Windows[id] = window

	err = a.NewCol(window)
	if err != nil {
		return err
	}

	return nil
}

// TCLToggle toggles the shell output
func (a *App) TCLToggle(id int, cmdid int) {
	sh, ok := a.Shells[id]
	if !ok {
		return
	}
	sh.Toggle(cmdid)
}

// TCLExec executes text from the input prompt
func (a *App) TCLExec(id int) {
	sh, ok := a.Shells[id]
	if !ok {
		return
	}
	sh.Exec()
}

// TCLNew creates a new split shell
func (a *App) TCLNew(colid int) {
	col, ok := a.Cols[colid]
	if !ok {
		return
	}
	a.NewShell(col)
}

// TCLNew creates a new Column
func (a *App) TCLNewCol(colid int) {
	col, ok := a.Cols[colid]
	if !ok {
		return
	}
	window := col.Window
	a.NewCol(window)
}

// TCLContext handles right/context clicks
func (a *App) TCLContext(shellId, x, y int) {
	sh, ok := a.Shells[shellId]
	if !ok {
		return
	}
	line := sh.Selected()
	if line == "" {
		line = sh.LineUnderCursor(x, y)
	}
	log.Println(line)
}

type Col struct {
	Id     int
	Name   string
	Window *Window
}

// NewCol creates a new Column in the window
func (a *App) NewCol(window *Window) error {
	id := a.colCount.Inc()
	name := fmt.Sprintf("%s.%s%d", window.Name, columnP, id)

	options := gothic.ArgMap{
		"root": window.Name,
		"col":  name,
	}

	err := a.ir.Eval(`ttk::panedwindow %{col} -orient vertical
		%{root} add %{col} -weight 1`, options)
	if err != nil {
		return err
	}

	col := &Col{
		Id:     id,
		Name:   name,
		Window: window,
	}
	a.Cols[id] = col

	err = a.NewShell(col)
	if err != nil {
		return err
	}

	return nil
}

// NewShell creates a new Shell in the given Col
func (a *App) NewShell(c *Col) error {
	id := a.shCount.Inc()
	name := fmt.Sprintf(".%s%d", shellP, id)
	var sh Shell = Shell{
		Id:        id,
		ir:        a.ir,
		name:      name,
		editor:    fmt.Sprintf("%s.%s%d", name, editorP, id),
		prompt:    fmt.Sprintf("%s.%s%d", name, promptP, id),
		input:     fmt.Sprintf("%s.%s%d", name, inputP, id),
		promptVar: fmt.Sprintf("%s%d", promptV, id),
		inputVar:  fmt.Sprintf("%s%d", inputV, id),
		ps1:       default_prompt,
		col:       c,
	}

	var options = gothic.ArgMap{
		"wrap":           "word",
		"editor-bg":      "#FDF6E3",
		"editor-sel-bg":  "#EBE4CE",
		"editor-fg":      "#333333",
		"editor-cursor":  "red",
		"font":           "Menlo 12",
		"editor-padding": 5,
		"prompt-bg":      "#415F69",
		"prompt-fg":      "#ffffff",
		"shell":          name,
		"editor":         sh.editor,
		"prompt":         sh.prompt,
		"input":          sh.input,
		"promptvar":      sh.promptVar,
		"inputvar":       sh.inputVar,
		"id":             id,
		"ps1":            sh.ps1,
		"col":            c.Name,
		"col-id":         c.Id,
	}

	err := a.ir.Eval(newShell, options)
	if err != nil {
		return err
	}
	a.Shells[id] = &sh

	return nil
}

// updateArgs updates the user defined values with the default values.
func updateArgs(orig, options gothic.ArgMap) {
	for key, value := range orig {
		orig[key] = value
	}
}
