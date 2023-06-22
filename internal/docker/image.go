package docker

import (
	"bufio"
	"burp/internal/auth"
	"burp/internal/server/responses"
	"burp/pkg/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/rs/zerolog/log"
	"io"
	"strings"
)

func HasImage(name string) (bool, error) {
	images, err := Client.ImageList(context.TODO(), types.ImageListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("reference", name)),
	})
	if err != nil {
		return false, err
	}
	for _, con := range images {
		con := con
		if utils.AnyMatch(con.RepoTags, func(b string) bool {
			return strings.Contains(b, name)
		}) {
			return true, nil
		}
	}
	return false, nil
}

func Pull(channel *chan any, image string) error {
	image = strings.ToLower(image)
	logger := log.With().Str("image", image).Logger()
	ref, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}
	domain := reference.Domain(ref)
	responses.ChannelSend(channel, responses.CreateChannelOk("Pulling image "+image+" from "+domain))
	logger.Info().Str("domain", domain).Msg("Pulling Image")

	var registryAuth string
	if creds, ok := auth.Docker[domain]; ok {
		config := &registry.AuthConfig{Username: creds.Username, Password: creds.Password}
		encoded, err := json.Marshal(config)
		if err != nil {
			return err
		}
		registryAuth = base64.URLEncoding.EncodeToString(encoded)
	}

	response, err := Client.ImagePull(context.TODO(), image, types.ImagePullOptions{
		RegistryAuth: registryAuth,
	})
	if err != nil {
		return err
	}
	defer func(response io.ReadCloser) {
		err := response.Close()
		if err != nil {
			logger.Err(err).Str("origin", "pull").Msg("Failed to close body")
		}
	}(response)
	return Handle(channel, logger, bufio.NewScanner(response))
}

func Build(channel *chan any, dir string, name string) error {
	logger := log.With().Str("build", name).Logger()

	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	if err != nil {
		return err
	}
	build, err := Client.ImageBuild(context.TODO(), tar, types.ImageBuildOptions{
		Tags: []string{fmt.Sprint("burp/", name)},
	})
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Err(err).Str("origin", "build").Msg("Failed to close body")
		}
	}(build.Body)
	if err = Handle(channel, logger, bufio.NewScanner(build.Body)); err != nil {
		return err
	}
	responses.ChannelSend(channel, responses.CreateChannelOk("Successfully created the image."))
	return nil
}
