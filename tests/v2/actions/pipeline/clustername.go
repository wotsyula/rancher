package pipeline

import (
	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/config"
)

// UpdateConfig is function that updates the cattle config's cluster name field which is
// the child of the rancher key in the cattle configuration.
func UpdateConfigClusterName(clusterName string) {
	rancherConfig := new(rancher.Config)
	config.LoadAndUpdateConfig(rancher.ConfigurationFileKey, rancherConfig, func() {
		rancherConfig.ClusterName = clusterName
	})
}
