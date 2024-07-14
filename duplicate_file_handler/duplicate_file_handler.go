package duplicate_file_handler

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type FilesBySize map[int64][]string
type FileToDelete struct {
	number int
	path   string
	size   int64
}
type FilesByHash map[string][]FileToDelete
type CollectingRequest struct {
	folder  string
	format  string
	sorting int
}

func Run() {

	var err error
	var filesBySize = make(FilesBySize)

	request := createCollectingRequest()

	err = groupFiles(request, &filesBySize)
	if err != nil {
		exitProgram(1, err.Error())
	}
	sortedKeys := sortKeys(&filesBySize, request.sorting)
	printFilesBySize(&filesBySize, sortedKeys)

	var duplicates *[]FileToDelete

	if yesOrNoQuestion("Check for duplicates?") {
		duplicates = processDuplicates(&filesBySize, sortedKeys)
	} else {
		exitProgram(0)
	}

	if yesOrNoQuestion("Delete files?") {
		deleteFiles(duplicates, sortedKeys)
	} else {
		exitProgram(0)
	}
}

func createCollectingRequest() CollectingRequest {
	var root string
	var format string
	var sorting int

	fmt.Println("Enter the directory to check duplicates:")
	_, err := fmt.Scanln(&root)
	if root == "" || err != nil {
		exitProgram(1, "Directory is not specified")
	}

	_, err = os.Stat(root)
	if os.IsNotExist(err) {
		exitProgram(1, "Directory does not exist")
	}
	fmt.Println("Enter file format:")
	_, err = fmt.Scanln(&format)
	if err != nil && err.Error() != "unexpected newline" {
		exitProgram(1, err.Error())
	}
	fmt.Println("Size sorting options:")
	fmt.Println("1. Descending")
	fmt.Println("2. Ascending")
	for {
		fmt.Println("Enter a sorting option:")
		_, err = fmt.Scanln(&sorting)
		if err != nil {
			exitProgram(1, err.Error())
		}
		if sorting == 1 || sorting == 2 {
			break
		}
		fmt.Println("Wrong option")
	}
	return CollectingRequest{root, format, sorting}
}

func groupFiles(request CollectingRequest, fileMap *FilesBySize) error {
	result := *fileMap
	err := filepath.Walk(request.folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			exitProgram(1, err.Error())
		}
		if !info.IsDir() {
			if request.format != "" {
				if filepath.Ext(path) != "."+request.format {
					return nil
				}
			}
			size := info.Size()
			result[size] = append(result[size], path)
		}
		return nil
	})
	return err
}

func sortKeys(fm *FilesBySize, sorting int) []int64 {
	keys := make([]int64, 0, len(*fm))
	for k := range *fm {
		keys = append(keys, k)
	}
	if sorting == 1 {
		sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
	} else {
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	}
	return keys
}

func printFilesBySize(fm *FilesBySize, sizes []int64) {
	for _, size := range sizes {
		files := (*fm)[size]
		fmt.Printf("%d bytes\n", size)
		for _, file := range files {
			fmt.Println(file)
		}
	}
}

func yesOrNoQuestion(question string) bool {
	for {
		var checkForDuplicates string
		fmt.Println(question)
		_, err := fmt.Scanln(&checkForDuplicates)
		if strings.ToLower(checkForDuplicates) == "yes" {
			return true
		}
		if err != nil || strings.ToLower(checkForDuplicates) == "no" {
			return false
		}
		fmt.Println("Wrong option")
	}
}

func processDuplicates(fm *FilesBySize, sortedKeys []int64) *[]FileToDelete {
	counter := 1
	var result []FileToDelete
	for _, size := range sortedKeys {
		filesBySize := (*fm)[size]
		filesByHash := getDuplicates(filesBySize)
		var duplicatesForPrinting = make(FilesByHash)
		for hash, files := range filesByHash {
			var filesToDelete []FileToDelete
			for _, filePath := range files {
				toDelete := FileToDelete{counter, filePath, size}
				counter++
				filesToDelete = append(filesToDelete, toDelete)
			}
			duplicatesForPrinting[hash] = append(duplicatesForPrinting[hash], filesToDelete...)
			result = append(result, filesToDelete...)
		}
		printDuplicates(&duplicatesForPrinting, size)
	}
	return &result
}

func getDuplicates(filePaths []string) map[string][]string {
	var filesByHash = make(map[string][]string)
	for _, fileName := range filePaths {
		hash := getHash(fileName)
		filesByHash[hash] = append(filesByHash[hash], fileName)
	}
	for hash, files := range filesByHash {
		if len(files) == 1 {
			delete(filesByHash, hash)
		}
	}
	return filesByHash
}

func printDuplicates(d *FilesByHash, size int64) {
	fmt.Printf("%d bytes\n", size)
	for hash, files := range *d {
		fmt.Printf("Hash: %s\n", hash)
		for _, file := range files {
			fmt.Printf("%d. %s\n", file.number, file.path)
		}
	}
}

func getHash(file string) string {
	f, err := os.Open(file)
	if err != nil {
		exitProgram(1, err.Error())
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		exitProgram(1, err.Error())
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func deleteFiles(duplicates *[]FileToDelete, sizes []int64) {
	nums := readIndexesToDelete()
	freeSpace := 0
	for _, num := range nums {
		file := (*duplicates)[num-1]
		err := os.Remove(file.path)
		if err != nil {
			exitProgram(1, err.Error())
		}
		freeSpace += int(file.size)
	}
	fmt.Printf("Total freed up space: %d bytes\n", freeSpace)
}

func readIndexesToDelete() []int {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("Enter file numbers to delete:")
		if scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				fmt.Println("Wrong format")
				continue
			}
			nums := strings.Fields(line)
			var indexes []int
			for _, numStr := range nums {
				num, err := strconv.Atoi(numStr)
				if err != nil {
					fmt.Println("Wrong format")
					continue
				}
				indexes = append(indexes, num)
			}
			if len(indexes) > 0 {
				return indexes
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading input:", err)
		}
	}
}

func exitProgram(code int, messages ...string) {
	for _, message := range messages {
		fmt.Println(message)
	}
	os.Exit(code)
}
