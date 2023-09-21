package syncers

import (
	"github.com/loft-sh/vcluster-sdk/plugin"
	"github.com/loft-sh/vcluster-sdk/syncer"
	synccontext "github.com/loft-sh/vcluster-sdk/syncer/context"
	"github.com/loft-sh/vcluster-sdk/syncer/translator"
	"github.com/loft-sh/vcluster-sdk/translate"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/equality"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	snapshotv1 "github.com/portworx/pxe-vcluster/apis/external-storage/snapshot/v1"
)

func init() {
	// Make sure our scheme is registered
	_ = snapshotv1.AddToScheme(plugin.Scheme)
}

func NewSnapshotSyncer(ctx *synccontext.RegisterContext) syncer.Base {
	return &snapshotSyncer{
		NamespacedTranslator: translator.NewNamespacedTranslator(
			ctx,
			"volumesnapshot",
			&snapshotv1.VolumeSnapshot{},
		),
	}
}

type snapshotSyncer struct {
	translator.NamespacedTranslator
}

var _ syncer.Initializer = &snapshotSyncer{}

func (s *snapshotSyncer) Init(ctx *synccontext.RegisterContext) error {
	if err := translate.EnsureCRDFromPhysicalCluster(
		ctx.Context,
		ctx.PhysicalManager.GetConfig(),
		ctx.VirtualManager.GetConfig(),
		snapshotv1.SchemeGroupVersion.WithKind("VolumeSnapshot"),
	); err != nil {
		return errors.Wrap(err, "ensure CRD VolumeSnapshot from physical cluster")
	}

	return nil
}

func (s *snapshotSyncer) SyncDown(ctx *synccontext.SyncContext, vObj client.Object) (ctrl.Result, error) {
	return s.SyncDownCreate(ctx, vObj, s.TranslateMetadata(vObj).(*snapshotv1.VolumeSnapshot))
}

func (s *snapshotSyncer) Sync(ctx *synccontext.SyncContext, pObj client.Object, vObj client.Object) (ctrl.Result, error) {
	return s.SyncDownUpdate(ctx, vObj, s.translateUpdate(pObj.(*snapshotv1.VolumeSnapshot), vObj.(*snapshotv1.VolumeSnapshot)))
}

func (s *snapshotSyncer) translateUpdate(pObj, vObj *snapshotv1.VolumeSnapshot) *snapshotv1.VolumeSnapshot {
	var updated *snapshotv1.VolumeSnapshot

	// check annotations & labels
	changed, updatedAnnotations, updatedLabels := s.TranslateMetadataUpdate(vObj, pObj)
	if changed {
		updated = newSnapshotIfNil(updated, pObj)
		updated.Labels = updatedLabels
		updated.Annotations = updatedAnnotations
	}

	// check spec
	if !equality.Semantic.DeepEqual(vObj.Spec, pObj.Spec) {
		updated = newSnapshotIfNil(updated, pObj)
		updated.Spec = vObj.Spec
	}

	return updated
}

func newSnapshotIfNil(updated *snapshotv1.VolumeSnapshot, pObj *snapshotv1.VolumeSnapshot) *snapshotv1.VolumeSnapshot {
	if updated == nil {
		return pObj.DeepCopy()
	}
	return updated
}
