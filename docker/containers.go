package docker

import (
	"burp/services"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/errdefs"
)

func GetContainer(name string) (*types.ContainerJSON, error) {
	con, err := Client.ContainerInspect(context.TODO(), "/"+name)
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &con, nil
}

func Deploy(image string, environments []string, ctr *services.Container) (*string, error) {
	name := "burp." + ctr.Name
	liveContainer, err := GetContainer(name)
	if err != nil {
		return nil, err
	}
	if liveContainer != nil {
		fmt.Println("Removing the container ", name, " with id ", liveContainer.ID)
		err = Client.ContainerRemove(context.TODO(), liveContainer.ID, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return nil, err
		}
	}
	exposedPorts, portBindings := ctr.GetPorts()
	mounts, err := ctr.GetVolumes()
	for _, mnt := range mounts {
		if mnt.Type != mount.TypeVolume {
			continue
		}
		hasVolume, err := HasVolume(mnt.Source)
		if err != nil {
			return nil, err
		}
		if !hasVolume {
			fmt.Println("Cannot find a volume named ", mnt.Source, ", attempting to create.")
			_, err = Client.VolumeCreate(context.TODO(), volume.CreateOptions{Name: mnt.Source})
			if err != nil {
				return nil, err
			}
		}
	}
	networkConfig := network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}
	for _, net := range ctr.Networks {
		networkConfig.EndpointsConfig[net] = &network.EndpointSettings{
			NetworkID: net,
		}
		hasNetwork, err := HasNetwork(net)
		if err != nil {
			return nil, err
		}
		if !hasNetwork {
			fmt.Println("Cannot find a network named ", net, ", attempting to create.")
			_, err := Client.NetworkCreate(context.TODO(), net, types.NetworkCreate{
				CheckDuplicate: true,
			})
			if err != nil {
				return nil, err
			}
		}
	}
	if err != nil {
		return nil, err
	}
	for key, value := range ctr.Environment {
		environments = append(environments, fmt.Sprint(key, "=", value))
	}
	hasImage, err := HasImage(image)
	if err != nil {
		return nil, err
	}
	if !hasImage {
		fmt.Println("Cannot find the image ", image, ", attempting to pull.")
		if err = Pull("mongo"); err != nil {
			return nil, err
		}
	}
	response, err := Client.ContainerCreate(context.TODO(), &container.Config{
		Hostname:     ctr.Hostname,
		User:         ctr.User,
		ExposedPorts: exposedPorts,
		Tty:          true,
		Env:          environments,
		Cmd:          ctr.Command,
		Image:        image,
		WorkingDir:   ctr.WorkingDirectory,
		Entrypoint:   ctr.Entrypoint,
	}, &container.HostConfig{
		LogConfig:     container.LogConfig{},
		PortBindings:  portBindings,
		RestartPolicy: ctr.GetRestartPolicy(),
		DNS:           ctr.DNS,
		Resources:     ctr.GetResources(),
		Mounts:        mounts,
	}, &networkConfig, nil, name)
	if err != nil {
		return nil, err
	}
	return &response.ID, nil
}
