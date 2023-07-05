package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

var InterLinkConfigInst InterLinkConfig

func NewInterLinkConfig() {
	if InterLinkConfigInst.set == false {
		var path string
		if os.Getenv("INTERLINKCONFIGPATH") != "" {
			path = os.Getenv("INTERLINKCONFIGPATH")
		} else {
			path = "/etc/interlink/InterLinkConfig.yaml"
		}

		if _, err := os.Stat(path); err != nil {
			log.Println("File " + path + " doesn't exist. You can set a custom path by exporting INTERLINKCONFIGPATH. Exiting...")
			os.Exit(-1)
		}

		yfile, err := os.ReadFile(path)
		if err != nil {
			log.Println("Error opening config file, exiting...")
			os.Exit(1)
		}
		yaml.Unmarshal(yfile, &InterLinkConfigInst)

		if os.Getenv("INTERLINKURL") != "" {
			InterLinkConfigInst.Interlinkurl = os.Getenv("INTERLINKURL")
		}

		if os.Getenv("SIDECARURL") != "" {
			InterLinkConfigInst.Sidecarurl = os.Getenv("SIDECARURL")
		}

		if os.Getenv("INTERLINKPORT") != "" {
			InterLinkConfigInst.Interlinkport = os.Getenv("INTERLINKPORT")
		}

		if os.Getenv("SIDECARSERVICE") != "" {
			if os.Getenv("SIDECARSERVICE") != "docker" && os.Getenv("SIDECARSERVICE") != "slurm" {
				fmt.Println("export SIDECARSERVICE as docker or slurm")
				os.Exit(-1)
			}
			InterLinkConfigInst.Sidecarservice = os.Getenv("SIDECARSERVICE")
		} else if InterLinkConfigInst.Sidecarservice != "docker" && InterLinkConfigInst.Sidecarservice != "slurm" {
			fmt.Println("Set \"docker\" or \"slurm\" in config file or export SIDECARSERVICE as ENV")
			os.Exit(-1)
		}

		if os.Getenv("SIDECARPORT") != "" && os.Getenv("SIDECARSERVICE") == "" {
			InterLinkConfigInst.Sidecarport = os.Getenv("SIDECARPORT")
			InterLinkConfigInst.Sidecarservice = "Custom Service"
		} else {
			switch InterLinkConfigInst.Sidecarservice {
			case "docker":
				InterLinkConfigInst.Sidecarport = "4000"

			case "slurm":
				InterLinkConfigInst.Sidecarport = "4001"

			default:
				fmt.Println("Define in InterLinkConfig.yaml one service between docker and slurm")
				os.Exit(-1)
			}
		}

		if os.Getenv("SBATCHPATH") != "" {
			InterLinkConfigInst.Sbatchpath = os.Getenv("SBATCHPATH")
		}

		if os.Getenv("SCANCELPATH") != "" {
			InterLinkConfigInst.Scancelpath = os.Getenv("SCANCELPATH")
		}

		if os.Getenv("TSOCKS") != "" {
			if os.Getenv("TSOCKS") != "true" && os.Getenv("TSOCKS") != "false" {
				fmt.Println("export TSOCKS as true or false")
				os.Exit(-1)
			}
			if os.Getenv("TSOCKS") == "true" {
				InterLinkConfigInst.Tsocks = true
			} else {
				InterLinkConfigInst.Tsocks = false
			}
		}

		if os.Getenv("TSOCKSPATH") != "" {
			path := os.Getenv("TSOCKSPATH")
			if _, err := os.Stat(path); err != nil {
				log.Println("File " + path + " doesn't exist. You can set a custom path by exporting TSOCKSPATH. Exiting...")
				os.Exit(-1)
			}

			InterLinkConfigInst.Tsockspath = path
		}

		if os.Getenv("VKTOKENFILE") != "" {
			path := os.Getenv("VKTOKENFILE")
			if _, err := os.Stat(path); err != nil {
				log.Println("File " + path + " doesn't exist. You can set a custom path by exporting VKTOKENFILE. Exiting...")
				os.Exit(-1)
			}

			InterLinkConfigInst.VKTokenFile = path
		} else {
			path = "/tmp/token"
			InterLinkConfigInst.VKTokenFile = path
		}

		InterLinkConfigInst.set = true
	}
}

func NewServiceAgent() {

	var path string

	if os.Getenv("CUSTOMKUBECONF") != "" {
		path = os.Getenv("CUSTOMKUBECONF")
	} else {
		path = "/tmp/sa.kubeconfig"
		sa := "apiVersion: v1\n" +
			"clusters:\n" +
			"- cluster:\n" +
			"	certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUJlRENDQVIyZ0F3SUJBZ0lCQURBS0JnZ3Foa2pPUFFRREFqQWpNU0V3SHdZRFZRUUREQmhyTTNNdGMyVnkKZG1WeUxXTmhRREUyTmpreU56azJPVE13SGhjTk1qSXhNVEkwTURnME9ERXpXaGNOTXpJeE1USXhNRGcwT0RFegpXakFqTVNFd0h3WURWUVFEREJock0zTXRjMlZ5ZG1WeUxXTmhRREUyTmpreU56azJPVE13V1RBVEJnY3Foa2pPClBRSUJCZ2dxaGtqT1BRTUJCd05DQUFUTjZMZzJSaFBtV09pUTdEUkNzenlJeDFnaERLR3l1K3hEaEhyR21BSU4KUkh3R0RqdEVJWEtuYXQrdmhIOE9wSVJPS0ZLK2xKNThDc3J0TW4vSGxkb3VvMEl3UURBT0JnTlZIUThCQWY4RQpCQU1DQXFRd0R3WURWUjBUQVFIL0JBVXdBd0VCL3pBZEJnTlZIUTRFRmdRVXQ2RzBzV3ZYbXVnTCtyYlN4V2lHCk1oNjNGT0l3Q2dZSUtvWkl6ajBFQXdJRFNRQXdSZ0loQU54NU1RUCt4SHBpL0NxVm1BVzBzOXZhaTlxYVZqb0UKNmg4dEJpQWxZU1dZQWlFQXJLQTdiR2poRjByT0cvZTN5YUNTQmNxeFhLUHRCZUE1S2hWSEdDeHpqbGs9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K\n" +
			"	server: https://0.0.0.0:43295\n" +
			"  name: k3d-mycluster\n" +
			"contexts:\n" +
			"- context:\n" +
			"	cluster: k3d-mycluster\n" +
			"	user: admin@k3d-mycluster\n" +
			"  name: k3d-mycluster\n" +
			"current-context: k3d-mycluster\n" +
			"kind: Config\n" +
			"preferences: {}\n" +
			"users:\n" +
			"- name: admin@k3d-mycluster\n" +
			"  user:\n"
		os.WriteFile(path, []byte(sa), 0644)
	}
	for {
		returnedVal := GenericRestCall("kubeconfig", path)
		if returnedVal == "200" {
			break
		} else {
			fmt.Println(returnedVal)
		}
	}
}

func GenericRestCall(requestKind string, body string) string {
	var returnValue, _ = json.Marshal("Error")
	request := GenericRequestType{Kind: requestKind, Body: body}

	bodyBytes, err := json.Marshal(request)
	reader := bytes.NewReader(bodyBytes)
	req, err := http.NewRequest(http.MethodPost, InterLinkConfigInst.Interlinkurl+":"+InterLinkConfigInst.Interlinkport+"/genericCall", reader)

	if err != nil {
		log.Println(err)
	}

	token, err := os.ReadFile(InterLinkConfigInst.VKTokenFile) // just pass the file name
	req.Header.Add("Authorization", "Bearer "+string(token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		time.Sleep(5 * time.Second)
	} else {
		returnValue, _ = ioutil.ReadAll(resp.Body)

		if string(returnValue) == "200" {
			return "200"
		}
	}

	return "400"
}
