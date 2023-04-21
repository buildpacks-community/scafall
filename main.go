package main

import (
	"log"

	"github.com/buildpacks-community/scafall/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
