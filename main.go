package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) < 3 {
		errorPrintf("usage: mygit <command> [<args>...]")
		os.Exit(1)
	}
	fmt.Println(os.Args)
	switch command := os.Args[2]; command {
	case "init":
		fmt.Printf("init command")
		must(initCMD())
	case "cat-file":
		fmt.Printf("cat-file command")
		if len(os.Args) != 5 {
			must(fmt.Errorf("usage: mygit cat-file <flag> <file>"))
		}

		if os.Args[3] != "-p" {
			must(fmt.Errorf("usage: mygit cat-file -p <file>"))
		}
		must(catFileCMD(os.Args[4]))
	case "hash-object":
		if len(os.Args) != 5 {
			must(fmt.Errorf("usage: mygit hash-object <flag> <file>"))
		}

		if os.Args[3] != "-w" {
			must(fmt.Errorf("usage: mygit hash-object -w <file>"))
		}
		fmt.Printf("hash-object command")
		must(hashObjectCMD(os.Args[4]))
	case "ls-tree":
		fmt.Println(os.Args)
		if len(os.Args) != 5 {
			must(fmt.Errorf("usage: mygit ls-tree <flag> <tree_sha>"))
		}

		if os.Args[3] != "--name-only" {
			must(fmt.Errorf("usage: mygit ls-tree --name-only <tree_sha>"))
		}
		fmt.Printf("ls-tree command")
		must(lsTreeCMD(os.Args[4]))
	case "write-tree":
		fmt.Printf("write-tree command")
	case "commit-tree":
		fmt.Printf("commit-tree command")
	case "clone":
		fmt.Printf("clone command")
	default:
		fmt.Printf("unknown command %s", command)
	}

	fmt.Println(os.Args)
}
