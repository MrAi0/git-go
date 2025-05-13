package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
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
		pos++
	}

	parts := bytes.Split(decompressedData[0:pos], []byte{' '})

	return decompressedData[pos+1:], string(parts[0]), nil
}

func gitObject(contentType string, content []byte) []byte {
	contentLength := len(content)
	contentDigitLen := numOfDigits(contentLength)

	result := make([]byte, 0, len(contentType)+contentLength+1+contentDigitLen+len(content))

	result = append(result, contentType...)

	result = append(result, ' ')

	result = append(result, []byte(fmt.Sprintf("%d", contentLength))...)

	result = append(result, 0)

	result = append(result, content...)

	return result

}

func numOfDigits(d int) int {
	count := 0
	for d != 0 {
		d /= 10
		count++
	}
	return count
}

func calculateSHA(content []byte) (string, error) {
	hash, err := getRawSHA(content)

	if err != nil {
		return "", bytes.ErrTooLarge
	}

	sha := hex.EncodeToString(hash[:])
	return sha, nil
}

func getRawSHA(content []byte) ([]byte, error) {
	hash := sha1.New()

	_, err := hash.Write(content)

	if err != nil {
		return []byte{}, err
	}

	hashSum := hash.Sum(nil)

	if len(hashSum) != 20 {
		return []byte{}, fmt.Errorf("malformed hash created with %d bytes", len(hashSum))
	}

	return []byte(hashSum), nil
}

func createObjectFile(sha string) (*os.File, error) {

	if len(sha) != 40 {
		return nil, fmt.Errorf("invalid length of sha object %d", len(sha))
	}

	dir, rest := sha[0:2], sha[2:]
	err := os.Mkdir(fmt.Sprintf("./.git/objects/%s", dir), fs.FileMode(0755))
	if err != nil && !os.IsExist(err) {
		return nil, err
	}

	fmt.Println("Hash--->", sha)

	return os.Create(fmt.Sprintf("./.git/objects/%s/%s", dir, rest))
}

func writeZipContent(w io.Writer, content io.Reader) error {
	z := zlib.NewWriter(w)

	defer z.Close()

	contentByte, err := io.ReadAll(content)
	if err != nil {
		return fmt.Errorf("file could not write content")
	}

	contentWrite, err := z.Write(contentByte)
	if err != nil {
		return fmt.Errorf("file could not write content")
	}

	if contentWrite != len(contentByte) {
		return fmt.Errorf(" content length and written bytes do not match")
	}

	return nil
}
