package burp

import (
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/docker"
	"burp/pkg/fileutils"
	"burp/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"math"
	"strconv"
	"strings"
)

type ContainerVolume struct {
	Type     string `toml:"type" json:"type"`
	Source   string `toml:"source" json:"source"`
	Target   string `toml:"target" json:"target"`
	ReadOnly bool   `toml:"readonly,omitempty" json:"readonly,omitempty"`
}

type ContainerResourceLimits struct {
	CPUs       *float32           `toml:"cpus,omitempty" json:"cpus,omitempty"`
	Memory     *ResourceSizeLimit `toml:"memory,omitempty" json:"memory,omitempty"`
	SwapMemory *ResourceSizeLimit `toml:"swap_memory,omitempty" json:"swap_memory,omitempty"`
}

type ResourceSizeLimit struct {
	uint64
}

type ContainerRestartPolicy struct {
	Name              string `toml:"name" json:"name"`
	MaximumRetryCount int    `toml:"maximum_retry_count,omitempty" json:"maximum_retry_count,omitempty"`
}

func (limit *ResourceSizeLimit) UnmarshalText(text []byte) error {
	if utils.IsNumeric(text) {
		i, err := strconv.ParseUint(string(text), 10, 64)
		if err != nil {
			return err
		}
		limit.uint64 = i
		return nil
	}
	d, err := datasize.Parse(text)
	if err != nil {
		return err
	}
	limit.uint64 = d.Bytes()
	return nil
}

func (ctr *Container) GetPorts() (nat.PortSet, nat.PortMap) {
	set := nat.PortSet{}
	portBindings := nat.PortMap{}
	for _, ports := range ctr.Ports {
		hostPort, ctrPort := nat.Port(ports[0]), nat.Port(ports[1])

		set[ctrPort] = struct{}{}
		portBindings[ctrPort] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: string(hostPort),
			},
		}
	}
	return set, portBindings
}

var SupportedVolumeTypes = []string{"volume", "bind"}
var ErrUnsupportedVolumeType = errors.New("container has invalid volume type, must be either: volume or bind")

func (ctr *Container) GetVolumes() ([]mount.Mount, error) {
	var mounts []mount.Mount
	for _, containerVolume := range ctr.Volumes {
		containerVolume := containerVolume
		if !utils.AnyMatchStringCaseInsensitive(SupportedVolumeTypes, containerVolume.Type) {
			return nil, ErrUnsupportedVolumeType
		}
		containerVolume.Type = strings.ToLower(containerVolume.Type)
		if containerVolume.Type == "bind" {
			translations := map[string]string{
				"$HOME": fileutils.GetHomeDir(),
			}
			for key, value := range translations {
				containerVolume.Source = strings.ReplaceAll(containerVolume.Source, key, value)
			}
		}
		mounts = append(mounts, mount.Mount{
			Type:     mount.Type(containerVolume.Type),
			Source:   containerVolume.Source,
			Target:   containerVolume.Target,
			ReadOnly: containerVolume.ReadOnly,
		})
	}
	return mounts, nil
}

func (ctr *Container) GetResources() container.Resources {
	resources := container.Resources{}
	if ctr.Limits != nil {
		if ctr.Limits.CPUs != nil {
			resources.CPUQuota = int64(math.Round(float64(*ctr.Limits.CPUs * float32(100_000))))
			resources.CPUPeriod = int64(math.Round(float64(*ctr.Limits.CPUs * float32(66_666.7))))
		}
		if ctr.Limits.Memory != nil {
			resources.Memory = int64(ctr.Limits.Memory.uint64)
		}
		if ctr.Limits.SwapMemory != nil {
			resources.MemorySwap = int64(ctr.Limits.SwapMemory.uint64)
		}
	}
	return resources
}

func (ctr *Container) GetRestartPolicy() container.RestartPolicy {
	policy := container.RestartPolicy{}
	if ctr.RestartPolicy != nil {
		policy.Name = ctr.RestartPolicy.Name
		policy.MaximumRetryCount = ctr.RestartPolicy.MaximumRetryCount
	}
	return policy
}

func (ctr *Container) exec(channel *chan any, logger *zerolog.Logger, contexts []string, task func(name string) error) bool {
	responses.Message(channel, contexts[0], " container (burp.", ctr.Name, ")....")
	if err := task("burp." + ctr.Name); err != nil {
		logger.Info().Err(err).Str("name", ctr.Name).Msg("Failed to " + contexts[1] + " container")
		responses.Error(channel, "Failed to "+contexts[1]+" container (burp."+ctr.Name+")", err)
		return false
	}
	responses.Message(channel, contexts[2], " container (burp.", ctr.Name, ")")
	return true
}

func (ctr *Container) Start(channel *chan any, logger *zerolog.Logger) bool {
	return ctr.exec(channel, logger, []string{"Starting", "start", "Started"}, docker.Start)
}

func (ctr *Container) Remove(channel *chan any, logger *zerolog.Logger) bool {
	return ctr.exec(channel, logger, []string{"Removing", "remove", "Removed"}, docker.Remove)
}

func (ctr *Container) Stop(channel *chan any, logger *zerolog.Logger) bool {
	return ctr.exec(channel, logger, []string{"Stopping", "stop", "Stopped"}, docker.Kill)
}

func (ctr *Container) Deploy(channel *chan any, image string, environments []string) (*string, error) {
	name := "burp." + ctr.Name
	logger := log.With().Str("name", ctr.Name).Logger()
	liveContainer, err := docker.GetContainer(name)
	if err != nil {
		return nil, err
	}
	if liveContainer != nil {
		logger.Warn().Str("id", liveContainer.ID).Msg("Removing Container")
		responses.ChannelSend(channel, responses.Create("Removing container with the id "+liveContainer.ID+" for container "+name))
		if err = docker.Remove(liveContainer.ID); err != nil {
			return nil, err
		}
	}
	exposedPorts, portBindings := ctr.GetPorts()
	mounts, err := ctr.GetVolumes()
	for _, mnt := range mounts {
		if mnt.Type != mount.TypeVolume {
			continue
		}
		hasVolume, err := docker.HasVolume(mnt.Source)
		if err != nil {
			return nil, err
		}
		if !hasVolume {
			logger := logger.With().Str("volume", mnt.Source).Logger()
			logger.Warn().Msg("Cannot find the volume specified")
			logger.Info().Msg("Creating volume")
			responses.ChannelSend(channel, responses.Create("Cannot find any volume named "+mnt.Source+", creating volume..."))
			_, err = docker.Client.VolumeCreate(context.TODO(), volume.CreateOptions{Name: mnt.Source})
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
		hasNetwork, err := docker.HasNetwork(net)
		if err != nil {
			return nil, err
		}
		if !hasNetwork {
			logger := logger.With().Str("network", net).Logger()
			logger.Warn().Msg("Cannot find the network specified")
			logger.Info().Msg("Creating network")
			responses.ChannelSend(channel, responses.Create("Cannot find any network named "+net+", creating network..."))
			_, err := docker.Client.NetworkCreate(context.TODO(), net, types.NetworkCreate{
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
	hasImage, err := docker.HasImage(image)
	if err != nil {
		return nil, err
	}
	if !hasImage {
		logger := logger.With().Str("image", image).Logger()
		logger.Warn().Msg("Cannot find the image specified")
		logger.Info().Msg("Pulling image")
		responses.ChannelSend(channel, responses.Create("Cannot find any image named "+image+", pulling image..."))
		if err = docker.Pull(channel, image); err != nil {
			return nil, err
		}
	}
	response, err := docker.Client.ContainerCreate(context.TODO(), &container.Config{
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
