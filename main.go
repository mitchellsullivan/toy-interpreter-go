package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FslInterpreter struct {
	vars  map[string]interface{}
	funcs map[string]interface{}
}

func NewFslInterpreter() *FslInterpreter {
	return &FslInterpreter{
		vars:  make(map[string]interface{}),
		funcs: make(map[string]interface{}),
	}
}

func (i *FslInterpreter) execFunction(functionName string, params map[string]interface{}) {
	currFunc := (i.funcs[functionName]).([]interface{})

	for _, currCommand := range currFunc {
		resolved := make(map[string]interface{})

		for k, val := range currCommand.(map[string]interface{}) {
			if k == "cmd" {
				continue
			} else if valStr, ok := val.(string); ok && valStr[0] == '#' {
				varName := valStr[1:]
				resolved[k] = i.vars[varName]
			} else if valStr, ok := val.(string); ok && valStr[0] == '$' {
				paramName := valStr[1:]
				resolved[k] = params[paramName]
			} else {
				resolved[k] = val
			}
		}

		cmdStr := currCommand.(map[string]interface{})["cmd"].(string)

		switch cmdStr {
		case "print":
			fmt.Println(resolved["value"])
		case "create":
			fallthrough
		case "update":
			id := resolved["id"].(string)
			i.vars[id] = resolved["value"]
		case "delete":
			id := resolved["id"].(string)
			delete(i.vars, id)
		case "add":
			id := resolved["id"].(string)
			operand1, operand2 := resolved["operand1"].(float64), resolved["operand2"].(float64)
			i.vars[id] = operand1 + operand2
		case "divide":
			id := resolved["id"].(string)
			operand1, operand2 := resolved["operand1"].(float64), resolved["operand2"].(float64)
			if operand2 == 0 {
				panic("cannot divide by zero")
			}
			i.vars[id] = operand1 / operand2
		default:
			if cmdStr[0] == '#' {
				functionName := cmdStr[1:]
				i.execFunction(functionName, resolved)
			}
		}
	}
}

func (i *FslInterpreter) runScript(script map[string]interface{}) {
	for k, v := range script {
		switch v.(type) {
		case []interface{}:
			i.funcs[k] = v
		default:
			i.vars[k] = v
		}
	}
	i.execFunction("init", nil)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename.json>")
		return
	}

	fileName := os.Args[1]
	script, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	var parsed interface{}

	err = json.Unmarshal(script, &parsed)
	if err != nil {
		fmt.Println("Could not unmarshall JSON")
		return
	}

	interpreter := NewFslInterpreter()

	switch parsed := parsed.(type) {
	case []interface{}:
		for _, currScript := range parsed {
			interpreter.runScript(currScript.(map[string]interface{}))
			fmt.Println()
		}
	case map[string]interface{}:
		interpreter.runScript(parsed)
	default:
		fmt.Println("JSON is not an object or array of objects")
		return
	}
}
