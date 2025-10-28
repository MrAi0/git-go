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
		if len(os.Args) != 3 {
			must(fmt.Errorf("usage: mygit write-tree"))
		}

		must(wirteTreeCMD())
		fmt.Printf("write-tree command")
	case "commit-tree":
		if len(os.Args) != 8 {
			must(fmt.Errorf("usage: mygit commit-tree <tree-sha> -p <commit-sha> -m <msg>"))
		}

		if os.Args[4] != "-p" || os.Args[6] != "-m" {
			must(fmt.Errorf("usage: mygit commit-tree <tree-sha> -p <commit-sha> -m <msg>"))
		}

		must(commitTreeCMD(os.Args[3], os.Args[5], os.Args[7]))
		fmt.Printf("commit-tree command")
	case "clone":
		if len(os.Args) != 4 {
			must(fmt.Errorf("usage: mygit clone <url>"))
		}
		fmt.Printf("clone command")
		must(cloneCMD(os.Args[2], os.Args[3]))
	default:
		fmt.Printf("unknown command %s", command)
	}

	fmt.Println(os.Args)
}
