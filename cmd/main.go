package main

import (
	"github.com/loft-sh/vcluster-sdk/plugin"

	"github.com/portworx/pxe-vcluster/internal/syncers"
)

func main() {
	ctx := plugin.MustInit()

	plugin.MustRegister(syncers.NewServiceSyncer(ctx))
	plugin.MustRegister(syncers.NewSnapshotSyncer(ctx))
	plugin.MustRegister(syncers.NewSnapshotDataSyncer(ctx))

	plugin.MustStart()
}
