package syncers

import (
	"context"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/loft-sh/vcluster-sdk/syncer"
	synccontext "github.com/loft-sh/vcluster-sdk/syncer/context"
)

const (
	pxLabelSelectorKey   = "name"
	pxLabelSelectorValue = "portworx-api"
)

func NewServiceSyncer(ctx *synccontext.RegisterContext) syncer.Base {
	return &pxServicesSyncer{}
}

type pxServicesSyncer struct{}

var _ syncer.Base = &pxServicesSyncer{}

func (s *pxServicesSyncer) Name() string {
	return "px-services-syncer"
}

var _ syncer.Initializer = &pxServicesSyncer{}

func (s *pxServicesSyncer) Init(ctx *synccontext.RegisterContext) error {
	vClient, err := client.New(ctx.VirtualManager.GetConfig(), client.Options{})
	if err != nil {
		return err
	}

	return updatePXService(ctx.Context, vClient)
}

func updatePXService(ctx context.Context, k8sClient client.Client) error {
	service := &v1.Service{}

	if err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: "kube-system",
		Name:      "portworx-api",
	}, service); err != nil {
		return errors.Wrap(err, "get portworx-api services")
	}

	if service.Labels == nil {
		service.Labels = map[string]string{}
	}

	service.Labels[pxLabelSelectorKey] = pxLabelSelectorValue

	if err := k8sClient.Update(ctx, service); err != nil {
		return errors.Wrap(err, "update portworx-api service")
	}

	return nil
}

func listPXServices(ctx context.Context, k8sClient client.Client) ([]v1.Service, error) {
	serviceList := &v1.ServiceList{}

	pxSvcReq, err := labels.NewRequirement(pxLabelSelectorKey, selection.Equals, []string{pxLabelSelectorValue})
	if err != nil {
		return nil, errors.Wrap(err, "create labels requirement for Px Services")
	}

	pxLabelSelector := labels.NewSelector()
	pxLabelSelector.Add(*pxSvcReq)

	if err := k8sClient.List(ctx, serviceList, &client.ListOptions{
		LabelSelector: pxLabelSelector,
	}); err != nil {
		return nil, errors.Wrap(err, "list services")
	}

	if len(serviceList.Items) == 0 {
		return nil, errors.New("no portworx services found")
	}

	return serviceList.Items, nil
}
