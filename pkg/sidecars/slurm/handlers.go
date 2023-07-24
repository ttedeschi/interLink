package slurm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	exec "github.com/alexellis/go-execute/pkg/v1"
	commonIL "github.com/intertwin-eu/interlink/pkg/common"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var JID []JidStruct

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Slurm Sidecar: received Submit call")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var req commonIL.Request
	json.Unmarshal(bodyBytes, &req)
	if err != nil {
		log.Println(err)
		return
	}

	if os.Getenv("KUBECONFIG") == "" {
		time.Sleep(time.Second)
	}

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		log.Println("Unable to create a valid config")
	}
	clientset, err = kubernetes.NewForConfig(config)

	for _, pod := range req.Pods {
		var metadata metav1.ObjectMeta
		var containers []v1.Container

		containers = pod.Spec.Containers
		metadata = pod.ObjectMeta

		for _, container := range containers {
			log.Println("Beginning script generation for container " + container.Name)
			commstr1 := []string{"singularity", "exec"}

			envs := prepare_envs(container)
			image := ""
			mounts := prepare_mounts(container, pod)
			if strings.HasPrefix(container.Image, "/") {
				if image_uri, ok := metadata.Annotations["slurm-job.knoc.io/image-root"]; ok {
					log.Println(image_uri)
					image = image_uri + container.Image
				} else {
					log.Println("image-uri annotation not specified for path in remote filesystem")
				}
			} else {
				image = "docker://" + container.Image
			}
			image = container.Image

			log.Println("Appending all commands together...")
			singularity_command := append(commstr1, envs...)
			singularity_command = append(singularity_command, mounts...)
			singularity_command = append(singularity_command, image)
			singularity_command = append(singularity_command, container.Command...)
			singularity_command = append(singularity_command, container.Args...)

			path := produce_slurm_script(container, metadata, singularity_command)
			out := slurm_batch_submit(path)
			handle_jid(container, out, *pod)
			log.Println(out)

			jid, err := os.ReadFile(commonIL.InterLinkConfigInst.DataRootFolder + container.Name + ".jid")
			if err != nil {
				log.Println("Unable to read JID from file")
			}
			JID = append(JID, JidStruct{JID: string(jid), Pod: *pod})
		}
	}

	w.Write([]byte(nil))
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Slurm Sidecar: received Stop call")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var req commonIL.Request
	err = json.Unmarshal(bodyBytes, &req)
	if err != nil {
		log.Println(err)
		return
	}

	for _, pod := range req.Pods {
		containers := pod.Spec.Containers

		for _, container := range containers {
			delete_container(container)
		}
	}
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Slurm Sidecar: received GetStatus call")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	var req commonIL.Request
	var resp commonIL.StatusResponse
	json.Unmarshal(bodyBytes, &req)
	if err != nil {
		log.Println(err)
		return
	}

	cmd := []string{"--me"}
	shell := exec.ExecTask{
		Command: "squeue",
		Args:    cmd,
		Shell:   true,
	}
	execReturn, err := shell.Execute()
	execReturn.Stdout = strings.ReplaceAll(execReturn.Stdout, "\n", "")

	if execReturn.Stderr != "" {
		log.Println("Unable to retrieve job status: " + execReturn.Stderr)
	}

	for _, pod := range req.Pods {
		var flag = false
		for _, jid := range JID {
			resp.PodName = append(resp.PodName, commonIL.PodName{Name: string(pod.Name)})

			cmd := []string{"-c", "squeue --me | grep " + jid.JID}
			shell := exec.ExecTask{
				Command: "bash",
				Args:    cmd,
				Shell:   true,
			}
			execReturn, _ := shell.Execute()

			if execReturn.Stderr != "" {
				log.Println("Unable to retrieve job status: " + execReturn.Stderr)
			} else if execReturn.Stdout != "" {
				flag = true
				log.Println(execReturn.Stdout)
			}
		}

		if flag {
			resp.PodStatus = append(resp.PodStatus, commonIL.PodStatus{PodStatus: commonIL.RUNNING})
		} else {
			resp.PodStatus = append(resp.PodStatus, commonIL.PodStatus{PodStatus: commonIL.STOP})
		}
	}
	resp.ReturnVal = "Status"

	bodyBytes, _ = json.Marshal(resp)

	w.Write(bodyBytes)
}

func SetKubeCFGHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Slurm Sidecar: received SetKubeCFG call")
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
