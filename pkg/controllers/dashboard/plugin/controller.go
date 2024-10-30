package plugin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/rancher/rancher/pkg/namespace"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/wrangler"
	"github.com/sirupsen/logrus"

	v1 "github.com/rancher/rancher/pkg/apis/catalog.cattle.io/v1"
	plugincontroller "github.com/rancher/rancher/pkg/generated/controllers/catalog.cattle.io/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func Register(
	ctx context.Context,
	wContext *wrangler.Context,
) {
	h := &handler{
		systemNamespace: namespace.UIPluginNamespace,
		plugin:          wContext.Catalog.UIPlugin(),
		pluginCache:     wContext.Catalog.UIPlugin().Cache(),
	}
	wContext.Catalog.UIPlugin().OnChange(ctx, "on-ui-plugin-change", h.OnPluginChange)
}

type handler struct {
	systemNamespace string
	plugin          plugincontroller.UIPluginController
	pluginCache     plugincontroller.UIPluginCache
}

func (h *handler) OnPluginChange(key string, plugin *v1.UIPlugin) (*v1.UIPlugin, error) {
	cachedPlugins, err := h.pluginCache.List(h.systemNamespace, labels.Everything())
	if err != nil {
		return plugin, fmt.Errorf("failed to list plugins from cache: %w", err)
	}
	err = Index.Generate(cachedPlugins)
	if err != nil {
		return plugin, fmt.Errorf("failed to generate index with cached plugins: %w", err)
	}
	var anonymousCachedPlugins []*v1.UIPlugin
	for _, cachedPlugin := range cachedPlugins {
		if cachedPlugin.Spec.Plugin.NoAuth {
			anonymousCachedPlugins = append(anonymousCachedPlugins, cachedPlugin)
		}
	}
	err = AnonymousIndex.Generate(anonymousCachedPlugins)
	if err != nil {
		return plugin, fmt.Errorf("failed to generate anonymous index with cached plugins: %w", err)
	}
	pattern := FSCacheRootDir + "/*/*"
	fsCacheFiles, err := fsCacheFilepathGlob(pattern)
	if err != nil {
		return plugin, fmt.Errorf("failed to get files from filesystem cache: %w", err)
	}
	FsCache.SyncWithIndex(&Index, fsCacheFiles)
	if plugin == nil {
		return plugin, nil
	}
	defer h.plugin.UpdateStatus(plugin)
	if plugin.Spec.Plugin.NoCache {
		plugin.Status.CacheState = Disabled
	} else {
		plugin.Status.CacheState = Pending
	}

	maxFileSize, err := strconv.ParseInt(settings.MaxUIPluginFileByteSize.Get(), 10, 64)
	if err != nil {
		logrus.Errorf("failed to convert setting MaxUIPluginFileByteSize to int64, using fallback. err: %s", err.Error())
		maxFileSize = settings.DefaultMaxUIPluginFileSizeInBytes
	}

	for _, p := range cachedPlugins {
		err2 := FsCache.SyncWithControllersCache(p)
		if errors.Is(err2, errMaxFileSizeError) {
			logrus.Errorf("one of the files is more than the defaultUIPluginFileByteSize limit %s", strconv.FormatInt(maxFileSize, 10))
			// update CRD to remove cache
			p.Spec.Plugin.NoCache = true
			_, err2 := h.plugin.Update(p)
			if err2 != nil {
				logrus.Errorf("failed to update plugin [%s] noCache flag: %s", p.Spec.Plugin.Name, err2.Error())
				continue
			}
			// delete files that were written
			err2 = FsCache.Delete(p.Spec.Plugin.Name, p.Spec.Plugin.Version)
			if err2 != nil {
				logrus.Error(err2)
				continue
			}
			p.Status.CacheState = Disabled
		} else {
			err = err2
		}
	}
	if err != nil {
		return plugin, fmt.Errorf("failed to sync filesystem cache with controller cache: %w", err)
	}
	if !plugin.Spec.Plugin.NoCache {
		plugin.Status.CacheState = Cached
	}

	return plugin, nil
}
