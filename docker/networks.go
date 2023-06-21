package docker

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/errdefs"
)

func HasNetwork(name string) (bool, error) {
	_, err := Client.NetworkInspect(context.TODO(), name, types.NetworkInspectOptions{})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
