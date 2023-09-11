package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	"github.com/intertwin-eu/interlink/pkg/interlink"
	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"
)

var Url string

func main() {
	var cancel context.CancelFunc

	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.DebugLevel)
	log.L = logruslogger.FromLogrus(logrus.NewEntry(logger))

	interlink.Ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	commonIL.NewInterLinkConfig()

	fmt.Println(commonIL.InterLinkConfigInst)

	go func() {
		time.Sleep(time.Millisecond * 50)
		for {
			var returnValue, _ = json.Marshal("Error")
			reader := bytes.NewReader(nil)
			req, err := http.NewRequest(http.MethodPost, commonIL.InterLinkConfigInst.VKurl+":"+commonIL.InterLinkConfigInst.VKport+"/sendCFG", reader)

			if err != nil {
				log.G(context.Background()).Error(err)
			}

			token, err := os.ReadFile(commonIL.InterLinkConfigInst.VKTokenFile) // just pass the file name
			if err != nil {
				log.G(context.Background()).Error(err)
			}
			req.Header.Add("Authorization", "Bearer "+string(token))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.G(context.Background()).Error(err)
				time.Sleep(5 * time.Second)
				continue
			} else {
				returnValue, _ = ioutil.ReadAll(resp.Body)
			}

			if resp.StatusCode == http.StatusOK {
				break
			} else {
				log.G(context.Background()).Error("Error " + err.Error() + " " + string(returnValue))
			}
		}
	}()

	mutex := http.NewServeMux()
	mutex.HandleFunc("/status", interlink.StatusHandler)
	mutex.HandleFunc("/create", interlink.CreateHandler)
	mutex.HandleFunc("/delete", interlink.DeleteHandler)
	mutex.HandleFunc("/setKubeCFG", interlink.SetKubeCFGHandler)
	err := http.ListenAndServe(":"+commonIL.InterLinkConfigInst.Interlinkport, mutex)
	if err != nil {
		log.G(interlink.Ctx).Fatal(err)
	}
}
