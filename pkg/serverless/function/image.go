package function

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
)

const serverIp = "localhost"

// CreateImage to create image for function
func CreateImage(path string, name string) error {
	// 1. create the image
	// 1.1 copy the target file to the func.py
	srcFile, err := os.Open(path)
	if err != nil {
		log.Error("[CreateImage] open src file error: ", err)
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile("/home/mini-k8s/pkg/serverless/imagedata/func.py", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Error("[CreateImage] open dst file error: ", err)
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		log.Error("[CreateImage] copy file error: ", err)
		return err
	}

	// 1.2 create the image
	// TODO: here I use the absolute path, need to change to relative path
	cmd := exec.Command("docker", "build", "-t", name, "/home/mini-k8s/pkg/serverless/imagedata/")
	err = cmd.Run()
	if err != nil {
		log.Error("[CreateImage] create image error: ", err)
		return err
	}

	cmd = exec.Command("docker", "tag", name, serverIp+":5000/"+name+":latest")
	err = cmd.Run()
	if err != nil {
		log.Error("[CreateImage] tag image error: ", err)
		return err
	}

	// 2. save the image to the registry
	err = SaveImage(name)
	if err != nil {
		log.Error("[CreateImage] save image error: ", err)
		return err
	}

	return nil
}

// save the image to the registry
func SaveImage(name string) error {
	// 1. tag the image
	imageName := serverIp + ":5000/" + name + ":latest"

	// 2. push the image into the registry
	cmd := exec.Command("docker", "push", imageName)
	err := cmd.Run()
	if err != nil {
		log.Error("[SaveImage] push image error: ", err)
		return err
	}

	return nil
}

// DeleteImage to delete image for function
func DeleteImage(name string) error {
	imageName := serverIp + ":5000/" + name + ":latest"
	cmd := exec.Command("docker", "rmi", imageName)
	err := cmd.Run()
	if err != nil {
		log.Error("[DeleteImage] delete image error: ", err)
		return err
	}
	return nil
}

// RunImage to run image for function
// TODO: need change to containerd
func RunImage(name string) error {
	// 1. run the image
	cmd := exec.Command("docker", "run", "-d", "--name", name, "localhost:5000/"+name+":latest")
	err := cmd.Run()
	if err != nil {
		log.Error("[RunImage] run image error: ", err)
		return err
	}
	return nil
}
