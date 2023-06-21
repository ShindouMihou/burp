package burpy

import (
	"burp/docker"
	"burp/services"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"log"
	"path/filepath"
)

func Deploy(burp *services.Burp) {
	dir, err := burp.Service.Clone()
	if err != nil {
		log.Fatalln(err)
		return
	}
	err = burp.Environment.Save(*dir)
	if err != nil {
		log.Fatalln(err)
		return
	}
	fmt.Println(*dir)
	if err := docker.Build(filepath.Join(*dir, burp.Service.Build), burp.Service.Name); err != nil {
		log.Fatalln(err)
		return
	}
	environments, err := burp.Environment.Read(*dir)
	if err != nil {
		log.Fatalln(err)
		return
	}
	var spawn []string
	for _, dependency := range burp.Dependencies {
		dependency := dependency
		id, err := docker.Deploy(dependency.Image, []string{}, &dependency.Container)
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println("Spawned the dependency ", dependency.Name, "(", id, ")")
		spawn = append(spawn, *id)
	}
	id, err := docker.Deploy(burp.Service.GetImage(), environments, &burp.Service.Container)
	if err != nil {
		log.Fatalln(err)
		return
	}
	spawn = append(spawn, *id)
	fmt.Println("Spawned the main application container (", id, ")")
	for _, id := range spawn {
		err := docker.Client.ContainerStart(context.TODO(), id, types.ContainerStartOptions{})
		if err != nil {
			log.Fatalln(err)
			return
		}
		fmt.Println("Started the container with id ", id)
	}
}
