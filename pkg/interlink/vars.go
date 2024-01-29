package interlink

import (
	"context"

	"k8s.io/client-go/kubernetes"

	"github.com/intertwin-eu/interlink/pkg/common"
)

var Ctx context.Context
var Clientset *kubernetes.Clientset

type InterLinkHandler struct {
	Config common.InterLinkConfig
}
