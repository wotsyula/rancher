package multiclusterapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/rancher/rancher/pkg/ref"

	"github.com/rancher/norman/httperror"
	"github.com/rancher/norman/store/transform"
	"github.com/rancher/norman/types"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/norman/types/values"
	"github.com/rancher/rancher/pkg/auth/providers"
	"github.com/rancher/rancher/pkg/auth/requests"
	"github.com/rancher/rancher/pkg/catalog/manager"
	client "github.com/rancher/rancher/pkg/client/generated/management/v3"
	"github.com/rancher/rancher/pkg/clusterrouter"
	v3 "github.com/rancher/rancher/pkg/generated/norman/management.cattle.io/v3"
	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/rancher/pkg/types/config"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func SetMemberStore(ctx context.Context, schema *types.Schema, mgmt *config.ScaledContext) {
	providers.Configure(ctx, mgmt)
	userLister := mgmt.Management.Users("").Controller().Lister()

	t := &transform.Store{
		Store: schema.Store,
		Transformer: func(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, opt *types.QueryOptions) (map[string]interface{}, error) {
			membersMapSlice, found := values.GetSlice(data, "members")
			if !found {
				return data, nil
			}
			for _, m := range membersMapSlice {
				if userID, ok := m["userId"].(string); ok && userID != "" {
					u, err := userLister.Get("", userID)
					if err != nil {
						logrus.Errorf("problem retrieving user for member %v from cache during member transformation: %v", data, err)
						return data, nil
					}
					// filtering and keeping system user accounts out of the members list
					for _, pid := range u.PrincipalIDs {
						if strings.HasPrefix(pid, "system://") {
							if opt != nil && opt.Options["ByID"] == "true" {
								return nil, httperror.NewAPIError(httperror.NotFound, "resource not found")
							}
							return nil, nil
						}
					}
				}
			}
			return data, nil
		},
	}

	s := &Store{
		Store:                 t,
		auth:                  requests.NewAuthenticator(ctx, clusterrouter.GetClusterID, mgmt),
		crtbLister:            mgmt.Management.ClusterRoleTemplateBindings("").Controller().Lister(),
		prtbLister:            mgmt.Management.ProjectRoleTemplateBindings("").Controller().Lister(),
		grbLister:             mgmt.Management.GlobalRoleBindings("").Controller().Lister(),
		grLister:              mgmt.Management.GlobalRoles("").Controller().Lister(),
		templateVersionLister: mgmt.Management.CatalogTemplateVersions("").Controller().Lister(),
		users:                 mgmt.Management.Users(""),
		rtLister:              mgmt.Management.RoleTemplates("").Controller().Lister(),
		mcappLister:           mgmt.Management.MultiClusterApps("").Controller().Lister(),
		catalogManager:        mgmt.CatalogManager,
	}

	schema.Store = s
}

type Store struct {
	types.Store
	auth                  requests.Authenticator
	crtbLister            v3.ClusterRoleTemplateBindingLister
	prtbLister            v3.ProjectRoleTemplateBindingLister
	grbLister             v3.GlobalRoleBindingLister
	grLister              v3.GlobalRoleLister
	templateVersionLister v3.CatalogTemplateVersionLister
	users                 v3.UserInterface
	rtLister              v3.RoleTemplateLister
	mcappLister           v3.MultiClusterAppLister
	catalogManager        manager.CatalogManager
}

func (s *Store) Create(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	if err := s.validateChartCompatibility("", data); err != nil {
		return nil, err
	}

	if err := checkDuplicateTargets(data); err != nil {
		return nil, err
	}
	data, err := s.setDisplayName(apiContext, schema, data)
	if err != nil {
		return nil, err
	}
	if err = s.ensureRolesNotDisabled(apiContext, data); err != nil {
		return nil, err
	}
	return s.Store.Create(apiContext, schema, data)
}

func (s *Store) Update(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}, id string) (map[string]interface{}, error) {
	if err := s.validateChartCompatibility(id, data); err != nil {
		return nil, err
	}

	data, err := s.setDisplayName(apiContext, schema, data)
	if err != nil {
		return nil, err
	}
	if err = s.ensureRolesNotDisabled(apiContext, data); err != nil {
		return nil, err
	}
	return s.Store.Update(apiContext, schema, data, id)
}

func (s *Store) setDisplayName(apiContext *types.APIContext, schema *types.Schema, data map[string]interface{}) (map[string]interface{}, error) {
	membersMapSlice, found := values.GetSlice(data, client.MultiClusterAppFieldMembers)
	if !found {
		return data, nil
	}
	token, err := s.auth.TokenFromRequest(apiContext.Request)
	if err != nil {
		return nil, err
	}
	if token.AuthProvider == providers.LocalProvider {
		// getPrincipal if called, will be called on local auth provider, and will not find a user with external auth principal ID
		// hence returning, since the only thing this method does is setting display name by getting it from the external auth provider.
		return data, nil
	}
	for _, m := range membersMapSlice {
		if principalID, ok := m[client.MemberFieldUserPrincipalID].(string); ok && principalID != "" && !strings.HasPrefix(principalID, "local://") {
			princ, err := providers.GetPrincipal(principalID, *token)
			if err != nil {
				return nil, err
			}
			if princ.DisplayName != "" {
				values.PutValue(m, princ.DisplayName, "displayName")
			}
		}
	}
	return data, nil
}

func (s *Store) ensureRolesNotDisabled(apiContext *types.APIContext, data map[string]interface{}) error {
	roles := convert.ToStringSlice(data[client.MultiClusterAppFieldRoles])
	if len(roles) == 0 {
		return httperror.NewAPIError(httperror.InvalidBodyContent, "No roles provided")
	}
	for _, role := range roles {
		rt, err := s.rtLister.Get("", role)
		if err != nil {
			if apierrors.IsNotFound(err) {
				errMsg := fmt.Sprintf("Provided role %v does not exist", role)
				return httperror.NewAPIError(httperror.InvalidBodyContent, errMsg)
			}
			return err
		}
		if rt.Locked {
			errMsg := fmt.Sprintf("Cannot assigned locked role %v to multiclusterapp", role)
			return httperror.NewAPIError(httperror.InvalidBodyContent, errMsg)
		}
	}
	return nil
}

func (s *Store) validateChartCompatibility(id string, data map[string]interface{}) error {
	rawTempVersion, ok := data["templateVersionId"]
	if !ok {
		return nil
	}

	tempVersion := rawTempVersion.(string)

	parts := strings.Split(tempVersion, ":")
	if len(parts) != 2 {
		return httperror.NewAPIError(httperror.InvalidBodyContent, "invalid templateVersionId")
	}

	template, err := s.templateVersionLister.Get(namespace.GlobalNamespace, parts[1])
	if err != nil {
		return err
	}

	var mcappCurrentTemplateVersion string
	mcappNamespace, mcappID := ref.Parse(id)
	if mcappID != "" {
		mcapp, err := s.mcappLister.Get(mcappNamespace, mcappID)
		if err != nil {
			return httperror.NewAPIError(httperror.NotFound, fmt.Sprintf("cannot find multiclusterapp [%s:%s]", mcappNamespace, mcappID))
		}
		mcappTemplateVersionParts := strings.Split(mcapp.Spec.TemplateVersionName, "-")
		mcappCurrentTemplateVersion = mcappTemplateVersionParts[len(mcappTemplateVersionParts)-1]
	}

	if err := s.catalogManager.ValidateRancherVersion(template, mcappCurrentTemplateVersion); err != nil {
		return httperror.NewAPIError(httperror.InvalidBodyContent, err.Error())
	}

	targets, _ := data["targets"].([]interface{})
	for _, target := range targets {
		targetMap, ok := target.(map[string]interface{})
		if !ok {
			continue
		}
		parts := strings.SplitN(targetMap["projectId"].(string), ":", 2)
		if len(parts) != 2 {
			continue
		}
		if err := s.catalogManager.ValidateKubeVersion(template, parts[0]); err != nil {
			return httperror.NewAPIError(httperror.InvalidBodyContent, err.Error())
		}
	}
	return nil
}

func checkDuplicateTargets(data map[string]interface{}) error {
	targets, _ := values.GetSlice(data, "targets")
	projectIDs := map[string]bool{}
	for _, target := range targets {
		id := convert.ToString(values.GetValueN(target, "projectId"))
		if id != "" {
			if _, ok := projectIDs[id]; ok {
				return httperror.NewAPIError(httperror.InvalidBodyContent, fmt.Sprintf("duplicate projects in targets %s", id))
			}
			projectIDs[id] = true
		}
	}
	return nil
}
