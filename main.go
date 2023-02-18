package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

var symbols = make(map[string]interface{})

func execFunction(functionName string, params map[string]interface{}) {
	currFunc := symbols[functionName]

	for _, currLine := range currFunc.([]interface{}) {
		evald := make(map[string]interface{})

		for k, val := range currLine.(map[string]interface{}) {
			if k == "cmd" {
				continue
			} else if valStr, ok := val.(string); ok && valStr[0] == '#' {
				varName := valStr[1:]
				evald[k] = symbols[varName]
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
			symbols[evald["id"].(string)] = evald["value"]
		case "update":
			symbols[evald["id"].(string)] = evald["value"]
		case "delete":
			delete(symbols, evald["id"].(string))
		case "add":
			symbols[evald["id"].(string)] = evald["operand1"].(float64) + evald["operand2"].(float64)
		case "divide":
			operand2 := evald["operand2"].(float64)
			if operand2 == 0 {
				panic("cannot divide by zero")
			}
			symbols[evald["id"].(string)] = evald["operand1"].(float64) / operand2
		default:
			if cmdStr, ok := currLine.(map[string]interface{})["cmd"].(string); ok && cmdStr[0] == '#' {
				functionName := cmdStr[1:]
				execFunction(functionName, evald)
			}
		}
	}
}

func runScript(script map[string]interface{}) {
	for k, v := range script {
		symbols[k] = v
	}
	execFunction("init", nil)
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

	switch parsed := parsed.(type) {
	case []interface{}:
		for _, currScript := range parsed {
			runScript(currScript.(map[string]interface{}))
			fmt.Println()
		}
	case map[string]interface{}:
		runScript(parsed)
	default:
		fmt.Println("JSON is not an object or array of objects")
		return
	}
}
