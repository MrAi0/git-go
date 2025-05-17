package main

import "os"

type GitTree struct {
	Mode    os.FileMode
	GitMode string
	Name    string
	SHA     [20]byte
}

type GitTrees []GitTree

const (
	defaultName    = "TestUser"
	defaultEmailID = "testuser@gmail.com"
)

type GitRefs struct {
	Hash string
	Name string
}
