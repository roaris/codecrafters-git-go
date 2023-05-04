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
	"sort"
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
	case "ls-tree":
		hash := os.Args[3]
		directoryName := hash[:2]
		fileName := hash[2:]

		f, _ := os.Open(filepath.Join(".git", "objects", directoryName, fileName))
		defer f.Close()

		r, _ := zlib.NewReader(f)
		store, _ := ioutil.ReadAll(r)
		l := strings.Split(string(store), "\u0000")

		for i := 1; i < len(l)-1; i++ {
			fmt.Println(strings.Split(l[i], " ")[1])
		}
	case "write-tree":
		currentDir, _ := os.Getwd()
		h, c := calcTreeHash(currentDir)
		treeHash := hex.EncodeToString(h)

		os.Mkdir(filepath.Join(".git", "objects", treeHash[:2]), 0755)
		var compressed bytes.Buffer
		w := zlib.NewWriter(&compressed)
		w.Write(c)
		w.Close()
		os.WriteFile(filepath.Join(".git", "objects", treeHash[:2], treeHash[2:]), compressed.Bytes(), 0644)

		fmt.Println(treeHash)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func calcTreeHash(dir string) ([]byte, []byte) {
	fileInfos, _ := ioutil.ReadDir(dir)
	type entry struct {
		fileName string
		b        []byte
	}
	var entries []entry
	contentSize := 0

	for _, fileInfo := range fileInfos {
		if fileInfo.Name() == ".git" {
			continue
		}
		if !fileInfo.IsDir() {
			f, _ := os.Open(filepath.Join(dir, fileInfo.Name()))
			b, _ := ioutil.ReadAll(f)
			s := fmt.Sprintf("blob %d\u0000%s", len(b), string(b))

			sha1 := sha1.New()
			io.WriteString(sha1, s)
			s = fmt.Sprintf("100644 %s\u0000", fileInfo.Name())
			b = append([]byte(s), sha1.Sum(nil)...)
			entries = append(entries, entry{fileInfo.Name(), b})
			contentSize += len(b)
		} else {
			b, _ := calcTreeHash(filepath.Join(dir, fileInfo.Name()))
			s := fmt.Sprintf("40000 %s\u0000", fileInfo.Name())
			b2 := append([]byte(s), b...)
			entries = append(entries, entry{fileInfo.Name(), b2})
			contentSize += len(b2)
		}
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].fileName < entries[j].fileName })
	s := fmt.Sprintf("tree %d\u0000", contentSize)
	b := []byte(s)

	for _, entry := range entries {
		b = append(b, entry.b...)
	}

	sha1 := sha1.New()
	io.WriteString(sha1, string(b))
	return sha1.Sum(nil), b
}
