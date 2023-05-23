package image

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"minik8s/pkg/kubelet/utils"
	"strings"
)

func imageFromLocal(client *containerd.Client, imageName string) containerd.Image {
	im, err := client.ImageService().Get(context.Background(), imageName)
	if err != nil {
		return nil
	}
	return containerd.NewImage(client, im)
}

func imageFromDefaultRegistry(client *containerd.Client, imageName string) containerd.Image {
	image, err := client.Pull(context.Background(), imageName, containerd.WithPullUnpack)
	if err != nil {
		return nil
	}
	return image
}
func EnsureImage(namespace string, client *containerd.Client, imageName string) containerd.Image {
	//has tag
	if !strings.Contains(imageName, ":") {
		imageName += ":latest"
	} else {
		if strings.Contains(imageName[strings.Index(imageName, ":"):], "/") {
			if !strings.Contains(imageName[strings.Index(imageName, ":")+1:], ":") {
				imageName += ":latest"
			}
		}
	}
	//fmt.Println(imageName)
	if strings.Contains(imageName, "master:5000") {
		output, err := utils.Ctl(namespace, "pull", "--insecure-registry", imageName)
		if err != nil {
			fmt.Println(output)
			return nil
		}
		return imageFromLocal(client, imageName)
	}
	if strings.Contains(imageName, "/") {
		return imageFromDefaultRegistry(client, imageName)
	}
	local := "master:5000/" + imageName
	_, err := utils.Ctl(namespace, "pull", "--insecure-registry", local)
	if err == nil {
		return imageFromLocal(client, local)
	}
	return imageFromDefaultRegistry(client, "docker.io/library/"+imageName)
}
