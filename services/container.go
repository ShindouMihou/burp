package services

import (
	"burp/utils"
	"errors"
	"github.com/c2h5oh/datasize"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
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
var UnsupportedVolumeType = errors.New("container has invalid volume type, must be either: volume or bind")

func (ctr *Container) GetVolumes() ([]mount.Mount, error) {
	var mounts []mount.Mount
	for _, volume := range ctr.Volumes {
		volume := volume
		if !utils.AnyMatchStringCaseInsensitive(SupportedVolumeTypes, volume.Type) {
			return nil, UnsupportedVolumeType
		}
		mounts = append(mounts, mount.Mount{
			Type:     mount.Type(strings.ToLower(volume.Type)),
			Source:   volume.Source,
			Target:   volume.Target,
			ReadOnly: volume.ReadOnly,
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
