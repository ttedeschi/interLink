package interlink

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	commonIL "github.com/intertwin-eu/interlink/pkg/common"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("InterLink: received Delete call")
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req *http.Request
	reader := bytes.NewReader(bodyBytes)

	switch commonIL.InterLinkConfigInst.Sidecarservice {
	case "docker":
		req, err = http.NewRequest(http.MethodPost, commonIL.InterLinkConfigInst.Sidecarurl+":"+commonIL.InterLinkConfigInst.Sidecarport+"/delete", reader)

	case "slurm":
		req, err = http.NewRequest(http.MethodPost, commonIL.InterLinkConfigInst.Sidecarurl+":"+commonIL.InterLinkConfigInst.Sidecarport+"/stop", reader)

	case "htcondor":
		req, err = http.NewRequest(http.MethodPost, commonIL.InterLinkConfigInst.Sidecarurl+":"+commonIL.InterLinkConfigInst.Sidecarport+"/stop", reader)
	default:
		break
	}

	log.Println("InterLink: forwarding Delete call to sidecar")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	returnValue, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("InterLink: deleted container " + string(returnValue))

	w.Write(returnValue)
}
