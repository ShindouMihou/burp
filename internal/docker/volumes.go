package docker

import (
	"context"
)

func HasVolume(name string) (bool, error) {
	return has(func() (any, error) {
		return Client.VolumeInspect(context.Background(), name)
	})
}
