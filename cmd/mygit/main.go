package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

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

			f, _ := os.Open(filepath.Join(".git/objects", directoryName, fileName))
			defer f.Close()

			r, _ := zlib.NewReader(f)
			store, _ := ioutil.ReadAll(r)
			fmt.Print(strings.Split(string(store), "\u0000")[1])
		}
	case "hash-object":
		if len(os.Args) >= 3 && os.Args[2] == "-w" {
			fileName := os.Args[3]
			f, _ := os.Open(fileName)
			b, _ := ioutil.ReadAll(f)
			s := fmt.Sprintf("blob %d\u0000%s", len(b), string(b))

			sha1 := sha1.New()
			io.WriteString(sha1, s)
			blobHash := hex.EncodeToString(sha1.Sum(nil))

			os.Mkdir(fmt.Sprintf(".git/objects/%s", blobHash[:2]), 0755)
			var compressed bytes.Buffer
			w := zlib.NewWriter(&compressed)
			w.Write([]byte(s))
			w.Close()
			os.WriteFile(fmt.Sprintf(".git/objects/%s/%s", blobHash[:2], blobHash[2:]), compressed.Bytes(), 0644)

			fmt.Println(blobHash)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
