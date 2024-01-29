package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	slurm "github.com/intertwin-eu/interlink/pkg/sidecars/slurm"
)

func main() {
	logger := logrus.StandardLogger()

	interLinkConfig, err := commonIL.NewInterLinkConfig()
	if err != nil {
		panic(err)
	}

	if interLinkConfig.VerboseLogging {
		logger.SetLevel(logrus.DebugLevel)
	} else if interLinkConfig.ErrorsOnlyLogging {
		logger.SetLevel(logrus.ErrorLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	log.L = logruslogger.FromLogrus(logrus.NewEntry(logger))
	log.G(context.Background()).Debug("Debug level: " + strconv.FormatBool(interLinkConfig.VerboseLogging))

	SidecarAPIs := slurm.SidecarHandler{
		Config: interLinkConfig,
	}

	mutex := http.NewServeMux()
	mutex.HandleFunc("/status", SidecarAPIs.StatusHandler)
	mutex.HandleFunc("/create", SidecarAPIs.SubmitHandler)
	mutex.HandleFunc("/delete", SidecarAPIs.StopHandler)
	mutex.HandleFunc("/getLogs", SidecarAPIs.GetLogsHandler)

	slurm.CreateDirectories(interLinkConfig)
	slurm.LoadJIDs(interLinkConfig)

	err = http.ListenAndServe(":"+interLinkConfig.Sidecarport, mutex)
	if err != nil {
		log.G(slurm.Ctx).Fatal(err)
	}
}
