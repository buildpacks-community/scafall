package main

import (
	"log"

	"github.com/buildpacks/scafall/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
