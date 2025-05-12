package main

import (
	"fmt"
	"os"
)

func initCMD() error {
	for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
	}

	headFileContents := []byte("ref: refs/heads/main\n")
	if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	fmt.Println("Initialized git directory")
	return nil
}

func catFileCMD(hash string) error {
	file, err := getFile(hash)

	if err != nil {
		return fmt.Errorf("cat file command: get file from hash: %w", err)
	}
	defer file.Close()

	content, objectType, err := readFile(file)

	if err != nil {
		return fmt.Errorf("error in reading object file: %s", err)
	}

	if objectType != "blob" {
		return fmt.Errorf("hash is not of type blob is %s", objectType)
	}
	fmt.Printf("%s", content)
	return nil

}
