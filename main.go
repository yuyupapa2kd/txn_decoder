package main

import (
	"fmt"
	"log"

	"github.com/the-medium/tx-decoder/abis"
	"github.com/the-medium/tx-decoder/api"
)

func main() {

	//db.InitDB()

	//container := abis.InitABIContainer()
	//err := container.StoreDefaultABIToABIContABIContainer()
	err := abis.StoreDefaultABIToABIContABIContainer()
	if err != nil {
		log.Fatal(err)
	}
	/*
		err = core.ConnectToABIContainer(container)
		if err != nil {
			log.Fatal(err)
		}
	*/
	//container.WatchDir()
	go abis.WatchDir()
	fmt.Println("watch mode start")

	r := api.SetRouter()
	go r.Run(":8080")
	fmt.Println("server start")
	/*
		//container.WatchDir()
		go abis.WatchDir()
		fmt.Println("watch mode start")
	*/
	select {}

}
