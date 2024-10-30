package managementuserlegacy

import (
	"context"

	"github.com/rancher/rancher/pkg/controllers/managementlegacy/compose/common"
	"github.com/rancher/rancher/pkg/controllers/managementuserlegacy/helm"
	"github.com/rancher/rancher/pkg/controllers/managementuserlegacy/systemimage"
	managementv3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/types/config"
)

func Register(ctx context.Context, mgmt *config.ScaledContext, cluster *config.UserContext, clusterRec *managementv3.Cluster, kubeConfigGetter common.KubeConfigGetter) error {
	helm.Register(ctx, mgmt, cluster, kubeConfigGetter)
	systemimage.Register(ctx, cluster)

	// register controller for API
	cluster.APIAggregation.APIServices("").Controller()
	return nil
}
