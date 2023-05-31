package utils

import (
	"bufio"
	"io"
)

func ReadLines(r *bufio.Reader) ([]string, error) {
	var lines []string

	for {
		line, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		lines = append(lines, string(line))
	}

	return lines, nil
}
