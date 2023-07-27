package docker

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	exec "github.com/alexellis/go-execute/pkg/v1"
	"github.com/containerd/containerd/log"
	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.G(Ctx).Info("Docker Sidecar: received GetStatus call")
	var resp commonIL.StatusResponse

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.G(Ctx).Error(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.G(Ctx).Debug("- Getting status for container " + container.Name)
			cmd := []string{"ps -aqf name=" + container.Name}

			shell := exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, err := shell.Execute()
			execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")

			if err != nil {
				log.G(Ctx).Error(err)
				return
			}

			if execReturn.Stderr != "" {
				log.G(Ctx).Error("-- Failed to get status for " + container.Name + " : " + execReturn.Stdout)
			} else {
				log.G(Ctx).Info("-- Container " + container.Name + " is running")
			}

			resp.PodName = append(resp.PodName, commonIL.PodName{Name: pod.Name})
			log.G(Ctx).Debug(execReturn.Stderr, execReturn.Stdout)

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
	log.G(Ctx).Info("Docker Sidecar: received Create call")
	var execReturn exec.ExecResult
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.G(Ctx).Error(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.G(Ctx).Info("- Creating container " + container.Name)
			cmd := []string{"run", "-d", "--name", container.Name}

			if commonIL.InterLinkConfigInst.ExportPodData {
				cmd = append(cmd, prepare_mounts(container, pod))
			}

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
				log.G(Ctx).Error(err)
				return
			}

			if execReturn.Stderr != "" {
				log.G(Ctx).Error("Unable to create container " + container.Name + " : " + execReturn.Stderr)
			} else {
				log.G(Ctx).Info("-- Created container " + container.Name)
			}

			shell = exec.ExecTask{
				Command: "docker",
				Args:    []string{"ps", "-aqf", "name=^" + container.Name + "$"},
				Shell:   true,
			}

			execReturn, err = shell.Execute()
			execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")
			if execReturn.Stderr != "" {
				log.G(Ctx).Error("Failed to retrieve " + container.Name + " ID : " + execReturn.Stderr)
			} else if execReturn.Stdout == "" {
				log.G(Ctx).Error("Container name not found. Maybe creation failed?")
			} else {
				log.G(Ctx).Debug("-- Retrieved " + container.Name + " ID: " + execReturn.Stdout)
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
	log.G(Ctx).Info("Docker Sidecar: received Delete call")
	var execReturn exec.ExecResult
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.G(Ctx).Error(err)
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)

	for _, pod := range req.Pods {
		for _, container := range pod.Spec.Containers {
			log.G(Ctx).Debug("- Deleting container " + container.Name)
			cmd := []string{"stop", container.Name}
			shell := exec.ExecTask{
				Command: "docker",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, _ = shell.Execute()

			if execReturn.Stderr != "" {
				log.G(Ctx).Error("-- Error stopping container " + container.Name + ". Skipping its removing")
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
				log.G(Ctx).Error("-- Error deleting container " + container.Name)
			} else {
				log.G(Ctx).Info("- Deleted container " + container.Name)
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
	log.G(Ctx).Info("Docker Sidecar: received SetKubeCFG call")
	path := "/tmp/.kube/"
	retCode := "200"
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.G(Ctx).Error(err)
	}

	var req commonIL.GenericRequestType
	json.Unmarshal(bodyBytes, &req)

	log.G(Ctx).Debug("- Creating folder to save KubeConfig")
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.G(Ctx).Error(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.G(Ctx).Debug("-- Created folder")
	}
	log.G(Ctx).Debug("- Creating the actual KubeConfig file")
	config, err := os.Create(path + "config")
	if err != nil {
		log.G(Ctx).Error(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.G(Ctx).Debug("-- Created file")
	}
	log.G(Ctx).Debug("- Writing configuration to file")
	_, err = config.Write([]byte(req.Body))
	if err != nil {
		log.G(Ctx).Error(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.G(Ctx).Info("-- Written configuration")
	}
	defer config.Close()
	log.G(Ctx).Debug("- Setting KUBECONFIG env")
	err = os.Setenv("KUBECONFIG", path+"config")
	if err != nil {
		log.G(Ctx).Error(err)
		retCode = "500"
		w.Write([]byte(retCode))
		return
	} else {
		log.G(Ctx).Info("-- Set KUBECONFIG to " + path + "config")
	}

	kubeconfig, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		log.G(Ctx).Error("Unable to create a valid config")
		return
	}
	Clientset, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.G(Ctx).Fatalln("Unable to set up a clientset")
	}

	w.Write([]byte(retCode))
}
