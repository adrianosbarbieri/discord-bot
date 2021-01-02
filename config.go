package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func loadAllFiles(audioMap map[string]string) map[string][]byte {
	fileMap := make(map[string][]byte)
	for s, path := range audioMap {
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("Could not open audio file: ", err)
			continue
		}

		b, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Println("Could not read audio file: ", err)
			continue
		}

		fileMap[s] = b
	}

	return fileMap
}

func readAudioConfig(configPath string) (map[string]string, error) {
	config := make(map[string]string)
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	for _, s := range lines {
		split := strings.Split(s, ";")
		if len(split) != 2 {
			continue
		}

		fmt.Println(split[0], split[1])

		config[split[0]] = split[1]
	}
	return config, nil
}
