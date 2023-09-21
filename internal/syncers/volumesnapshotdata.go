package syncers

import (
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

func NewSnapshotDataSyncer(ctx *synccontext.RegisterContext) syncer.Base {
	return &snapshotDataSyncer{
		NamespacedTranslator: translator.NewNamespacedTranslator(
			ctx,
			"volumesnapshotdata",
			&snapshotv1.VolumeSnapshotData{},
		),
	}
}

type snapshotDataSyncer struct {
	translator.NamespacedTranslator
}

var _ syncer.Initializer = &snapshotDataSyncer{}

func (s *snapshotDataSyncer) Init(ctx *synccontext.RegisterContext) error {
	if err := translate.EnsureCRDFromPhysicalCluster(
		ctx.Context,
		ctx.PhysicalManager.GetConfig(),
		ctx.VirtualManager.GetConfig(),
		snapshotv1.SchemeGroupVersion.WithKind("VolumeSnapshotData"),
	); err != nil {
		return errors.Wrap(err, "ensure CRD VolumeSnapshot from physical cluster")
	}

	return nil
}

func (s *snapshotDataSyncer) SyncDown(ctx *synccontext.SyncContext, vObj client.Object) (ctrl.Result, error) {
	return s.SyncDownCreate(ctx, vObj, s.TranslateMetadata(vObj).(*snapshotv1.VolumeSnapshotData))
}

func (s *snapshotDataSyncer) Sync(ctx *synccontext.SyncContext, pObj client.Object, vObj client.Object) (ctrl.Result, error) {
	return s.SyncDownUpdate(
		ctx,
		vObj,
		s.translateUpdate(pObj.(*snapshotv1.VolumeSnapshotData), vObj.(*snapshotv1.VolumeSnapshotData)),
	)
}

func (s *snapshotDataSyncer) translateUpdate(pObj, vObj *snapshotv1.VolumeSnapshotData) *snapshotv1.VolumeSnapshotData {
	var updated *snapshotv1.VolumeSnapshotData

	// check annotations & labels
	changed, updatedAnnotations, updatedLabels := s.TranslateMetadataUpdate(vObj, pObj)
	if changed {
		updated = newSnapshotDataIfNil(updated, pObj)
		updated.Labels = updatedLabels
		updated.Annotations = updatedAnnotations
	}

	// check spec
	if !equality.Semantic.DeepEqual(vObj.Spec, pObj.Spec) {
		updated = newSnapshotDataIfNil(updated, pObj)
		updated.Spec = vObj.Spec
	}

	return updated
}

func newSnapshotDataIfNil(
	updated *snapshotv1.VolumeSnapshotData,
	pObj *snapshotv1.VolumeSnapshotData,
) *snapshotv1.VolumeSnapshotData {
	if updated == nil {
		return pObj.DeepCopy()
	}
	return updated
}
