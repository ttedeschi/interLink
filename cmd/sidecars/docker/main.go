package main

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	docker "github.com/intertwin-eu/interlink/pkg/sidecars/docker"
)

func main() {
	var cancel context.CancelFunc

	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	log.L = logruslogger.FromLogrus(logrus.NewEntry(logger))

	docker.Ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	commonIL.NewInterLinkConfig()

	mutex := http.NewServeMux()
	mutex.HandleFunc("/status", docker.StatusHandler)
	mutex.HandleFunc("/create", docker.CreateHandler)
	mutex.HandleFunc("/delete", docker.DeleteHandler)
	err := http.ListenAndServe(":"+commonIL.InterLinkConfigInst.Sidecarport, mutex)

	if err != nil {
		log.L.Fatal(err)
	}
}
