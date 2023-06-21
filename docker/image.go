package docker

import (
	"bufio"
	"burp/utils"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
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

func Pull(image string) error {
	logger := log.With().Str("pull", image).Logger()
	response, err := Client.ImagePull(context.TODO(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func(response io.ReadCloser) {
		err := response.Close()
		if err != nil {
			logger.Err(err).Str("origin", "pull").Msg("Failed to close body")
		}
	}(response)
	return Handle(logger, bufio.NewScanner(response))
}

func Build(dir string, name string) error {
	tar, err := archive.TarWithOptions(dir, &archive.TarOptions{})
	logger := log.With().Str("build", name).Logger()
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
	if err = Handle(logger, bufio.NewScanner(build.Body)); err != nil {
		return err
	}
	logger.Info().Msg("Created Image")
	return nil
}
