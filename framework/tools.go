package framework

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/asmaloney/gactar/actr"
	"github.com/asmaloney/gactar/amod"

	"github.com/asmaloney/gactar/util/filesystem"
	"github.com/asmaloney/gactar/util/python"
)

// Some tools for working with our frameworks

// Setup will check that the executable exists and then use it to identify itself.
func Setup(info *Info) (err error) {
	_, err = filesystem.CheckForExecutable(info.ExecutableName)
	if err != nil {
		return
	}

	err = identifyYourself(info.Name, info.ExecutableName)
	if err != nil {
		return
	}

	for _, packageName := range info.PythonRequiredPackages {
		err = python.CheckForPackage(info.ExecutableName, packageName)
		if err != nil {
			return
		}
	}

	return
}

func ParseInitialBuffers(model *actr.Model, initialBuffers InitialBuffers) (parsed ParsedInitialBuffers, err error) {
	parsed = ParsedInitialBuffers{}

	for bufferName, bufferInit := range initialBuffers {
		buffer := model.LookupBuffer(bufferName)
		if buffer == nil {
			err = fmt.Errorf("ERROR cannot initialize buffer '%s' - not found in model '%s'", bufferName, model.Name)
			return
		}

		pattern, parseErr := amod.ParseChunk(model, bufferInit)
		if parseErr != nil {
			err = fmt.Errorf("ERROR in initial buffer  '%s' - %s", bufferName, parseErr)
			return
		}

		parsed[bufferName] = pattern
	}

	return
}

func PythonValuesToStrings(values *[]*actr.Value, quoteStrings bool) []string {
	str := make([]string, len(*values))
	for i, v := range *values {
		if v.Var != nil {
			str[i] = strings.TrimPrefix(*v.Var, "?")
		} else if v.Str != nil {
			if quoteStrings {
				str[i] = fmt.Sprintf("'%s'", *v.Str)
			} else {
				str[i] = *v.Str
			}
		} else if v.Number != nil {
			str[i] = *v.Number
		}
		// v.ID should not be possible because of validation
	}

	return str
}

// identifyYourself outputs version info and the path to an executable.
func identifyYourself(frameworkName, exeName string) (err error) {
	cmd := exec.Command(exeName, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	version := strings.TrimSpace(string(output))

	cmd = exec.Command("which", exeName)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s: Using %s from %s", frameworkName, version, string(output))

	return
}
