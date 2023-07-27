package main

import "github.com/sethvargo/go-githubactions"

func main() {
	act := githubactions.New()
	act.Infof("Hello, world!")
}
