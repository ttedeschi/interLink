package interlink

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("InterLink: received GetStatus call")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	reader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequest(http.MethodGet, commonIL.InterLinkConfigInst.Sidecarurl+":"+commonIL.InterLinkConfigInst.Sidecarport+"/status", reader)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("InterLink: forwarding GetStatus call to sidecar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	returnValue, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("InterLink: status " + string(returnValue))

	w.Write(returnValue)
}
