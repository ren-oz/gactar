package pyactr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gitlab.com/asmaloney/gactar/actr"
)

type PyACTR struct {
	model     *actr.Model
	className string
	tmpPath   string
}

// Initialize will check for python3 and the ccm package, and create a tmp dir to save files for running.
// Note that this directory is not currently created in the proper place - it should end up in the OS's
// tmp directory. It is created locally so we can look at and debug the generated python files.
func Initialize() (p *PyACTR, err error) {
	_, err = checkForExecutable("python3")
	if err != nil {
		return
	}

	identifyPython()

	err = checkForPackage("ccm")
	if err != nil {
		return
	}

	p = &PyACTR{tmpPath: "tmp"}

	err = os.MkdirAll(p.tmpPath, os.ModePerm)
	if err != nil {
		return
	}

	return
}

// SetModel sets our model and saves the python class name we are going to use.
func (p *PyACTR) SetModel(model *actr.Model) (err error) {
	if model.Name == "" {
		err = fmt.Errorf("model is missing name")
		return
	}

	p.model = model
	p.className = fmt.Sprintf("gactar_%s", strings.Title(p.model.Name))

	return
}

// Run generates the python code from the amod file, writes it to disk, creates a "run" file
// to actually run the model, and returns the output (stdout and stderr combined).
func (p *PyACTR) Run(initialGoal string) (output []byte, err error) {
	_, err = p.WriteModel(p.tmpPath)
	if err != nil {
		return
	}

	runFile, err := p.writeRunFile(p.tmpPath, initialGoal)
	if err != nil {
		return
	}

	cmd := exec.Command("python3", runFile)

	output, err = cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("%s", string(output))
		return
	}

	return
}

// WriteModel converts the internal actr.Model to python and writes it to a file.
func (p *PyACTR) WriteModel(path string) (outputFileName string, err error) {
	outputFileName = fmt.Sprintf("%s.py", p.className)
	if path != "" {
		outputFileName = fmt.Sprintf("%s/%s", path, outputFileName)
	}

	f, err := os.Create(outputFileName)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString("# This file is generated by gactar\n")
	f.WriteString(fmt.Sprintf("# %s\n", p.model.Description))
	f.WriteString("\n")

	imports := []string{"ACTR"}

	if len(p.model.Buffers) > 0 {
		imports = append(imports, "Buffer")
	}

	if len(p.model.Memories) > 0 {
		imports = append(imports, "Memory")
	}

	if len(p.model.TextOutputs) > 0 {
		imports = append(imports, "TextOutput")
	}

	f.WriteString(fmt.Sprintf("from ccm.lib.actr import %s\n\n\n", strings.Join(imports, ", ")))

	f.WriteString(fmt.Sprintf("class %s(ACTR):\n", p.className))

	for _, buf := range p.model.Buffers {
		f.WriteString(fmt.Sprintf("\t%s = Buffer()\n", buf.Name))
	}

	for _, memory := range p.model.Memories {
		additionalInit := []string{}

		if memory.Latency != nil {
			additionalInit = append(additionalInit, fmt.Sprintf("latency=%v", *memory.Latency))
		}

		if memory.Threshold != nil {
			additionalInit = append(additionalInit, fmt.Sprintf("threshold=%v", *memory.Threshold))
		}

		if memory.MaxTime != nil {
			additionalInit = append(additionalInit, fmt.Sprintf("maximum_time=%v", *memory.MaxTime))
		}

		if memory.FinstSize != nil {
			additionalInit = append(additionalInit, fmt.Sprintf("finst_size=%v", *memory.FinstSize))
		}

		if memory.FinstTime != nil {
			additionalInit = append(additionalInit, fmt.Sprintf("finst_time=%v", *memory.FinstTime))
		}

		if len(additionalInit) > 0 {
			f.WriteString(fmt.Sprintf("\t%s = Memory(%s, %s)\n", memory.Name, memory.Buffer.Name, strings.Join(additionalInit, ", ")))
		} else {
			f.WriteString(fmt.Sprintf("\t%s = Memory(%s)\n", memory.Name, memory.Buffer.Name))
		}
	}

	for _, textOutput := range p.model.TextOutputs {
		f.WriteString(fmt.Sprintf("\t%s = TextOutput()\n", textOutput.Name))
	}

	f.WriteString("\n")

	if p.model.Logging {
		f.WriteString("\tdef __init__(self):\n")
		f.WriteString("\t\tsuper().__init__(log=True)\n")
		f.WriteString("\n")
	}

	if len(p.model.Initializers) > 0 {
		f.WriteString("\tdef init():\n")

		for _, init := range p.model.Initializers {
			f.WriteString(fmt.Sprintf("\t\t%s.add(%s)\n", init.MemoryName, init.Text))
		}

		f.WriteString("\n")
	}

	for _, production := range p.model.Productions {
		f.WriteString(fmt.Sprintf("\tdef %s(", production.Name))

		numMatches := len(production.Matches)
		for i, match := range production.Matches {
			outputMatch(f, match)

			if i != numMatches-1 {
				f.WriteString(", ")
			}
		}

		f.WriteString("):\n")

		if production.DoPython != nil {
			for _, doItem := range production.DoPython {
				f.WriteString(fmt.Sprintf("\t\t%s", doItem))
			}
		} else if production.DoStatements != nil {
			for _, statement := range production.DoStatements {
				outputStatement(f, statement)
			}
		}

		f.WriteString("\n")
	}

	return
}

func outputMatch(f *os.File, match *actr.Match) {
	if match.Text != nil {
		f.WriteString(fmt.Sprintf("%s=%s", match.Name, *match.Text))
		return
	}

	f.WriteString(fmt.Sprintf("%s='%s'", match.Name, *match.Pattern))
}

func outputStatement(f *os.File, s *actr.Statement) {
	if s.Set != nil {
		var text string
		if s.Set.Text != nil {
			text = *s.Set.Text
		} else if s.Set.Pattern != nil {
			text = "'" + s.Set.Pattern.String() + "'"
		}

		if s.Set.Field != nil {
			if s.Set.Field.ArgNum != nil {
				f.WriteString(fmt.Sprintf("\t\t%s.modify(_%d=%s)\n", s.Set.BufferName, *s.Set.Field.ArgNum, text))
			} else if s.Set.Field.FieldName != nil {
				f.WriteString(fmt.Sprintf("\t\t%s.modify(%s=%s)\n", s.Set.BufferName, *s.Set.Field.FieldName, text))
			}
		} else {
			f.WriteString(fmt.Sprintf("\t\t%s.set(%s)\n", s.Set.BufferName, text))
		}
	} else if s.Recall != nil {
		f.WriteString(fmt.Sprintf("\t\t%s.request('%s')\n", s.Recall.MemoryName, s.Recall.Pattern))
	} else if s.Clear != nil {
		for _, name := range s.Clear.BufferNames {
			f.WriteString(fmt.Sprintf("\t\t%s.clear()\n", name))
		}
	} else if s.Print != nil {
		f.WriteString(fmt.Sprintf("\t\tprint(%s)\n", strings.Join(s.Print.Args, ",")))
	} else if s.Write != nil {
		f.WriteString(fmt.Sprintf("\t\t%s.write(%s)\n", s.Write.TextOutputName, strings.Join(s.Write.Args, ",")))
	}
}

func identifyPython() {
	cmd := exec.Command("python3", "--version")
	output, _ := cmd.CombinedOutput()

	version := strings.TrimSpace(string(output))

	cmd = exec.Command("which", "python3")
	output, _ = cmd.CombinedOutput()

	fmt.Printf("Using %s from %s", version, string(output))
}

func checkForExecutable(exe string) (path string, err error) {
	path, err = exec.LookPath(exe)
	if err != nil {
		err = fmt.Errorf("cannot find '%s' in your path", exe)
		return
	}

	return
}

func checkForPackage(packageName string) (err error) {
	importCmd := fmt.Sprintf("import %s", packageName)

	cmd := exec.Command("python3", "-c", importCmd)

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("python package '%s' not found. Please ensure it is installed with pip or is in your PYTHONPATH env variable", packageName)
		return
	}

	return
}

func (p *PyACTR) writeRunFile(path, initialGoal string) (outputFileName string, err error) {
	outputFileName = fmt.Sprintf("%s_Run.py", p.className)
	if path != "" {
		outputFileName = fmt.Sprintf("%s/%s", path, outputFileName)
	}

	f, err := os.Create(outputFileName)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString("# This file is generated by gactar\n\n")

	f.WriteString(fmt.Sprintf("from %s import %s\n\n\n", p.className, p.className))

	f.WriteString(fmt.Sprintf("model = %s()\n", p.className))
	f.WriteString(fmt.Sprintf("model.goal.set('%s')\n", initialGoal))
	f.WriteString("model.run()\n")

	return
}
