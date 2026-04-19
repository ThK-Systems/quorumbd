// Package worker provides common worker functionality
package worker

import (
	"context"

	"github.com/google/uuid"

	"quorumbd.net/common/helper/errorhelper"

	"quorumbd.net/middleware-common/coreconnection"
)

type Worker interface {
	Run(parentCtx context.Context, workerExitCh chan<- WorkerExit, middlewareUUID uuid.UUID, coreEndpoint coreconnection.CoreEndpoint)
	RestartOnCoreReconnect() bool
	String() string
}

type WorkerExit struct {
	worker Worker
	kind   errorhelper.ExitKind
	err    error
}

func NewWorkerExit(worker Worker, err error) WorkerExit {
	return WorkerExit{
		worker: worker,
		kind:   errorhelper.ClassifyError(err),
		err:    err,
	}
}

func (we WorkerExit) GetKind() errorhelper.ExitKind {
	return we.kind
}

func (we WorkerExit) GetError() error {
	return we.err
}

func (we WorkerExit) String() string {
	result := "WorkerExit(worker: " + we.worker.String() + ", kind: " + we.kind.String()
	if we.err != nil {
		result += ", err: " + we.err.Error()
	}
	return result + ")"
}
