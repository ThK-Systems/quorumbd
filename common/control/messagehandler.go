package control

import "context"

type MessageHandler interface {
	HandleMessageBlocking(ctx context.Context, msg ControlMessage)
}
