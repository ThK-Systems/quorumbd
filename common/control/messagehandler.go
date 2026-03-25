package control

import "context"

type MessageHandler interface {
	HandleMessage(ctx context.Context, msg ControlMessage)
}
