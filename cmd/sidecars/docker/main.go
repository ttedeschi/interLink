package main

import (
	"log"
	"net/http"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	docker "github.com/intertwin-eu/interlink/pkg/sidecars/docker"
)

func main() {

	commonIL.NewInterLinkConfig()

	mutex := http.NewServeMux()
	mutex.HandleFunc("/status", docker.StatusHandler)
	mutex.HandleFunc("/create", docker.CreateHandler)
	mutex.HandleFunc("/delete", docker.DeleteHandler)
	mutex.HandleFunc("/genericCall", docker.GenericCallHandler)

	err := http.ListenAndServe(":"+commonIL.InterLinkConfigInst.Sidecarport, mutex)
	if err != nil {
		log.Fatal(err)
	}
}
