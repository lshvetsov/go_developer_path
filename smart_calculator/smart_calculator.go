package smart_calculator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	about             = "The program calculates a sum(+), subtraction(-), multiplication(*) and division(-) of numbers supporting parenthesis"
	invalidAssignment = "Invalid assignment"
	invalidIdentifier = "Invalid identifier"
	unknownVariable   = "Unknown variable"
	invalidExpression = "Invalid expression"
)

var memory = make(map[string]int)

func Run() {

	fmt.Println("Enter a command, expression or /help for help:")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		str := strings.TrimSpace(scanner.Text())
		switch {
		case strings.Contains(str, "/exit"):
			fmt.Println("Bye!")
			return
		case strings.Contains(str, "/help"):
			fmt.Println(about)
		case str == "":
		case strings.HasPrefix(str, "/"):
			fmt.Println("Unknown command")
		case strings.ContainsAny(str, "+-/*()"):
			postfixString, err := infixToPostfix(str)
			result, err := calculate(postfixString)
			if err != nil {
				fmt.Println(invalidExpression)
			} else {
				fmt.Println(result)
			}
		case strings.IndexFunc(str, unicode.IsLetter) == 0:
			processVariables(str)
		default:
			fmt.Println(invalidExpression)
		}
	}
}

func calculate(postfix []string) (int, error) {
	var stack []int
	for _, token := range postfix {
		switch token {
		case "+":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a+b)
		case "-":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a-b)
		case "*":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a*b)
		case "/":
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			stack = append(stack, a/b)
		default:
			num, err := resolve(token)
			if err != nil {
				return 0, err
			}
			stack = append(stack, num)
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}
	return stack[0], nil
}

func infixToPostfix(str string) ([]string, error) {
	str = strings.ReplaceAll(str, " ", "")
	re := regexp.MustCompile(`\d+|[a-zA-Z]+|\+|\-|\*|\/|\(|\)`)
	tokens := re.FindAllString(str, -1)

	var stack []string
	var postfix []string

	for i, token := range tokens {
		switch token {
		case "+", "-":
			if i == 0 || tokens[i-1] == "(" || tokens[i-1] == "+" || tokens[i-1] == "-" || tokens[i-1] == "*" || tokens[i-1] == "/" {
				// Unary plus or minus
				postfix = append(postfix, "0") // Append a zero to the postfix
				stack = append(stack, token)   // And treat the unary operator as a binary one
			} else {
				// Binary plus or minus
				for len(stack) > 0 && (stack[len(stack)-1] == "+" || stack[len(stack)-1] == "-" || stack[len(stack)-1] == "*" || stack[len(stack)-1] == "/") {
					postfix = append(postfix, stack[len(stack)-1])
					stack = stack[:len(stack)-1]
				}
				stack = append(stack, token)
			}
		case "*", "/":
			for len(stack) > 0 && (stack[len(stack)-1] == "*" || stack[len(stack)-1] == "/") {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		case "(":
			stack = append(stack, token)
		case ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				postfix = append(postfix, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) > 0 && stack[len(stack)-1] == "(" {
				stack = stack[:len(stack)-1]
			} else {
				return nil, fmt.Errorf(invalidExpression)
			}
		default:
			postfix = append(postfix, token)
		}
	}

	for len(stack) > 0 {
		postfix = append(postfix, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}

	return postfix, nil
}

func resolve(str string) (int, error) {
	if strings.IndexFunc(str, unicode.IsLetter) == 0 {
		if val, ok := memory[str]; ok {
			return val, nil
		}
		return 0, fmt.Errorf(unknownVariable)
	} else {
		return strconv.Atoi(str)
	}
}

func processVariables(str string) {
	str = strings.Replace(str, " ", "", -1)
	fields := strings.Split(str, "=")
	length := len(fields)
	if length < 1 || length > 2 {
		fmt.Println(invalidAssignment)
		return
	}
	matched, err := regexp.MatchString("^[a-zA-Z]+$", fields[0])
	if err != nil || matched == false {
		fmt.Println(invalidIdentifier)
		return
	}
	switch length {
	case 1:
		if val, ok := memory[fields[0]]; ok {
			fmt.Println(val)
		} else {
			fmt.Println(unknownVariable)
		}
		return
	case 2:
		key := fields[0]
		value, err := strconv.Atoi(fields[1])
		if err != nil {
			if val, ok := memory[fields[1]]; ok {
				memory[key] = val
			} else {
				fmt.Println(invalidIdentifier)
			}
			return
		} else {
			memory[key] = value
		}
	}
}
