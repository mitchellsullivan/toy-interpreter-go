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
	currFunc, _ := (i.funcs[functionName]).([]interface{})

	for _, currCommand := range currFunc {
		currCommandMap, _ := currCommand.(map[string]interface{})
		resolved := make(map[string]interface{})

		for k, val := range currCommandMap {
			if k == "cmd" {
				continue
			}

			valStr, ok := val.(string)

			if ok && valStr[0] == '#' {
				varName := valStr[1:]
				resolved[k] = i.vars[varName]
			} else if ok && valStr[0] == '$' && params != nil {
				paramName := valStr[1:]
				resolved[k] = params[paramName]
			} else {
				resolved[k] = val
			}
		}

		cmdStr := currCommandMap["cmd"].(string)

		switch cmdStr {
		case "print":
			fmt.Println(resolved["value"])
		case "create":
			fallthrough
		case "update":
			id, _ := resolved["id"].(string)
			i.vars[id] = resolved["value"]
		case "delete":
			id, _ := resolved["id"].(string)
			delete(i.vars, id)
		case "add":
			id, _ := resolved["id"].(string)
			op1, _ := resolved["operand1"].(float64)
			op2, _ := resolved["operand2"].(float64)
			i.vars[id] = op1 + op2
		case "divide":
			id, _ := resolved["id"].(string)
			op1, _ := resolved["operand1"].(float64)
			op2, _ := resolved["operand2"].(float64)
			if op2 == 0 {
				panic("cannot divide by zero")
			}
			i.vars[id] = op1 / op2
		case "multiply":
			id, _ := resolved["id"].(string)
			op1, _ := resolved["operand1"].(float64)
			op2, _ := resolved["operand2"].(float64)
			i.vars[id] = op1 * op2
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
		fmt.Println("Could not unmarshal JSON")
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
