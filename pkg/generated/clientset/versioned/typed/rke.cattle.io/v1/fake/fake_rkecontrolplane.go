/*
Copyright 2024 Rancher Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package fake

import (
	"context"

	v1 "github.com/rancher/rancher/pkg/apis/rke.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeRKEControlPlanes implements RKEControlPlaneInterface
type FakeRKEControlPlanes struct {
	Fake *FakeRkeV1
	ns   string
}

var rkecontrolplanesResource = v1.SchemeGroupVersion.WithResource("rkecontrolplanes")

var rkecontrolplanesKind = v1.SchemeGroupVersion.WithKind("RKEControlPlane")

// Get takes name of the rKEControlPlane, and returns the corresponding rKEControlPlane object, and an error if there is any.
func (c *FakeRKEControlPlanes) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.RKEControlPlane, err error) {
	emptyResult := &v1.RKEControlPlane{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(rkecontrolplanesResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.RKEControlPlane), err
}

// List takes label and field selectors, and returns the list of RKEControlPlanes that match those selectors.
func (c *FakeRKEControlPlanes) List(ctx context.Context, opts metav1.ListOptions) (result *v1.RKEControlPlaneList, err error) {
	emptyResult := &v1.RKEControlPlaneList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(rkecontrolplanesResource, rkecontrolplanesKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.RKEControlPlaneList{ListMeta: obj.(*v1.RKEControlPlaneList).ListMeta}
	for _, item := range obj.(*v1.RKEControlPlaneList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested rKEControlPlanes.
func (c *FakeRKEControlPlanes) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(rkecontrolplanesResource, c.ns, opts))

}

// Create takes the representation of a rKEControlPlane and creates it.  Returns the server's representation of the rKEControlPlane, and an error, if there is any.
func (c *FakeRKEControlPlanes) Create(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.CreateOptions) (result *v1.RKEControlPlane, err error) {
	emptyResult := &v1.RKEControlPlane{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(rkecontrolplanesResource, c.ns, rKEControlPlane, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.RKEControlPlane), err
}

// Update takes the representation of a rKEControlPlane and updates it. Returns the server's representation of the rKEControlPlane, and an error, if there is any.
func (c *FakeRKEControlPlanes) Update(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.UpdateOptions) (result *v1.RKEControlPlane, err error) {
	emptyResult := &v1.RKEControlPlane{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(rkecontrolplanesResource, c.ns, rKEControlPlane, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.RKEControlPlane), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRKEControlPlanes) UpdateStatus(ctx context.Context, rKEControlPlane *v1.RKEControlPlane, opts metav1.UpdateOptions) (result *v1.RKEControlPlane, err error) {
	emptyResult := &v1.RKEControlPlane{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(rkecontrolplanesResource, "status", c.ns, rKEControlPlane, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.RKEControlPlane), err
}

// Delete takes name of the rKEControlPlane and deletes it. Returns an error if one occurs.
func (c *FakeRKEControlPlanes) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(rkecontrolplanesResource, c.ns, name, opts), &v1.RKEControlPlane{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRKEControlPlanes) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(rkecontrolplanesResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.RKEControlPlaneList{})
	return err
}

// Patch applies the patch and returns the patched rKEControlPlane.
func (c *FakeRKEControlPlanes) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.RKEControlPlane, err error) {
	emptyResult := &v1.RKEControlPlane{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(rkecontrolplanesResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.RKEControlPlane), err
}
