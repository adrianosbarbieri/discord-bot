package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Audio struct {
	name string
	id   int
	path string
	buf  []byte
}

func LoadAllFiles(audios []Audio) {
	for i, _ := range audios {
		file, err := os.Open(audios[i].path)
		if err != nil {
			fmt.Println("Could not open audio file: ", err)
			continue
		}

		b, err := io.ReadAll(file)
		if err != nil {
			fmt.Println("Could not read audio file: ", err)
			continue
		}

		audios[i].buf = b
	}
}

func ReadAudioConfig(configPath string) ([]Audio, error) {
	config := make([]Audio, 0)

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

	for i, s := range lines {
		split := strings.Split(s, ";")
		if len(split) != 2 {
			continue
		}

		audio := Audio{
			name: split[0],
			id:   i + 1,
			path: split[1],
			buf:  nil,
		}

		config = append(config, audio)
	}

	return config, nil
}
