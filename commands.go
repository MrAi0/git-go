package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
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

func hashObjectCMD(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return fmt.Errorf("error in reading file %w", err)
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error in reading file %w", err)
	}

	contentToWrite := gitObject("blob", content)
	fileSHA, err := calculateSHA(contentToWrite)
	if err != nil {
		return fmt.Errorf("error in calculating SHA")
	}
	newFile, err := createObjectFile(fileSHA)
	if err != nil {
		return fmt.Errorf("error in creating the object file")
	}

	err = writeZipContent(newFile, bytes.NewReader(contentToWrite))
	if err != nil {
		return fmt.Errorf("error writing to the object folder")
	}

	fmt.Println("File SHA: ", fileSHA)
	return nil
}

func lsTreeCMD(hash string) error {
	file, err := getFile(hash)
	if err != nil {
		return fmt.Errorf("ks tree command: get file from hash: %w", err)
	}

	defer file.Close()

	text, objectType, err := readFile(file)
	if err != nil {
		return fmt.Errorf("error in reading object file: %w", err)
	}

	if objectType != "tree" {
		return fmt.Errorf("object not of type tree")
	}

	tree, err := parseTreeObject(text)
	if err != nil {
		return fmt.Errorf("error in reading tree object")
	}

	for i := range tree {
		fmt.Println(tree[i].Name)
	}

	return nil
}

func wirteTreeCMD() error {
	treeSHA, err := writeTree(".")

	if err != nil {
		return fmt.Errorf("error in writing tree %w", err)
	}

	fmt.Println(hex.EncodeToString(treeSHA[:]))
	return nil
}

func commitTreeCMD(treeSHA, commitSHA, commitMsg string) error {
	if len(treeSHA) != 40 {
		return fmt.Errorf("invalid treeSHA")
	}

	if len(commitSHA) != 40 {
		return fmt.Errorf("invalid commitSHA")
	}

	content, err := writeCommitContent(treeSHA, commitMsg, commitSHA)
	if err != nil {
		return fmt.Errorf("write commit file: %w", err)
	}

	fullContent := gitObject("commit", content)

	fullContentSHA, err := calculateSHA(fullContent)
	if err != nil {
		return fmt.Errorf("calculate SHA: %w", err)
	}

	file, err := createObjectFile(fullContentSHA)
	if err != nil {
		return fmt.Errorf("create object file: %w", err)
	}
	err = writeZipContent(file, bytes.NewReader(fullContent))
	if err != nil {
		return fmt.Errorf("write object file %s", err)
	}

	fmt.Printf("%s", fullContentSHA)
	return nil
}

func cloneCMD(url string) error {
	packContent, err := fetchPackFile(url)
	if err != nil {
		return fmt.Errorf("could not fetch pack file: %w", err)
	}

	// packStrContent := string(packContent)
	refsArr, err := extractRefs(packContent)
	if err != nil {
		return fmt.Errorf("error extracting refs %w", err)
	}
	// fmt.Println(packStrContent)
	fmt.Println(refsArr)

	return nil
}
