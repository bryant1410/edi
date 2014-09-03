package main

import (
	"bufio"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/nsf/gothic"
	"github.com/satran/edi/utils"
)

// Prefixes for all the widgets
const (
	shellP  string = "shell"
	editorP string = "editor"
	inputP  string = "input"
	promptP string = "prompt"
)

// Prefixes for widget variables.
const (
	inputV  string = "inputvar"
	promptV string = "promptvar"
)

const (
	default_prompt string = "» "
)

// Regex for splitting command and arguments.
var cmdRegex = regexp.MustCompile("'.+'|\".+\"|\\S+")

var cmdCount *utils.Counter

func init() {
	cmdCount = utils.NewCounter()
}

type Shell struct {
	Id        int
	ir        *gothic.Interpreter
	name      string
	editor    string
	prompt    string
	input     string
	promptVar string
	inputVar  string
	ps1       string
	col       *Col
}

// CurrentLineNo returns the current line number.
func (s *Shell) CurrentLineNo() (int, error) {
	r, err := s.ir.EvalAsString("%{0} index insert", s.editor)
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
	line, err := s.ir.EvalAsString(`%{0} get %{1}.0 %{1}.end`, s.editor, lineno)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

// CurrentLine returns the current line.
func (s *Shell) CurrentLine() string {
	line, err := s.ir.EvalAsString(`%{0} get "insert linestart" "insert lineend"`, s.editor)
	if err != nil {
		log.Println("Error fetching line", err)
		return ""
	}
	return line
}

func (s *Shell) Exec() {
	line, err := s.ir.EvalAsString(`set %{0}`, s.inputVar)
	if err != nil {
		log.Println("Error fetching prompt text", err)
		return
	}
	// Get the last line number.
	linestr, err := s.ir.EvalAsString(`%{0} index end`, s.editor)
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
	s.ir.Eval(`set %{0} ""`, s.inputVar)
}

func (s *Shell) setId(id int, lineno int, line string) error {
	curr := s.Line(lineno)
	if curr != "" {
		err := s.ir.Eval("%{0} insert %{1}.end {\n}", s.editor, lineno)
		if err != nil {
			return err
		}
		lineno++
	}

	// Inserting the clickable prompt
	if err := s.ir.Eval("%{0} insert %{1}.0 {%{2}} toggle%{3}",
		s.editor, lineno, s.ps1, id); err != nil {
		return err
	}
	// The event which handles hiding and showing the output
	if err := s.ir.Eval(`%{0} tag bind toggle%{2} <1> {edi::Toggle %{1} %{2}; break}`,
		s.editor, s.Id, id); err != nil {
		return err
	}

	// Insert the command
	if err := s.ir.Eval("%{0} insert %{1}.end {%{2}\n} cmd%{3}",
		s.editor, lineno, line, id); err != nil {
		return err
	}

	// Create the output tag
	if err := s.ir.Eval("%{0} insert %{1}.end {\n} out%{2}",
		s.editor, lineno, id); err != nil {
		return err
	}
	return s.ir.Eval("%{0} tag add cmd-%{1} %{2}.0 %{2}.end", s.editor, id, lineno+1)
}

func (s *Shell) Toggle(id int) {
	toggled, err := s.ir.EvalAsBool("%{0} tag cget out%{1} -elide", s.editor, id)
	if err != nil {
		// When initially the tag is not toggled it returns a ""
		toggled = false
	}
	if toggled {
		s.ir.Eval("%{0} tag configure out%{1} -elide false", s.editor, id)
	} else {
		s.ir.Eval("%{0} tag configure out%{1} -elide true", s.editor, id)
	}
}

func (s *Shell) run(id int, text string) {
	log.Println("Executing ", text)
	var oscmd *exec.Cmd
	parsed := cmdRegex.FindAllString(text, -1)

	for i, _ := range parsed {
		parsed[i] = strings.Trim(parsed[i], "'\"")
	}

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
	err := s.ir.Eval("%{0} insert out%{1}.last %{2%q} out%{1}", s.editor, id, line)
	if err != nil {
		log.Println(err)
	}
}
