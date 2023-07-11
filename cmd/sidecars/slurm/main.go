package main

import (
	"log"
	"net/http"

	slurm "github.com/cloud-pg/interlink/pkg/sidecars/slurm"
	commonIL "github.com/intertwin-eu/interlink/pkg/common"
)

func main() {

	commonIL.NewInterLinkConfig()

	mutex := http.NewServeMux()
	mutex.HandleFunc("/status", slurm.StatusHandler)
	mutex.HandleFunc("/submit", slurm.SubmitHandler)
	mutex.HandleFunc("/stop", slurm.StopHandler)
	mutex.HandleFunc("/setKubeCFG", slurm.SetKubeCFGHandler)

	err := http.ListenAndServe(":"+commonIL.InterLinkConfigInst.Sidecarport, mutex)
	if err != nil {
		log.Fatal(err)
	}
}
