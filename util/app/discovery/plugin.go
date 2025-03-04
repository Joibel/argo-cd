package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	pluginclient "github.com/argoproj/argo-cd/v2/cmpserver/apiclient"
	"github.com/argoproj/argo-cd/v2/common"
	log "github.com/sirupsen/logrus"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type pluginType int

const (
	sidecar pluginType = iota
	service
)

type pluginService struct {
	servicePlugins []*plugin
	serviceMutex   sync.RWMutex
	informer       *informerscorev1.ServiceInformer
}

type plugin struct {
	name       string
	pluginType pluginType
	address    string
	owner      string
}

func (p *pluginType) clientSetType() pluginclient.ClientType {
	switch *p {
	case sidecar:
		return pluginclient.Sidecar
	case service:
		return pluginclient.Service
	default:
		log.Debugf("Unexpected pluginType %d", *p)
		return pluginclient.Sidecar
	}
}

func kubernetesClient() (*kubernetes.Clientset, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(cfg)
}

func newPluginService() *pluginService {
	ps := pluginService{
		servicePlugins: make([]*plugin, 0),
	}
	c, err := kubernetesClient()
	if err == nil {
		go ps.serviceWatcher(c)
	} else {
		// This is fine if this is what the user wants
		log.Warnf("Unable to uses plugins from services, ensure service account token is mounted (%s)", err)
	}
	return &ps
}

func (s *pluginService) getServicePlugins() ([]*plugin, error) {
	return s.servicePlugins, nil
}

func (s *pluginService) getSidecarPlugins() ([]*plugin, error) {
	plugins := make([]*plugin, 0)
	pluginSockFilePath := common.GetPluginSockFilePath()
	log.WithFields(log.Fields{
		common.SecurityField:    common.SecurityLow,
		common.SecurityCWEField: common.SecurityCWEMissingReleaseOfFileDescriptor,
	}).Debugf("pluginSockFilePath is: %s", pluginSockFilePath)

	fileList, err := os.ReadDir(pluginSockFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to list all plugins in dir, error=%w", err)
	}
	for _, file := range fileList {
		if file.Type() == os.ModeSocket {
			name, foundSock := strings.CutSuffix(file.Name(), `.sock`)
			if foundSock {
				plugins = append(plugins, &plugin{
					name:       name,
					pluginType: sidecar,
					address:    filepath.Join(pluginSockFilePath, file.Name()),
					owner:      file.Name(),
				})
			}
		}
	}
	return plugins, nil
}

func (s *pluginService) getAllPlugins() ([]*plugin, error) {
	s.serviceMutex.RLock()
	defer s.serviceMutex.RUnlock()
	servicePlugins, err := s.getServicePlugins()
	if err != nil {
		return nil, err
	}
	sidecarPlugins, err := s.getSidecarPlugins()
	if err != nil {
		return nil, err
	}
	return append(sidecarPlugins, servicePlugins...), nil
}

// Gets a plugin by name or returns nil if no such plugin
func (s *pluginService) getPluginByName(name string) (*plugin, error) {
	plugins, err := s.getAllPlugins()
	if err != nil {
		return nil, err
	}
	for _, plugin := range plugins {
		if name == plugin.name {
			return plugin, nil
		}
	}
	return nil, nil
}
