package vcs

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var commands = []string{"config", "add", "log", "commit", "checkout"}
var commandsCache = map[string]string{
	commands[0]: "Get and set a username.",
	commands[1]: "Add a file to the index.",
	commands[2]: "Show commit logs.",
	commands[3]: "Save changes.",
	commands[4]: "Restore a file.",
}

type LogMessage struct {
	Hash    string
	Author  string
	Message string
}

func Run() {

	fmt.Println("This is version control system. Before you start, do you want to know how to use it?")
	var needHelp bool
	_, err := fmt.Scanln(&needHelp)
	if !needHelp && err != nil {
		help()
		os.Exit(0)
	}

	fmt.Println("Enter command and arguments:")
	fmt.Println("Example: config username, add file, log, commit -m \"message\", checkout commit_id")
	var input []string
	var arg string
	for {
		n, err := fmt.Scanln(&arg)
		if n == 0 || err != nil {
			break
		}
		input = append(input, arg)
	}

	err = os.MkdirAll("./vcs", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll("./vcs/commits", os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	command := input[1]
	procArguments := input[2:]

	switch command {
	case "config":
		ConfigCommand(procArguments)
	case "add":
		AddCommand(procArguments)
	case "log":
		LogCommand()
	case "commit":
		CommitCommand(procArguments)
	case "checkout":
		CheckoutCommand(procArguments)
	default:
		wrongCommand(command)
	}
}

func CheckoutCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Commit id was not passed.")
		return
	}
	hash := args[0]
	hashes := getHashes()

	for _, item := range hashes {
		if item == hash {
			fileNames := *getTrackedFiles()
			for _, fileName := range fileNames {
				source := fmt.Sprintf("./vcs/commits/%s/%s", hash, fileName)
				if err := copyFile(source, fileName); err != nil {
					log.Fatal(err)
				}
			}
			fmt.Printf("Switched to commit %s.\n", hash)
			return
		}
	}
	fmt.Println("Commit does not exist.")
}

func CommitCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Message was not passed.")
		return
	}

	message := args[0]
	user := getUser()
	fileNames := *getTrackedFiles()
	messages := *getLogMessages()

	sha256Hash, err := getHash(fileNames)
	if err != nil {
		log.Fatal(err)
	}

	if len(messages) == 0 || messages[0].Hash != sha256Hash {

		destinationDir := fmt.Sprintf("./vcs/commits/%s", sha256Hash)
		if err := os.MkdirAll(destinationDir, os.ModePerm); err != nil {
			log.Fatal(err)
		}

		for _, fileName := range fileNames {
			destPath := fmt.Sprintf("%s/%s", destinationDir, fileName)
			if err := copyFile(fileName, destPath); err != nil {
				cleanup(destinationDir)
				log.Fatal(err)
			}
		}

		if err = logVscMessage(append([]LogMessage{{Hash: sha256Hash, Author: user, Message: message}}, messages...)); err != nil {
			cleanup(destinationDir)
			log.Fatal(err)
		}
		fmt.Println("Changes are committed.")
		return
	}

	if messages[0].Hash == sha256Hash {
		fmt.Print("Nothing to commit.")
	}
}

func LogCommand() {
	logMessages := *getLogMessages()

	if len(logMessages) == 0 {
		fmt.Println("No commits yet.")
		return
	}

	for _, msg := range logMessages {
		fmt.Printf("commit %s\nAuthor: %s\n%s\n", msg.Hash, msg.Author, msg.Message)
	}
}

func AddCommand(args []string) {
	fileNames := *getTrackedFiles()
	if len(args) == 0 && len(fileNames) == 0 {
		fmt.Print(commandsCache[commands[1]])
		return
	}
	if len(args) == 1 {
		file, err := os.OpenFile("./vcs/index.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
		if err != nil {
			log.Fatal(err)
		}
		defer closeFile(file)
		filename := args[0]
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			fmt.Printf("Can't find '%s'.\n", filename)
			return
		}
		if _, err = fmt.Fprintln(file, filename); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The file '%s' is tracked.", filename)
		return
	}
	fmt.Println("Tracked files:")
	for _, v := range fileNames {
		fmt.Println(v)
	}
}

func ConfigCommand(args []string) {
	file, err := os.OpenFile("./vcs/config.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	currentUser := scanner.Text()
	if len(args) == 1 {
		newUser := args[0]
		if err := os.WriteFile(file.Name(), []byte(newUser), 0644); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The username is %s.\n", newUser)
		return
	}
	if currentUser != "" {
		fmt.Printf("The username is %s.\n", currentUser)
		return
	}
	fmt.Println("Please, tell me who you are.")
}

func wrongCommand(command string) {
	fmt.Printf("'%s' is not a SVCS command.", command)
}

func help() {
	fmt.Println("These are SVCS commands:")
	for _, key := range commands {
		fmt.Printf("%-10s %s\n", key, commandsCache[key])
	}
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func getTrackedFiles() *[]string {
	file, err := os.OpenFile("./vcs/index.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	fileNames := make([]string, 0, 10)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileNames = append(fileNames, scanner.Text())
	}
	return &fileNames
}

func getHashes() []string {
	logMessages := *getLogMessages()
	hashes := make([]string, 0, 10)
	for _, msg := range logMessages {
		hashes = append(hashes, msg.Hash)
	}
	return hashes
}

func getLogMessages() *[]LogMessage {
	file, err := os.OpenFile("./vcs/log.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	logMessages := make([]LogMessage, 0, 10)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ";")
		if len(fields) == 3 {
			logMessages = append(logMessages, LogMessage{
				Hash:    fields[0],
				Author:  fields[1],
				Message: fields[2],
			})
		}
	}
	return &logMessages
}

func getUser() string {
	file, err := os.OpenFile("./vcs/config.txt", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile(file)
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	return scanner.Text()
}

func getHash(fileNames []string) (string, error) {
	sha256Hashier := sha256.New()
	for _, filename := range fileNames {
		err := hashFile(sha256Hashier, filename)
		if err != nil {
			log.Fatal(err)
			return "", err
		}
	}
	return hex.EncodeToString(sha256Hashier.Sum(nil)), nil
}

func hashFile(hasher io.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer closeFile(file)

	_, err = io.Copy(hasher, file)
	return err
}

func copyFile(sourcePath string, destinationPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer closeFile(sourceFile)
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer closeFile(destinationFile)
	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func cleanup(dirPath string) {
	err := os.RemoveAll(dirPath)
	if err != nil {
		log.Println("Failed to clean up:", err)
	} else {
		log.Println("Cleaned up:", dirPath)
	}
}

func logVscMessage(messages []LogMessage) error {
	vscLog, err := os.OpenFile("./vcs/log.txt", os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer closeFile(vscLog)

	for _, msg := range messages {
		_, err = fmt.Fprintf(vscLog, "%s;%s;%s\n", msg.Hash, msg.Author, msg.Message)
		if err != nil {
			return err
		}
	}

	return nil
}
