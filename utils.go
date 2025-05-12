package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func errorPrintf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a...)
}

func must(err error) {
	if err != nil {
		errorPrintf("%s\n", err)
		os.Exit(1)
	}
}

func getFile(hash string) (*os.File, error) {
	dir := hash[0:2]
	rem := hash[2:]
	path := filepath.Join(".git/objects", dir, rem)
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no object found %w", hash)
	}
	return file, nil
}

func readFile(r io.Reader) ([]byte, string, error) {
	zlibReader, err := zlib.NewReader(r)
	if err != nil {
		return nil, "", err
	}
	defer zlibReader.Close()

	decompressedData, err := io.ReadAll(zlibReader)
	if err != nil {
		return nil, "", err
	}

	pos := 0
	for _, str := range decompressedData {
		if str == 0 {
			break
		}
	}

	parts := bytes.Split(decompressedData[0:pos], []byte{' '})

	return decompressedData[pos+1:], string(parts[0]), nil
}
