package docker

import (
	"bufio"
	"burp/utils"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/archive"
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
	response, err := Client.ImagePull(context.TODO(), image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer func(response io.ReadCloser) {
		err := response.Close()
		if err != nil {
			fmt.Println("Failed to close Docker Pull body: ", err)
		}
	}(response)
	return Handle(bufio.NewScanner(response))
}

func Build(dir string, name string) error {
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
			fmt.Println("Failed to close Docker Build body: ", err)
		}
	}(build.Body)
	if err = Handle(bufio.NewScanner(build.Body)); err != nil {
		return err
	}
	fmt.Println("Successfully created burp/" + name)
	return nil
}
