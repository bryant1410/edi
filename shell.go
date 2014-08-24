package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/nsf/gothic"
	"github.com/satran/edi/utils"
)

const (
	toplevel     string = "w"
	shPrefix     string = "shell"
	inputPrefix  string = "input"
	promptPrefix string = "prompt"
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
	toplevel   string
	ShellName  string
	PromptName string
	InputName  string
	PromptVar  string
	InputVar   string
	PS1        string
}

func NewShell(ir *gothic.Interpreter, options gothic.ArgMap, root bool) (*Shell, error) {
	id := shCount.Inc()
	var tl string
	if !root {
		tl = fmt.Sprintf(".%s%d", toplevel, id)
	}
	var s Shell = Shell{
		Id:         id,
		ir:         ir,
		toplevel:   tl,
		ShellName:  fmt.Sprintf("%s.%s", tl, shPrefix),
		PromptName: fmt.Sprintf("%s.%s", tl, promptPrefix),
		InputName:  fmt.Sprintf("%s.%s", tl, inputPrefix),
		PromptVar:  fmt.Sprintf("%s%d", promptVar, id),
		InputVar:   fmt.Sprintf("%s%d", inputVar, id),
		PS1:        default_prompt,
	}

	options["shell"] = s.ShellName
	options["prompt"] = s.PromptName
	options["input"] = s.InputName
	options["promptvar"] = s.PromptVar
	options["inputvar"] = s.InputVar
	options["id"] = id
	options["prompt-value"] = s.PS1
	options["toplevel"] = strings.Trim(s.toplevel, ".")

	if !root {
		err := ir.Eval(`tk::toplevel %{0}`, s.toplevel)
		if err != nil {
			return nil, err
		}
	}

	err := ir.Eval(newShell, options)
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
