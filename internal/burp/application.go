package burp

import (
	"burp/cmd/burp-agent/server/responses"
	"burp/internal/docker"
	"burp/pkg/fileutils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strconv"
)

var TemporaryFilesFolder = fileutils.JoinHomePath(".burpy", ".build", ".files")
var UnpackedFilesFolder = fileutils.JoinHomePath(".burpy", "home")

// From is a method generally used in streaming routes, it decodes the Toml file and checks if there is any error.
// If there is an error, it sends to the logger and to the stream.
func (application *Application) From(bits []byte, logger *zerolog.Logger, channel *chan any) bool {
	if err := toml.Unmarshal(bits, application); err != nil {
		logger.Err(err).Msg("Failed to parse TOML file into Burp services")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to parse TOML file into Burp services", err.Error()))
		return false
	}
	return true
}

// Package packages all the files that was declared to be transported to the server-side.
// This should be run only in the client-side (CLI).
func (application *Application) Package() error {
	var hashes []HashedInclude
	dir := filepath.Join(TemporaryFilesFolder, application.Service.Name)
	for _, include := range application.Includes {
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
		hashes = append(hashes, HashedInclude{Include: include, Hash: *hash})
	}
	marshal, err := json.Marshal(hashes)
	if err != nil {
		return err
	}
	err = fileutils.Save(filepath.Join(dir, "meta.json"), marshal)
	if err != nil {
		return err
	}
	tarName := fmt.Sprint(application.Service.Name, "_includes.tar.gz")
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

// CleanRemnants cleans up all the deployment remnants that were involved in this service.
func (application *Application) CleanRemnants() error {
	paths := []string{
		filepath.Join(TemporaryFilesFolder, ".packaged", fmt.Sprint(application.Service.Name, "_includes.tar.gz")),
		filepath.Join(TemporaryFilesFolder, application.Service.Name),
		filepath.Join(TemporaryCloneFolder, application.Service.Name),
	}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

// Deploy deploys the application, this builds the main service's image and spawns a container for it
// before spawning the dependency containers (pulling when needed) and then starting all the containers.
func (application *Application) Deploy(channel *chan any, environments []string) {
	dir, err := application.Service.Clone()
	if err != nil {
		log.Err(err).Msg("Cloning Service")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to clone repository", err.Error()))
		return
	}
	if application.Environment.ServerSide {
		environments, err = application.Environment.Read(*dir)
		if err != nil {
			log.Err(err).Str("dir", *dir).Msg("Reading Environment")
			responses.ChannelSend(channel, responses.CreateChannelError("Failed to read environment properties", err.Error()))
			return
		}
	} else {
		log.Info().Int("len", len(environments)).Msg("Using environment variables from client.")
		responses.ChannelSend(channel, responses.Create("Using environment variables from client with length of "+strconv.FormatInt(int64(len(environments)), 10)))
	}
	log.Info().Str("dir", *dir).Msg("Build Path")
	if err := docker.Build(channel, filepath.Join(*dir, application.Service.Build), application.Service.Name); err != nil {
		log.Err(err).Msg("Building Image")
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to save build image", err.Error()))
		return
	}
	var spawn []string
	for _, dependency := range application.Dependencies {
		dependency := dependency
		id, err := dependency.Container.Deploy(channel, dependency.Image, []string{})
		if err != nil {
			log.Err(err).Str("name", dependency.Name)
			responses.ChannelSend(channel, responses.CreateChannelError("Failed to spawn dependency container "+dependency.Name, err.Error()))
			return
		}
		if id != nil {
			responses.ChannelSend(channel, responses.Create("Spawned container "+dependency.Name+" with id "+*id))
			log.Info().Str("name", dependency.Name).Str("id", *id).Msg("Spawning Container")
			spawn = append(spawn, *id)
		}
	}
	id, err := application.Service.Container.Deploy(channel, application.Service.GetImage(), environments)
	if err != nil {
		responses.ChannelSend(channel, responses.CreateChannelError("Failed to spawn container "+application.Service.Name, err.Error()))
		log.Err(err).Str("name", application.Service.Name).Msg("Spawning Container")
		return
	}
	if id != nil {
		spawn = append(spawn, *id)
		responses.ChannelSend(channel, responses.Create("Spawned container "+application.Service.Name+" with id "+*id))
		log.Info().Str("name", application.Service.Name).Str("id", *id).Msg("Spawned Container")
	}
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
