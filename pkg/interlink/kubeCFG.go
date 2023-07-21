package interlink

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
)

func SetKubeCFGHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("InterLink: received SetKubeCFG call")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req *http.Request
	reader := bytes.NewReader(bodyBytes)
	var returnValue []byte

	for {
		req, err = http.NewRequest(http.MethodPost, commonIL.InterLinkConfigInst.Sidecarurl+":"+commonIL.InterLinkConfigInst.Sidecarport+"/setKubeCFG", reader)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("InterLink: forwarding SetKubeCFG call to sidecar")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
		}

		returnValue, _ = ioutil.ReadAll(resp.Body)
		fmt.Println(string(returnValue))

		if string(returnValue) == "200" {
			log.Println("InterLink: received a valid response")
			break
		}
		log.Println("InterLink: received a not valid response: " + string(returnValue))
	}

	w.Write(returnValue)
}
