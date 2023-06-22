package docker

import (
	"burp/internal/server/responses"
	"burp/internal/services"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/errdefs"
	"github.com/rs/zerolog/log"
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

func Kill(name string) error {
	if err := Client.ContainerKill(context.TODO(), name, "SIGKILL"); err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func Start(name string) error {
	if err := Client.ContainerStart(context.TODO(), name, types.ContainerStartOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func Remove(name string) error {
	if err := Client.ContainerRemove(context.TODO(), name, types.ContainerRemoveOptions{Force: true}); err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func Deploy(channel *chan any, image string, environments []string, ctr *services.Container) (*string, error) {
	name := "burp." + ctr.Name
	logger := log.With().Str("name", ctr.Name).Logger()
	liveContainer, err := GetContainer(name)
	if err != nil {
		return nil, err
	}
	if liveContainer != nil {
		logger.Warn().Str("id", liveContainer.ID).Msg("Removing Container")
		responses.ChannelSend(channel, responses.CreateChannelOk("Removing container with the id "+liveContainer.ID+" for container "+name))
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
			logger := logger.With().Str("volume", mnt.Source).Logger()
			logger.Warn().Msg("Cannot find the volume specified")
			logger.Info().Msg("Creating volume")
			responses.ChannelSend(channel, responses.CreateChannelOk("Cannot find any volume named "+mnt.Source+", creating volume..."))
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
			logger := logger.With().Str("network", net).Logger()
			logger.Warn().Msg("Cannot find the network specified")
			logger.Info().Msg("Creating network")
			responses.ChannelSend(channel, responses.CreateChannelOk("Cannot find any network named "+net+", creating network..."))
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
		logger := logger.With().Str("image", image).Logger()
		logger.Warn().Msg("Cannot find the image specified")
		logger.Info().Msg("Pulling image")
		responses.ChannelSend(channel, responses.CreateChannelOk("Cannot find any image named "+image+", pulling image..."))
		if err = Pull(channel, "mongo"); err != nil {
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
