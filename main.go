package main

import (
	"GoDeveloperPath/duplicate_file_handler"
	"GoDeveloperPath/loan_calulator"
	"GoDeveloperPath/smart_calculator"
	"GoDeveloperPath/vcs"
	"bufio"
	"fmt"
	"os"
	"strconv"
)

var projects = []string{
	"1. Duplicate File Handler",
	"2. Smart Calculator",
	"3. Loan Calculator",
	"4. Version Control System",
}

func main() {
	projectNum, err := chooseProject()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	switch projectNum {
	case 1:
		fmt.Println("You are running Duplicate File Handler")
		duplicate_file_handler.Run()
	case 2:
		fmt.Println("You are running Smart Calculator")
		smart_calculator.Run()
	case 3:
		fmt.Println("You are running Loan Calculator")
		loan_calulator.Run()
	case 4:
		fmt.Println("You are running Version Control System")
		vcs.Run()
	default:
		fmt.Println("Unknown project")
	}
}

func chooseProject() (int, error) {
	var num int
	var err error
	fmt.Println("Choose a project")
	for k, p := range projects {
		fmt.Printf("%d - %s\n", k, p)
	}
	fmt.Println("Enter the project's number or q to quit")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		switch text {
		case "q":
			os.Exit(0)
		default:
			num, err = strconv.Atoi(text)
			if err != nil {
				fmt.Println("Wrong format")
				continue
			}
			return num, nil
		}
	}
	return 0, err
}
