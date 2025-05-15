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
	"sort"
	"strconv"
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

func getRawSHA(content []byte) ([20]byte, error) {
	hash := sha1.New()

	n, err := hash.Write(content)

	if err != nil {
		return [20]byte{}, err
	}

	if n != len(content) {
		return [20]byte{}, fmt.Errorf("mismatch in the bytes and content %d and %d", n, len(content))
	}

	hashSum := hash.Sum(nil)

	if len(hashSum) != 20 {
		return [20]byte{}, fmt.Errorf("malformed hash created with %d bytes", len(hashSum))
	}

	var resArray [20]byte
	copy(resArray[:], hashSum)

	return resArray, nil
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

func parseTreeObject(content []byte) ([]GitTree, error) {

	res := []GitTree{}

	nameStart := 0
	spaceStart := 0

	for i := 0; i < len(content); i++ {
		curr := GitTree{}
		if content[i] == ' ' {
			fileMode := content[spaceStart:i]
			mode, err := strconv.Atoi(string(fileMode))
			if err != nil {
				return nil, err
			}

			curr.Mode = fs.FileMode(mode)
			nameStart = i + 1
		}

		if content[i] == 0 {
			name := content[nameStart:i]

			if i+1+20 > len(content) {
				return nil, fmt.Errorf("unexpected end of content while reading SHA")
			}

			var sha [20]byte
			copy(sha[:], content[i+1:i+1+20])

			curr.Name = string(name)
			curr.SHA = sha

			spaceStart = i + 21
			i += 20
			res = append(res, curr)
		}
	}
	return res, nil
}

func writeTree(dirPath string) ([20]byte, error) {
	var buffer bytes.Buffer
	entries := []GitTree{}

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {

		if err != nil {
			return fmt.Errorf("error accessing %s: %w", path, err)
		}

		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		if d.IsDir() {
			if path == dirPath {
				return nil
			}

			subTreeSHA, err := writeTree(path)
			if err != nil {
				return err
			}
			entries = append(entries, GitTree{
				Mode:    d.Type(),
				GitMode: "40000",
				Name:    d.Name(),
				SHA:     subTreeSHA,
			})

			return filepath.SkipDir
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file %s: %w", path, err)
		}
		defer file.Close()

		fileContent, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		fullContent := gitObject("blob", fileContent)

		rawSHA, err := getRawSHA(fullContent)
		if err != nil {
			return fmt.Errorf("calculate file SHA for %s: %w", path, err)
		}

		mode := "100644" // Default mode for regular files
		if d.Type().Perm()&0111 != 0 {
			mode = "100755"
		}

		entries = append(entries, GitTree{
			Mode:    d.Type(),
			GitMode: mode,
			Name:    d.Name(),
			SHA:     rawSHA,
		})
		return nil

	})

	if err != nil {
		return [20]byte{}, err
	}

	_, err = GitTrees(entries).writeTo(&buffer)
	if err != nil {
		return [20]byte{}, err
	}

	return bufferToFile(&buffer)
}

func (t GitTrees) writeTo(w io.Writer) (int64, error) {
	sort.Slice(t, func(i, j int) bool {
		return t[i].Name < t[j].Name
	})

	var n int64

	for _, entry := range t {
		n1, err := fmt.Fprintf(w, "%s %s", entry.GitMode, entry.Name)
		if err != nil {
			return n, err
		}

		n += int64(n1)
		n2, err := w.Write([]byte{0})
		if err != nil {
			return n, err
		}

		n += int64(n2)
		n3, err := w.Write(entry.SHA[:])
		if err != nil {
			return n, err
		}

		n += int64(n3)

	}

	return n, nil
}

func bufferToFile(buffer *bytes.Buffer) ([20]byte, error) {
	treeContent := buffer.Bytes()

	treeRawSHA, err := getRawSHA((gitObject("tree", treeContent)))
	if err != nil {
		return [20]byte{}, err
	}

	treeSHA := hex.EncodeToString(treeRawSHA[:])
	treeFile, err := createObjectFile(treeSHA)
	if err != nil {
		if os.IsExist(err) {
			return treeRawSHA, nil
		}
		return [20]byte{}, fmt.Errorf("couldn't create tree object file: %w", err)
	}

	defer treeFile.Close()
	err = writeZipContent(treeFile, bytes.NewReader(gitObject("tree", treeContent)))

	if err != nil {
		return [20]byte{}, err
	}

	return treeRawSHA, nil
}
