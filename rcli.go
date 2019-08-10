package rcli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type typeChecker func(value string) (result interface{}, err error)

func stringChecker(value string) (interface{}, error) {
	return value, nil
}

func boolChecker(value string) (interface{}, error) {
	return strconv.ParseBool(value)
}

func intChecker(value string) (interface{}, error) {
	return strconv.Atoi(value)
}

func floatChecker(value string) (interface{}, error) {
	return strconv.ParseFloat(value, 0)
}

var (
	typeCheckers = map[string]typeChecker{
		"bool":  boolChecker,
		"int":   intChecker,
		"float": floatChecker,
	}
	argRegexp = regexp.MustCompile(`<(\w+)(:(\w+))?(=([\w.]+))?>`)
)

type arg struct {
	name         string
	checker      typeChecker
	defaultValue interface{}
}

func (a *arg) isOptional() bool {
	return a.defaultValue != nil
}

func newArgFromRegexpMatch(match []string) *arg {
	name := match[1]
	rawType := match[3]
	rawDefaultValue := match[5]

	var (
		checker typeChecker
		ok      bool
	)
	if checker, ok = typeCheckers[rawType]; !ok {
		checker = stringChecker
	}

	var defaultValue interface{} = nil
	if rawDefaultValue != "" {
		var err error
		defaultValue, err = checker(rawDefaultValue)
		if err != nil {
			panic(fmt.Sprintf("invalid default value for type %s: %s", rawType, rawDefaultValue))
		}
	}

	return &arg{
		name:         name,
		checker:      checker,
		defaultValue: defaultValue,
	}
}

type command struct {
	name    string
	usage   string
	args    []*arg
	handler reflect.Value
}

func (c *command) minArgCount() int {
	minArgCount := 0
	for _, arg := range c.args {
		if !arg.isOptional() {
			minArgCount++
		}
	}
	return minArgCount
}

func (c *command) printUsage(program string) {
	_, _ = fmt.Fprintf(os.Stderr, "usage: %s %s\n", program, c.usage)
}

type App struct {
	commands []*command
}

func NewApp() *App {
	return &App{
		commands: make([]*command, 0),
	}
}

func (a *App) Command(usage string, handler interface{}) {
	if reflect.TypeOf(handler).Kind() != reflect.Func {
		panic(fmt.Sprintf("handler must be a func, not %T", handler))
	}

	argMatches := argRegexp.FindAllStringSubmatch(usage, -1)
	args := make([]*arg, len(argMatches))
	hasOptional := false
	for i, argMatch := range argMatches {
		args[i] = newArgFromRegexpMatch(argMatch)
		if hasOptional && !args[i].isOptional() {
			panic("optional args must come after non-optional args")
		}
		if args[i].isOptional() {
			hasOptional = true
		}
	}

	name := strings.Split(usage, " ")[0]
	if strings.HasPrefix(name, "<") {
		panic("first part of usage must be command name")
	}

	a.commands = append(a.commands, &command{
		name:    name,
		usage:   usage,
		args:    args,
		handler: reflect.ValueOf(handler),
	})
}

func (a *App) printUsage(program string) {
	newline := ""
	if len(a.commands) > 1 {
		newline = "\n"
	}
	_, _ = fmt.Fprintf(os.Stderr, "usage:%s", newline)
	for _, cmd := range a.commands {
		_, _ = fmt.Fprintf(os.Stderr, " %s %s\n", program, cmd.usage)
	}
}

func (a *App) Run(args []string) {
	program, args := args[0], args[1:]
	program = filepath.Base(program)

	if len(args) < 1 {
		a.printUsage(program)
		os.Exit(1)
	}

	name := args[0]
	var targetCmd *command
	for _, cmd := range a.commands {
		if cmd.name == name {
			targetCmd = cmd
		}
	}
	if targetCmd == nil {
		a.printUsage(program)
		os.Exit(1)
	}

	// -1 for command name
	if len(args)-1 < targetCmd.minArgCount() {
		targetCmd.printUsage(program)
		os.Exit(1)
	}

	values := make([]reflect.Value, len(targetCmd.args))
	for i, arg := range targetCmd.args {
		var (
			value interface{}
			err   error
		)

		if !arg.isOptional() || len(args)-2 >= i {
			value, err = arg.checker(args[i+1])
		} else {
			value = arg.defaultValue
		}

		if err != nil {
			targetCmd.printUsage(program)
			os.Exit(1)
		}

		values[i] = reflect.ValueOf(value)
	}

	targetCmd.handler.Call(values)
}
