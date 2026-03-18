package app

import (
	"context"

	errorhelper "quorumbd.net/common/helper/errorhelper"
)

type Worker interface {
	run(parentCtx context.Context, workerExitCh chan<- WorkerExit) error
	restartOnCoreReconnect() bool
	getCoreConnectionEpoch() uint32
}

type WorkerExit struct {
	worker Worker
	kind   errorhelper.ExitKind
	err    error
}

func newWorkerExit(worker Worker, err error) WorkerExit {
	return WorkerExit{
		worker: worker,
		kind:   errorhelper.ClassifyError(err),
		err:    err,
	}
}
