package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	exitKey = "/exit"
	helpKey = "/help"
)

var priotiritisedOperators = [][]byte{
	{'*', '/'},
	{'+', '-'},
}

type Expression struct {
	varName   string
	operands  []int
	operators []byte
}

type Calculator struct {
	vars map[string]int
	Expression
}

func main() {
	calculator := NewCalculator()
	for {
		expression, err := getExpression()

		if err != nil {
			log.Fatalln(err)
		}
		if expression == "" {
			continue
		}

		calculator.reset()

		if err = calculator.handleExpression(expression); err != nil {
			fmt.Println(err)
		}
	}
}

func NewCalculator() *Calculator {
	return &Calculator{Expression: NewExpression(), vars: make(map[string]int)}
}

func NewExpression() Expression {
	return Expression{operands: make([]int, 0), operators: make([]byte, 0)}
}

func (c *Calculator) reset() {
	c.varName = ""
	c.operands = []int{}
	c.operators = []byte{}
}

func (c *Calculator) handleExpression(expression string) error {
	for {
		subExp := regexp.MustCompile(`\([^\(\)]+\)`).FindString(expression)
		if subExp == "" {
			break
		}

		line := strings.Replace(subExp, "(", "", 1)
		line = strings.Replace(line, ")", "", 1)

		if n, _, err := c.calculate(prepareExpression(line), true); err != nil {
			return err
		} else {
			expression = strings.Replace(expression, subExp, strconv.Itoa(n), 1)
		}
	}

	expression = prepareExpression(expression)

	if expression == "" {
		return nil
	}

	if expression[0] == '/' {
		return c.handleCommand(expression)
	}

	if n, print, err := c.calculate(expression, false); err != nil {
		return err
	} else if print {
		fmt.Println(n)
	}

	return nil
}

func (c *Calculator) calculate(expression string, subExpression bool) (int, bool, error) {
	if err := c.parseExpression(expression); err != nil {
		return 0, false, err
	}

	if len(c.operators) == 0 {
		if subExpression {
			return 0, false, errors.New("Invalid expression")
		}
		result := c.operands[0]
		c.reset()
		return result, true, nil
	} else if c.operators[0] == '=' {
		if subExpression {
			return 0, false, errors.New("Invalid expression")
		}
		c.vars[c.varName] = c.operands[0]
		c.reset()
		return 0, false, nil
	} else {
		for _, ops := range priotiritisedOperators {
			for {
				i := findAny(ops, c.operators)
				if -1 == i {
					break
				}

				op := c.operators[i]

				var result int

				switch op {
				case '*':
					result = c.operands[i] * c.operands[i+1]
				case '/':
					result = c.operands[i] / c.operands[i+1]
				case '+':
					result = c.operands[i] + c.operands[i+1]
				case '-':
					result = c.operands[i] - c.operands[i+1]
				}

				c.operators = removeByIndexByte(c.operators, i)
				c.operands = removeByIndexInt(c.operands, i+1)
				c.operands[i] = result
			}

			if len(c.operators) == 0 {
				break
			}
		}

		result := c.operands[0]
		c.reset()
		return result, true, nil
	}
}

func (c *Calculator) handleCommand(cmd string) error {
	switch cmd {
	case helpKey:
		fmt.Println("This program calculates arithmetic expressions (addition and subtraction)")
	case exitKey:
		fmt.Println("Bye!")
		os.Exit(0)
	default:
		return errors.New("Unknown command")
	}
	return nil
}

func (c *Calculator) getValue(variable string) (int, error) {
	if isNumeric(variable) {
		n, _ := strconv.Atoi(variable)
		return n, nil
	} else if !isValidVariable(variable) {
		return 0, errors.New("Invalid identifier")
	} else if n, ok := c.vars[variable]; ok {
		return n, nil
	} else {
		return 0, errors.New("Unknown variable")
	}
}

func (c *Calculator) parseExpression(expression string) error {
	if regexp.MustCompile(`[\*\/]{2,}`).MatchString(expression) {
		return errors.New("Invalid expression")
	}
	if strings.Count(expression, "(") != strings.Count(expression, ")") {
		return errors.New("Invalid expression")
	}
	parts := strings.Split(expression, " ")

	if len(parts) == 1 {
		if n, err := c.getValue(parts[0]); err != nil {
			return err
		} else {
			c.operands = append(c.operands, n)
			return nil
		}
	}

	if strings.Contains(expression, "=") {
		if len(parts) != 3 || parts[1] != "=" {
			return errors.New("Invalid expression")
		} else if !isValidVariable(parts[0]) {
			return errors.New("Invalid identifier")
		} else if n, err := c.getValue(parts[2]); err != nil {
			return err
		} else {
			c.varName = parts[0]
			c.operands = append(c.operands, n)
			c.operators = append(c.operators, '=')
			return nil
		}
	}

	for i, v := range parts {
		if i%2 == 0 {
			if n, err := c.getValue(v); err != nil {
				return err
			} else {
				c.operands = append(c.operands, n)
			}
		} else {
			if v[0] == '-' && len(v)%2 == 0 {
				c.operators = append(c.operators, '+')
			} else {
				c.operators = append(c.operators, v[0])
			}
		}
	}

	return nil
}

func getExpression() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	line, err := reader.ReadString('\n')

	if err != nil {
		return "", err
	}

	return prepareExpression(line), nil
}

func findAny(needle []byte, haystack []byte) int {
	for i, h := range haystack {
		for _, n := range needle {
			if h == n {
				return i
			}
		}
	}
	return -1
}

func prepareExpression(line string) string {
	line = strings.TrimSpace(line)
	if len(line) > 0 && line[0] == '/' {
		return line
	}
	for strings.Contains(line, "  ") {
		line = strings.ReplaceAll(line, "  ", " ")
	}

	reg := regexp.MustCompile(`([\+\=\*\/]+)([a-zA-Z0-9\)]+)`)
	line = reg.ReplaceAllString(line, "${1} ${2}")
	reg = regexp.MustCompile(`([a-zA-Z0-9\)]+)([\+\-\=\*\/]+)`)
	line = reg.ReplaceAllString(line, "${1} ${2}")
	reg = regexp.MustCompile(`([\-]{2,})([a-zA-Z0-9\)]+)`)
	line = reg.ReplaceAllString(line, "${1} ${2}")

	line = regexp.MustCompile(`\(\s+`).ReplaceAllString(line, "(")
	line = regexp.MustCompile(`\s+\)`).ReplaceAllString(line, ")")

	return line
}

func isValidVariable(str string) bool {
	return regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(str)
}

func isNumeric(str string) bool {
	return regexp.MustCompile(`^-?[0-9]+$`).MatchString(str)
}

func removeByIndexInt(slice []int, index int) []int {
	switch index {
	case 0:
		return slice[1:]
	case len(slice) - 1:
		return slice[:len(slice)-1]
	default:
		return append(slice[:index], slice[index+1:]...)
	}
}

func removeByIndexByte(slice []byte, index int) []byte {
	switch index {
	case 0:
		return slice[1:]
	case len(slice) - 1:
		return slice[:len(slice)-1]
	default:
		return append(slice[:index], slice[index+1:]...)
	}
}
