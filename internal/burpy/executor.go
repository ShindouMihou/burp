package burpy

import (
	"burp/internal/docker"
	"burp/internal/services"
	"burp/pkg/fileutils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

var TemporaryFilesFolder = filepath.Join(".burp", ".temp", ".files")

func Package(burp *services.Burp) error {
	var hashes []services.HashedInclude
	dir := filepath.Join(TemporaryFilesFolder, burp.Service.Name)
	for _, include := range burp.Includes {
		name := filepath.Join(dir, "pkg", filepath.Base(include.Target))
		hash, err := fileutils.Copy(include.Source, name)
		if err != nil {
			return err
		}
		log.Info().Str("file", name).Str("hash", *hash).Msg("Copied File")
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
	log.Info().Str("file", "hashes.json").Msg("Saved File")
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
		filepath.Join(services.TemporaryCloneFolder, burp.Service.Name),
	}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}
	return nil
}

func Deploy(burp *services.Burp) {
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
