// Copyright © 2021 FORTH-ICS
// Copyright © 2017 The virtual-kubelet authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"

	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/intertwin-eu/interlink/pkg/virtualkubelet"
	"github.com/sirupsen/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/log"
	logruslogger "github.com/virtual-kubelet/virtual-kubelet/log/logrus"
	"github.com/virtual-kubelet/virtual-kubelet/node"
)

type Config struct {
	ConfigPath        string
	NodeName          string
	OperatingSystem   string
	InternalIP        string
	DaemonPort        int32
	KubeClusterDomain string
}

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logrus.StandardLogger()
	log.L = logruslogger.FromLogrus(logrus.NewEntry(logger))

	cfg := Config{
		ConfigPath:      "",
		NodeName:        "",
		OperatingSystem: "Linux",
		InternalIP:      "127.0.0.1",
		DaemonPort:      10250,
	}

	kubecfg, _ := rest.InClusterConfig()

	localClient := kubernetes.NewForConfigOrDie(kubecfg)

	nodeProvider, _ := virtualkubelet.NewProvider(cfg.ConfigPath, cfg.NodeName, cfg.OperatingSystem, cfg.InternalIP, cfg.DaemonPort)

	nc, _ := node.NewNodeController(
		nodeProvider, nodeProvider.GetNode(), localClient.CoreV1().Nodes(),
	)

	// // https://github.com/liqotech/liqo/blob/master/cmd/virtual-kubelet/root/root.go#L195C49-L195C49
	// // https://github.com/liqotech/liqo/blob/master/cmd/virtual-kubelet/root/http.go#L76-L84
	// // https://github.com/liqotech/liqo/blob/master/cmd/virtual-kubelet/root/http.go#L93

	// err = setupHTTPServer(ctx, podProvider.PodHandler(), localClient, remoteConfig, c)
	// if err != nil {
	// log.G(ctx).Fatal(err)
	// }

	if err := nc.Run(ctx); err != nil {
		log.G(ctx).Fatal(err)
	}
}
