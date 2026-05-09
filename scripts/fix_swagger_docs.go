package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

func main() {
	const path = "docs/docs.go"

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	var out bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "LeftDelim:") || strings.HasPrefix(trimmed, "RightDelim:") {
			continue
		}
		out.WriteString(line)
		out.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "scan %s: %v\n", path, err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, out.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
		os.Exit(1)
	}
}
