package util

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"time"
)

const RandInFileName = "RandInFile"
const RandInListName = "RandInList"
const RandInRangeName = "RandInRange"

func RandInRange(min, max int) int {
	var generated int
	rand.Seed(time.Now().UnixNano())
	generated = rand.Intn(max-min) + min
	return generated
}

func RandInFile(filePath string) string {
	return RandInList(fileToList(filePath))
}

func RandInList(items []string) string {
	var generated string
	generated = items[RandInRange(0, len(items))]
	return generated
}

func fileToList(filePath string) []string {
	var fileItems []string
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileItems = append(fileItems, scanner.Text())
	}
	return fileItems
}
