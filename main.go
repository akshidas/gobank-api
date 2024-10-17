package main

import "log"

func main() {
	store, err := NewPostgresStore()
	if err != err {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	apiServer := &ApiServer{port: ":3000", store: store}
	apiServer.Run()
}
