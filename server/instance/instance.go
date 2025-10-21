package instance

import "context"

type InstanceUpdate struct {
	JsonData string `json:"jsonData"`
}

type Instance interface {
	OnStop(ctx context.Context) error
	OnUpdate(ctx context.Context, param InstanceUpdate) error
}
