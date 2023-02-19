package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type FslInterpreter struct {
	symbols map[string]interface{}
}

func NewFslInterpreter() *FslInterpreter {
	return &FslInterpreter{
		symbols: make(map[string]interface{}),
	}
}

func (i *FslInterpreter) execFunction(functionName string, params map[string]interface{}) {
	currFunc := i.symbols[functionName]

	for _, currLine := range currFunc.([]interface{}) {
		evald := make(map[string]interface{})

		for k, val := range currLine.(map[string]interface{}) {
			if k == "cmd" {
				continue
			} else if valStr, ok := val.(string); ok && valStr[0] == '#' {
				varName := valStr[1:]
				evald[k] = i.symbols[varName]
			} else if valStr, ok := val.(string); ok && valStr[0] == '$' {
				paramName := valStr[1:]
				evald[k] = params[paramName]
			} else {
				evald[k] = val
			}
		}

		switch currLine.(map[string]interface{})["cmd"] {
		case "print":
			fmt.Println(evald["value"])
		case "create":
			i.symbols[evald["id"].(string)] = evald["value"]
		case "update":
			i.symbols[evald["id"].(string)] = evald["value"]
		case "delete":
			delete(i.symbols, evald["id"].(string))
		case "add":
			i.symbols[evald["id"].(string)] = evald["operand1"].(float64) + evald["operand2"].(float64)
		case "divide":
			operand2 := evald["operand2"].(float64)
			if operand2 == 0 {
				panic("cannot divide by zero")
			}
			i.symbols[evald["id"].(string)] = evald["operand1"].(float64) / operand2
		default:
			if cmdStr, ok := currLine.(map[string]interface{})["cmd"].(string); ok && cmdStr[0] == '#' {
				functionName := cmdStr[1:]
				i.execFunction(functionName, evald)
			}
		}
	}
}

func (i *FslInterpreter) runScript(script map[string]interface{}) {
	for k, v := range script {
		i.symbols[k] = v
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
