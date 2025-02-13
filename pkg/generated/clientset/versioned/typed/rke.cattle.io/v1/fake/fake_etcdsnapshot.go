/*
Copyright 2025 Rancher Labs, Inc.

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

// FakeETCDSnapshots implements ETCDSnapshotInterface
type FakeETCDSnapshots struct {
	Fake *FakeRkeV1
	ns   string
}

var etcdsnapshotsResource = v1.SchemeGroupVersion.WithResource("etcdsnapshots")

var etcdsnapshotsKind = v1.SchemeGroupVersion.WithKind("ETCDSnapshot")

// Get takes name of the eTCDSnapshot, and returns the corresponding eTCDSnapshot object, and an error if there is any.
func (c *FakeETCDSnapshots) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ETCDSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(etcdsnapshotsResource, c.ns, name), &v1.ETCDSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ETCDSnapshot), err
}

// List takes label and field selectors, and returns the list of ETCDSnapshots that match those selectors.
func (c *FakeETCDSnapshots) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ETCDSnapshotList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(etcdsnapshotsResource, etcdsnapshotsKind, c.ns, opts), &v1.ETCDSnapshotList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ETCDSnapshotList{ListMeta: obj.(*v1.ETCDSnapshotList).ListMeta}
	for _, item := range obj.(*v1.ETCDSnapshotList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested eTCDSnapshots.
func (c *FakeETCDSnapshots) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(etcdsnapshotsResource, c.ns, opts))

}

// Create takes the representation of a eTCDSnapshot and creates it.  Returns the server's representation of the eTCDSnapshot, and an error, if there is any.
func (c *FakeETCDSnapshots) Create(ctx context.Context, eTCDSnapshot *v1.ETCDSnapshot, opts metav1.CreateOptions) (result *v1.ETCDSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(etcdsnapshotsResource, c.ns, eTCDSnapshot), &v1.ETCDSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ETCDSnapshot), err
}

// Update takes the representation of a eTCDSnapshot and updates it. Returns the server's representation of the eTCDSnapshot, and an error, if there is any.
func (c *FakeETCDSnapshots) Update(ctx context.Context, eTCDSnapshot *v1.ETCDSnapshot, opts metav1.UpdateOptions) (result *v1.ETCDSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(etcdsnapshotsResource, c.ns, eTCDSnapshot), &v1.ETCDSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ETCDSnapshot), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeETCDSnapshots) UpdateStatus(ctx context.Context, eTCDSnapshot *v1.ETCDSnapshot, opts metav1.UpdateOptions) (*v1.ETCDSnapshot, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(etcdsnapshotsResource, "status", c.ns, eTCDSnapshot), &v1.ETCDSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ETCDSnapshot), err
}

// Delete takes name of the eTCDSnapshot and deletes it. Returns an error if one occurs.
func (c *FakeETCDSnapshots) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(etcdsnapshotsResource, c.ns, name, opts), &v1.ETCDSnapshot{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeETCDSnapshots) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(etcdsnapshotsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ETCDSnapshotList{})
	return err
}

// Patch applies the patch and returns the patched eTCDSnapshot.
func (c *FakeETCDSnapshots) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ETCDSnapshot, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(etcdsnapshotsResource, c.ns, name, pt, data, subresources...), &v1.ETCDSnapshot{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.ETCDSnapshot), err
}
