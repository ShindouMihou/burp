package burpy

import (
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/docker"
	"burp/internal/services"
	"burp/pkg/fileutils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

var TemporaryFilesFolder = fileutils.JoinHomePath(".burpy", ".build", ".files")
var UnpackedFilesFolder = fileutils.JoinHomePath(".burpy", "home")

func Package(burp *services.Burp) error {
	var hashes []services.HashedInclude
	dir := filepath.Join(TemporaryFilesFolder, burp.Service.Name)
	for _, include := range burp.Includes {
		include := include
		name := filepath.Join(dir, "pkg", filepath.Base(include.Target))
		hash, err := fileutils.Copy(include.Source, name)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) && !include.Required {
				continue
			}
			return err
		}
		include.Source = filepath.Join("pkg", filepath.Base(include.Target))
		hashes = append(hashes, services.HashedInclude{Include: include, Hash: *hash})
	}
	marshal, err := json.Marshal(hashes)
	if err != nil {
		return err
	}
	err = fileutils.Save(filepath.Join(dir, "meta.json"), marshal)
	if err != nil {
		return err
	}
	tarName := fmt.Sprint(burp.Service.Name, "_includes.tar.gz")
	tarName = filepath.Join(TemporaryFilesFolder, ".packaged", tarName)
	err = fileutils.Tar(dir, tarName)
	if err != nil {
		return err
	}
	err = os.RemoveAll(dir)
	if err != nil {
		return err
	}
	return nil
}

// Clear cleans up all the deployment packages that were involved in this service.
func Clear(burp *services.Burp) error {
	paths := []string{
		filepath.Join(TemporaryFilesFolder, ".packaged", fmt.Sprint(burp.Service.Name, "_includes.tar.gz")),
		filepath.Join(TemporaryFilesFolder, burp.Service.Name),
		filepath.Join(services.TemporaryCloneFolder, burp.Service.Name),
	}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func Deploy(channel *chan any, burp *services.Burp) {
	dir, err := burp.Service.Clone()
	if err != nil {
		log.Err(err).Msg("Cloning Service")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to clone repository", err.Error()))
		return
	}
	err = burp.Environment.Save(*dir)
	if err != nil {
		log.Err(err).Msg("Saving Environment File")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to save environment file", err.Error()))
		return
	}
	log.Info().Str("dir", *dir).Msg("Build Path")
	if err := docker.Build(channel, filepath.Join(*dir, burp.Service.Build), burp.Service.Name); err != nil {
		log.Err(err).Msg("Building Image")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to save build image", err.Error()))
		return
	}
	environments, err := burp.Environment.Read(*dir)
	if err != nil {
		log.Err(err).Str("dir", *dir).Msg("Reading Environment")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to read environment properties", err.Error()))
		return
	}
	var spawn []string
	for _, dependency := range burp.Dependencies {
		dependency := dependency
		id, err := docker.Deploy(channel, dependency.Image, []string{}, &dependency.Container)
		if err != nil {
			log.Err(err).Str("name", dependency.Name)
			responses.ChannelSend(channel, responses.CreateChannelError("Failed to spawn dependency container "+dependency.Name, err.Error()))
			return
		}
		responses.ChannelSend(channel, responses.Create("Spawned container "+dependency.Name+" with id "+*id))
		log.Info().Str("name", dependency.Name).Str("id", *id).Msg("Spawning Container")
		spawn = append(spawn, *id)
	}
	id, err := docker.Deploy(channel, burp.Service.GetImage(), environments, &burp.Service.Container)
	if err != nil {
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to spawn container "+burp.Service.Name, err.Error()))
		log.Err(err).Str("name", burp.Service.Name).Msg("Spawning Container")
		return
	}
	spawn = append(spawn, *id)
	responses.ChannelSend(channel, responses.Create("Spawned container "+burp.Service.Name+" with id "+*id))
	log.Info().Str("name", burp.Service.Name).Str("id", *id).Msg("Spawned Container")
	for _, id := range spawn {
		err := docker.Client.ContainerStart(context.TODO(), id, types.ContainerStartOptions{})
		if err != nil {
			responses.ChannelSend(channel, responses.CreateChannelError("Failed to start container with id "+id, err.Error()))
			log.Err(err).Str("id", id).Msg("Starting Container")
			return
		}
		responses.ChannelSend(channel, responses.Create("started container with id "+id))
		log.Info().Str("id", id).Msg("Started Container")
	}
}
