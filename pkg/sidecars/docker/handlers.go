package docker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	exec "github.com/alexellis/go-execute/pkg/v1"
	commonIL "github.com/intertwin-eu/interlink/pkg/common"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Docker Sidecar: received GetStatus call")
	var resp commonIL.StatusResponse

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.Println("Getting status for container " + container.Name)
			cmd := []string{"ps -aqf name= " + container.Name}

			shell := exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, err := shell.Execute()
			execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")

			if err != nil {
				log.Println(err)
				return
			}

			if execReturn.Stderr != "" {
				log.Println("Failed to get status for " + container.Name + " : " + execReturn.Stdout)
			} else {
				log.Println("Got status for " + container.Name)
			}

			resp.PodName = append(resp.PodName, commonIL.PodName{Name: pod.Name})
			log.Println(execReturn.Stderr, execReturn.Stdout)

			if execReturn.Stdout == "" {
				resp.PodStatus = append(resp.PodStatus, commonIL.PodStatus{PodStatus: commonIL.STOP})
			} else {
				resp.PodStatus = append(resp.PodStatus, commonIL.PodStatus{PodStatus: commonIL.RUNNING})
			}
		}
	}

	resp.ReturnVal = "Status"
	bodyBytes, _ = json.Marshal(resp)

	w.Write(bodyBytes)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Docker Sidecar: received Create call")
	var execReturn exec.ExecResult
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.Println("Creating container " + container.Name)
			cmd := []string{"run", "-d", "--name", container.Name}

			cmd = append(cmd, container.Image)

			for _, command := range container.Command {
				cmd = append(cmd, command)
			}
			for _, args := range container.Args {
				cmd = append(cmd, args)
			}

			shell := exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}

			execReturn, err = shell.Execute()
			if err != nil {
				log.Println(err)
				return
			}

			if execReturn.Stderr != "" {
				log.Println("Unable to create container " + container.Name + " : " + execReturn.Stderr)
			} else {
				log.Println("Successfully create container " + container.Name)
			}

			shell = exec.ExecTask{
				Command: "docker",
				Args:    []string{"ps", "-aqf", "name=^" + container.Name + "$"},
				Shell:   true,
			}

			execReturn, err = shell.Execute()
			execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")
			if execReturn.Stderr != "" {
				log.Println("Failed to retrieve " + container.Name + " ID : " + execReturn.Stderr)
			} else if execReturn.Stdout == "" {
				log.Println("Container name not found. Maybe creation failed?")
			} else {
				log.Println("Successfully retrieved " + container.Name + " ID: " + execReturn.Stdout)
			}
		}
	}
	if err != nil {
		w.Write([]byte(execReturn.Stderr))
	} else {
		w.Write([]byte("All containers for submitted Pods have been created"))
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Docker Sidecar: received Delete call")
	var execReturn exec.ExecResult
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.Println("Deleting container " + container.Name)
			cmd := []string{"stop", container.Name}
			shell := exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, _ = shell.Execute()

			if execReturn.Stderr != "" {
				log.Println("Error stopping container " + container.Name + ". Skipping its removing")
				continue
			}

			cmd = []string{"rm", execReturn.Stdout}
			shell = exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, _ = shell.Execute()
			execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")

			if execReturn.Stderr != "" {
				log.Println("Error deleting container " + container.Name)
			}
		}
	}

	if err != nil {
		w.Write([]byte(execReturn.Stderr))
	} else {
		w.Write([]byte("All containers for submitted Pods have been deleted"))
	}
}

func SetKubeCFGHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Docker Sidecar: received SetKubeCFG call")
	path := "/tmp/.kube/"
	retCode := "200"
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	var req commonIL.GenericRequestType
	json.Unmarshal(bodyBytes, &req)

	log.Println("Creating folder to save KubeConfig")
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Println(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.Println("Successfully created folder")
	}
	log.Println("Creating the actual KubeConfig file")
	config, err := os.Create(path + "config")
	if err != nil {
		log.Println(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.Println("Successfully created file")
	}
	log.Println("Writing configuration to file")
	_, err = config.Write([]byte(req.Body))
	if err != nil {
		log.Println(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.Println("Successfully written configuration")
	}
	defer config.Close()
	log.Println("Setting KUBECONFIG env")
	err = os.Setenv("KUBECONFIG", path+"config")
	if err != nil {
		log.Println(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.Println("Successfully set KUBECONFIG to " + path + "config")
	}

	w.Write([]byte(retCode))
}
