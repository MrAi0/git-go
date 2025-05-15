package main

import "os"

type GitTree struct {
	Mode    os.FileMode
	GitMode string
	Name    string
	SHA     [20]byte
}

type GitTrees []GitTree
