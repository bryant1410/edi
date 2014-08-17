package main

import (
	"fmt"
	"log"
	"bufio"
	"strconv"
	"strings"
	"os/exec"
	"regexp"

	"github.com/nsf/gothic"
)

const (
	cmdname   string = ".command" // Name used for the Tcl/Tk text view for command prompt
	prompt	string = "» "
)

// Regex for splitting command and arguments.
var cmdRegex = regexp.MustCompile("'.+'|\".+\"|\\S+")

// App defines the user interface for Edi.
type App struct {
	ir *gothic.Interpreter
}

var cmdCount = 0

// NewApp creates a new Tcl/Tk command and editor windows.
func NewApp(name string) *App {
	title := fmt.Sprintf("package require Tk; wm title . %s", name)
	ir := gothic.NewInterpreter(title)
	app := App{ir: ir}
	app.ir.RegisterCommands("edi", &app)
	err := app.newCommand(nil)
	if err != nil {
		log.Println(err)
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
		%{cname} configure -padx %{padding}
		%{cname} configure -pady %{padding}
		%{cname} configure -insertbackground %{cursor-color}
		bind %{cname} <KeyPress-Return> {edi::Exec; break}
	`, options)
	line, err := a.CurrentLineNo()
	if err != nil {
		log.Fatal("Cant find cursor.")
	}
	a.addPrompt(line)

	return err
}

// CurrentLine returns the current line number.
func (a *App) CurrentLineNo() (int, error) {
	r, err := a.ir.EvalAsString("%{0} index insert", cmdname)
	if err != nil {
		return -1, err
	}
	line, err := strconv.Atoi(strings.Split(r, ".")[0])
	if err != nil {
		return -1, err
	}
	return line, nil
}

func (a *App) Line(lineno int) string {
	line, err := a.ir.EvalAsString(`%{0} get %{1}.0 %{1}.end`, cmdname, lineno)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

// CurrentLine returns the current line.
func (a *App) CurrentLine() string {
	line, err := a.ir.EvalAsString(`%{0} get "insert linestart" "insert lineend"`, cmdname)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

func (a *App) TCL_Exec() {
	lineno, err := a.CurrentLineNo()
	if err != nil {
		log.Println("Cant fetch line.", err)
		return
	}
	if err != nil {
		log.Println("Cant fetch line.", err)
		return
	}
	line := a.CurrentLine()
	if !strings.HasPrefix(line, prompt) {
		return
	}

	cmdCount = cmdCount + 1
	err = a.setId(cmdCount, lineno)
	err = a.addPrompt(lineno + 1)
	if err != nil {
		log.Println("Cant set id.", err)
		return
	}
	line = strings.TrimLeft(line, prompt)
	go a.run(cmdCount, line)
}

func (a *App) addPrompt(line int) error{
	return a.ir.Eval("%{0} insert %{1}.0 %{2%q}", cmdname, line, prompt)
}

func (a *App) setId(id int, line int) error {
	err := a.ir.Eval("%{0} tag add cmd-%{1} %{2}.0 %{2}.end", cmdname, id, line)
	if err != nil {
		return err
	}
	return a.ir.Eval("%{0} insert %{1}.end {\n} out-%{2}", cmdname, line, id)
}

func (a *App) run(id int, text string) {
	log.Println("Executing ", text)
	var oscmd *exec.Cmd
	parsed := cmdRegex.FindAllString(text, -1)

	if len(parsed) < 1{
		return
	} else if len(parsed) > 1{
		oscmd = exec.Command(parsed[0], parsed[1:]...)
	} else {
		oscmd = exec.Command(parsed[0])
	}

	stdout, err := oscmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		a.Append(id, err.Error())
		return
	}
	stderr, err := oscmd.StderrPipe()
	if err != nil {
		log.Println(err)
		a.Append(id, err.Error())
		return
	}

	err = oscmd.Start()
	if err != nil {
		a.Append(id, err.Error())
		log.Println(err)
		return
	}

	reader := bufio.NewReader(stdout)
	readerErr := bufio.NewReader(stderr)
	go a.readAndPush(id, readerErr)
	a.readAndPush(id, reader)

	oscmd.Wait()
}

func (a *App) readAndPush(id int, reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		a.Append(id, line)
		if err != nil {
			break
		}
	}
}

func (a *App) Append(id int, line string) {
	err := a.ir.Eval("%{0} insert out-%{1}.last %{2%q} out-%{1}", cmdname, id, line)
	if err != nil {
		log.Println(err)
	}
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
		"padding": 5,
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
