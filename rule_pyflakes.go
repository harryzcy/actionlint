package actionlint

import (
	"bytes"
	"io"
	"os/exec"
	"sync"
)

type shellIsPythonKind int

const (
	shellIsPythonKindUnspecified shellIsPythonKind = iota
	shellIsPythonKindPython
	shellIsPythonKindNotPython
)

func getShellIsPythonKind(shell *String) shellIsPythonKind {
	if shell == nil {
		return shellIsPythonKindUnspecified
	}
	if shell.Value == "python" {
		return shellIsPythonKindPython
	}
	return shellIsPythonKindNotPython
}

// RulePyflakes is a rule to check Python scripts at 'run:' using pyflakes.
// https://github.com/PyCQA/pyflakes
type RulePyflakes struct {
	RuleBase
	cmd                   string
	workflowShellIsPython shellIsPythonKind
	jobShellIsPython      shellIsPythonKind
	wg                    sync.WaitGroup
	mu                    sync.Mutex
}

// NewRulePyflakes creates new RulePyflakes instance. Parameter executable can be command name
// or relative/absolute file path. When the given executable is not found in system, it returns
// an error.
func NewRulePyflakes(executable string, debug io.Writer) (*RulePyflakes, error) {
	p, err := exec.LookPath(executable)
	if err != nil {
		return nil, err
	}
	r := &RulePyflakes{
		RuleBase:              RuleBase{name: "pyflakes", dbg: debug},
		cmd:                   p,
		workflowShellIsPython: shellIsPythonKindUnspecified,
		jobShellIsPython:      shellIsPythonKindUnspecified,
	}
	return r, nil
}

// VisitJobPre is callback when visiting Job node before visiting its children.
func (rule *RulePyflakes) VisitJobPre(n *Job) error {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.jobShellIsPython = getShellIsPythonKind(n.Defaults.Run.Shell)
	}
	return nil
}

// VisitJobPost is callback when visiting Job node after visiting its children.
func (rule *RulePyflakes) VisitJobPost(n *Job) error {
	rule.jobShellIsPython = shellIsPythonKindUnspecified // reset
	return nil
}

// VisitWorkflowPre is callback when visiting Workflow node before visiting its children.
func (rule *RulePyflakes) VisitWorkflowPre(n *Workflow) error {
	if n.Defaults != nil && n.Defaults.Run != nil {
		rule.workflowShellIsPython = getShellIsPythonKind(n.Defaults.Run.Shell)
	}
	return nil
}

// VisitWorkflowPost is callback when visiting Workflow node after visiting its children.
func (rule *RulePyflakes) VisitWorkflowPost(n *Workflow) error {
	// TODO: Check errors caused in goroutines to run pyflakes and returns it

	// Wait all pyflakes processes finish
	rule.wg.Wait()
	rule.workflowShellIsPython = shellIsPythonKindUnspecified // reset

	return nil
}

// VisitStep is callback when visiting Step node.
func (rule *RulePyflakes) VisitStep(n *Step) error {
	if rule.cmd == "" {
		return nil
	}

	run, ok := n.Exec.(*ExecRun)
	if !ok || run.Run == nil {
		return nil
	}

	if !rule.isPythonShell(run) {
		return nil
	}

	rule.wg.Add(1)
	go rule.runPyflakes(rule.cmd, run.Run.Value, run.RunPos)
	return nil
}

func (rule *RulePyflakes) isPythonShell(r *ExecRun) bool {
	if r.Shell != nil {
		return r.Shell.Value == "python"
	}

	if rule.jobShellIsPython != shellIsPythonKindUnspecified {
		return rule.jobShellIsPython == shellIsPythonKindPython
	}

	return rule.workflowShellIsPython == shellIsPythonKindPython
}

func (rule *RulePyflakes) runPyflakes(executable, src string, pos *Pos) {
	defer rule.wg.Done()

	src = sanitizeExpressionsInScript(src) // Defiend at rule_shellcheck.go
	rule.debug("%s: Run pyflakes for Python script:\n%s", pos, src)

	cmd := exec.Command(executable)
	cmd.Stderr = nil
	rule.debug("%s: Running pyflakes command: %s", pos, cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		rule.debug("%s: Could not make stdin pipe: %v", pos, err)
		return
	}
	if _, err := io.WriteString(stdin, src); err != nil {
		rule.debug("%s: Could not write stdin: %v", pos, err)
		return
	}
	stdin.Close()

	b, err := cmd.Output()
	if err != nil {
		rule.debug("%s: Command %s failed: %v", pos, cmd, err)
	}
	if len(b) == 0 {
		return
	}

	rule.mu.Lock()
	defer rule.mu.Unlock()
	for len(b) > 0 {
		b = rule.parseNextError(b, pos)
	}
}

func (rule *RulePyflakes) parseNextError(output []byte, pos *Pos) []byte {
	// Eat "<stdin>:"
	idx := bytes.Index(output, []byte("<stdin>:"))
	if idx == -1 {
		rule.debug("%s: error message does not start with \"<stdin>\": %q", pos, output)
		return nil
	}
	output = output[idx+len("<stdin>:"):]

	idx = bytes.IndexByte(output, '\n')
	if idx == -1 {
		rule.debug("%s: error message does not end with \\n: %q", pos, output)
		return nil
	}
	msg := output[:idx]
	output = output[idx+1:]

	rule.errorf(pos, "pyflakes reported issue in this script: %s", msg)
	return output
}
