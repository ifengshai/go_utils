package main

import (
	"bufio"
	"fmt"
	"github.com/golang-module/carbon/v2"
	"os"
)

func main() {

	fmt.Println(carbon.Now().ToDateTimeString())
	os.Exit(1)
	file, err := os.Open("./../file/a.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		fmt.Println(line)
	}

	if scanner.Err() != nil {
		fmt.Println(scanner.Err())
	}
}
