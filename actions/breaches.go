package actions

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func updateBFile(bfilePath string) {
	file, err := os.Open(bfilePath)
	if err != nil {
		log.Fatal(err)
	}
	//close the file when we're done
	defer file.Close()

	//read the file line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		fmt.Printf("line: %s\n", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
