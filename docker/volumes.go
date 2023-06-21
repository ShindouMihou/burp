package docker

import (
	"context"
	"github.com/docker/docker/errdefs"
)

func HasVolume(name string) (bool, error) {
	_, err := Client.VolumeInspect(context.TODO(), name)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
