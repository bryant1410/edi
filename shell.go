package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"bufio"
	"os/exec"

	"github.com/nsf/gothic"
	"github.com/satran/edi/utils"
)

const (
	shPrefix     string = ".shell"
	inputPrefix  string = ".input"
	promptPrefix string = ".prompt"
	inputVar     string = "inputvar"
	promptVar    string = "promptvar"

	default_prompt string = "» "
)

var (
	shCount  *utils.Counter
	cmdCount *utils.Counter
)

func init() {
	shCount = utils.NewCounter()
	cmdCount = utils.NewCounter()
}

type Shell struct {
	Id         int
	ir         *gothic.Interpreter
	ShellName  string
	PromptName string
	InputName  string
	PromptVar  string
	InputVar   string
	PS1	string
}

func NewShell(ir *gothic.Interpreter, options gothic.ArgMap) (*Shell, error) {
	id := shCount.Inc()
	var s Shell = Shell{
		Id:         id,
		ir:         ir,
		ShellName:  fmt.Sprintf("%s%d", shPrefix, id),
		PromptName: fmt.Sprintf("%s%d", promptPrefix, id),
		InputName:  fmt.Sprintf("%s%d", inputPrefix, id),
		PromptVar:  fmt.Sprintf("%s%d", promptVar, id),
		InputVar:   fmt.Sprintf("%s%d", inputVar, id),
		PS1:	default_prompt,
	}

	options["shell"] = s.ShellName
	options["prompt"] = s.PromptName
	options["input"] = s.InputName
	options["promptvar"] = s.PromptVar
	options["inputvar"] = s.InputVar
	options["id"] = id
	options["prompt-value"] = s.PS1

	err := ir.Eval(`
		grid [text %{shell}] -row 0 -column 0 -sticky nwes -columnspan 2
		%{shell} configure -highlightthickness 0 
		%{shell} configure -wrap %{wrap}
		%{shell} configure -insertofftime 0
		%{shell} configure -bg %{shell-bg} -fg %{shell-fg}
		%{shell} configure -font %{font%q}
		%{shell} configure -padx %{shell-padding}
		%{shell} configure -pady %{shell-padding}
		%{shell} configure -insertbackground %{shell-cursor}

		grid [label %{prompt} -text %{prompt-value%q}] -row 1 -column 0 -sticky w
		%{prompt} configure -bg %{prompt-bg} -fg %{prompt-fg}
		%{prompt} configure -padx 0
		%{prompt} configure -pady 0
		%{prompt} configure -font %{font%q}

		grid [entry %{input}] -row 1 -column 1 -sticky nwes
		%{input} configure -bg %{prompt-bg} -fg %{prompt-fg}
		%{input} configure -highlightthickness 0
		%{input} configure -insertbackground %{shell-cursor}
		%{input} configure -font %{font%q}
		%{input} configure -borderwidth 0
		%{input} configure -takefocus 1
		%{input} configure -textvariable %{inputvar}
		bind %{input} <KeyPress-Return> {edi::Exec %{id}; break}

		grid columnconfigure . 0 -weight 0
		grid rowconfigure . 0 -weight 1

		grid columnconfigure . 1 -weight 1
		grid rowconfigure . 1 -weight 0
	`, options)

	if err != nil {
		return nil, err
	}
	return &s, nil
}

// CurrentLineNo returns the current line number.
func (s *Shell) CurrentLineNo() (int, error) {
	r, err := s.ir.EvalAsString("%{0} index insert", s.ShellName)
	if err != nil {
		return -1, err
	}
	line, err := strconv.Atoi(strings.Split(r, ".")[0])
	if err != nil {
		return -1, err
	}
	return line, nil
}

func (s *Shell) Line(lineno int) string {
	line, err := s.ir.EvalAsString(`%{0} get %{1}.0 %{1}.end`, s.ShellName, lineno)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

// CurrentLine returns the current line.
func (s *Shell) CurrentLine() string {
	line, err := s.ir.EvalAsString(`%{0} get "insert linestart" "insert lineend"`, s.ShellName)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

func (s *Shell) Exec() {
	line, err := s.ir.EvalAsString(`set %{0}`, s.InputVar)
	if err != nil {
		log.Println("Error fetching prompt text", err)
		return
	}
	// Get the last line number.
	linestr, err := s.ir.EvalAsString(`%{0} index end`, s.ShellName)
	if err != nil {
		log.Println("Error fetching lineno ", err)
		return
	}
	lineno, err := strconv.Atoi(strings.Split(linestr, ".")[0])
	lineno--

	cmd := cmdCount.Inc()
	err = s.setId(cmd, lineno, line)
	if err != nil {
		log.Println("Cant set id.", err)
		return
	}
	go s.run(cmd, line)
	s.ir.Eval(`set %{0} ""`, s.InputVar)
}

func (s *Shell) setId(id int, lineno int, line string) error {
	curr := s.Line(lineno)
	if curr != "" {
		err := s.ir.Eval("%{0} insert %{1}.end {\n}", s.ShellName, lineno)
		if err != nil {
			return err
		}
		lineno++
	}

	// Inserting the clickable prompt
	if err := s.ir.Eval("%{0} insert %{1}.0 {%{2}} toggle%{3}",
		s.ShellName, lineno, s.PS1, id); err != nil {
		return err
	}
	// The event which handles hiding and showing the output
	if err := s.ir.Eval(`%{0} tag bind toggle%{2} <1> {edi::Toggle %{1} %{2}; break}`,
		s.ShellName, s.Id, id); err != nil {
		return err
	}

	// Insert the command
	if err := s.ir.Eval("%{0} insert %{1}.end {%{2}\n} cmd%{3}",
		s.ShellName, lineno, line, id); err != nil {
		return err
	}

	// Create the output tag
	if err := s.ir.Eval("%{0} insert %{1}.end {\n} out%{2}",
		s.ShellName, lineno, id); err != nil {
		return err
	}
	return s.ir.Eval("%{0} tag add cmd-%{1} %{2}.0 %{2}.end", s.ShellName, id, lineno+1)
}

func (s *Shell) Toggle(id int) {
	toggled, err := s.ir.EvalAsBool("%{0} tag cget out%{1} -elide", s.ShellName, id)
	if err != nil {
		// When initially the tag is not toggled it returns a ""
		toggled = false
	}
	if toggled {
		s.ir.Eval("%{0} tag configure out%{1} -elide false", s.ShellName, id)
	} else {
		s.ir.Eval("%{0} tag configure out%{1} -elide true", s.ShellName, id)
	}
}

func (s *Shell) run(id int, text string) {
	log.Println("Executing ", text)
	var oscmd *exec.Cmd
	parsed := cmdRegex.FindAllString(text, -1)

	if len(parsed) < 1 {
		return
	} else if len(parsed) > 1 {
		oscmd = exec.Command(parsed[0], parsed[1:]...)
	} else {
		oscmd = exec.Command(parsed[0])
	}

	stdout, err := oscmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		s.Append(id, err.Error())
		return
	}
	stderr, err := oscmd.StderrPipe()
	if err != nil {
		log.Println(err)
		s.Append(id, err.Error())
		return
	}

	err = oscmd.Start()
	if err != nil {
		s.Append(id, err.Error())
		log.Println(err)
		return
	}

	reader := bufio.NewReader(stdout)
	readerErr := bufio.NewReader(stderr)
	go s.readAndPush(id, readerErr)
	s.readAndPush(id, reader)

	oscmd.Wait()
}

func (s *Shell) readAndPush(id int, reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		s.Append(id, line)
		if err != nil {
			break
		}
	}
}

func (s *Shell) Append(id int, line string) {
	err := s.ir.Eval("%{0} insert out%{1}.last %{2%q} out%{1}", s.ShellName, id, line)
	if err != nil {
		log.Println(err)
	}
}
