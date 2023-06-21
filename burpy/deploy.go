package burpy

import (
	"burp/docker"
	"burp/services"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

func Deploy(burp *services.Burp) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	dir, err := burp.Service.Clone()
	if err != nil {
		log.Err(err).Msg("Cloning Service")
		return
	}
	err = burp.Environment.Save(*dir)
	if err != nil {
		log.Err(err).Msg("Saving Environment File")
		return
	}
	log.Info().Str("dir", *dir).Msg("Build Path")
	if err := docker.Build(filepath.Join(*dir, burp.Service.Build), burp.Service.Name); err != nil {
		log.Err(err).Msg("Building Image")
		return
	}
	environments, err := burp.Environment.Read(*dir)
	if err != nil {
		log.Err(err).Str("dir", *dir).Msg("Reading Environment")
		return
	}
	var spawn []string
	for _, dependency := range burp.Dependencies {
		dependency := dependency
		id, err := docker.Deploy(dependency.Image, []string{}, &dependency.Container)
		if err != nil {
			log.Err(err).Str("name", dependency.Name)
			return
		}
		log.Info().Str("name", dependency.Name).Str("id", *id).Msg("Spawning Container")
		spawn = append(spawn, *id)
	}
	id, err := docker.Deploy(burp.Service.GetImage(), environments, &burp.Service.Container)
	if err != nil {
		log.Err(err).Str("name", burp.Service.Name).Msg("Spawning Container")
		return
	}
	spawn = append(spawn, *id)
	log.Info().Str("name", burp.Service.Name).Str("id", *id).Msg("Spawned Container")
	for _, id := range spawn {
		err := docker.Client.ContainerStart(context.TODO(), id, types.ContainerStartOptions{})
		if err != nil {
			log.Err(err).Str("id", id).Msg("Starting Container")
			return
		}
		log.Info().Str("id", id).Msg("Started Container")
	}
}
