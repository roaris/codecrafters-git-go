package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/ioutil"

	// Uncomment this block to pass the first stage!
	"os"
)

// Usage: your_git.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage!
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/master\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")
	case "cat-file":
		if len(os.Args) >= 3 && os.Args[2] == "-p" {
			hash := os.Args[3]
			directoryName := hash[:2]
			fileName := hash[2:]

			f, err := os.Open(fmt.Sprintf(".git/objects/%s/%s", directoryName, fileName))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file: %s\n", err)
			}
			defer f.Close()
			b, err := ioutil.ReadAll(f)

			var buf bytes.Buffer
			buf.Write(b)
			r, _ := zlib.NewReader(&buf)
			b = make([]byte, 1024)
			r.Read(b)

			for i := 0; i < 1024; i++ {
				if b[i] == 0 { // ファイルサイズの後のヌルバイト
					for j := i + 1; j < 1024; j++ {
						if b[j] == 0 {
							fmt.Print(string(b[i+1 : j]))
							break
						}
					}
					break
				}
			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
