package main

import (
	"log"

	"github.com/AidanDelaney/scafall/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}
