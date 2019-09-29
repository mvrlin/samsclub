package main

import (
	"log"

	"github.com/mvrlin/samsclub/pkg/samsclub"
)

func main() {
	sc, err := samsclub.New()
	if err != nil {
		log.Fatalln(err)
	}

	if err = sc.Run(); err != nil {
		log.Fatalln(err)
	}
}
