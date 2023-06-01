package function

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"strings"
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


// find the image
func FindImage(name string) bool {
	cmd := exec.Command("docker", "images", name)

	// check the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error("[FindImage] get output error: ", err)
		return false
	}

	result := strings.TrimSpace(string(output))
	log.Info("[FindImage] the result is: ", result)

	if strings.Contains(result, name) {
		return true
	} else {
		return false
	}
}


// DeleteImage to delete image for function
func DeleteImage(name string) error {
	// if the image not exist, just ignore
	imageName := serverIp + ":5000/" + name + ":latest"
	if FindImage(imageName) {
		cmd := exec.Command("docker", "rmi", imageName)
		err := cmd.Run()
		if err != nil {
			log.Error("[DeleteImage] delete first image error: ", err)
			return err
		}
	}
	
	if FindImage(name) {
		cmd := exec.Command("docker", "rmi", name + ":latest") 
		err := cmd.Run()
		if err != nil {
			log.Error("[DeleteImage] delete second image error: ", err)
			return err
		}
	}
	
	log.Info("[DeleteImage] delete image finished")
	return nil
}

// RunImage to run image for function
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
