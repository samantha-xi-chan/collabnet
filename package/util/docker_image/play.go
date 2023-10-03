package docker_image

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
)

func PullImageBlo(ctx context.Context, imageName string) (e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		e = errors.Wrap(err, "NewClientWithOpts: ")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v\n", err)
		}
	}()

	log.Println("ImagePull starting")
	resp, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "cli.ImagePull: ")
	}
	defer resp.Close()

	_, err = io.Copy(os.Stdout, resp)
	if err != nil {
		return errors.Wrap(err, "io.Copy: ")
	}

	return nil
}

func IsImageExists(ctx context.Context, imageName string) (x bool, e error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		e = errors.Wrap(err, "NewClientWithOpts: ")
		return
	}
	defer func() {
		if err := cli.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v\n", err)
		}
	}()

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return false, err
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}

	return false, nil
}

func RemoveImage(ctx context.Context, imageName string) (err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {

	}

	imageInspect, _, err := cli.ImageInspectWithRaw(context.Background(), imageName)
	if err != nil {
		if client.IsErrNotFound(err) {
			fmt.Printf("Image '%s' not found.\n", imageName)
		} else {
			fmt.Printf("Failed to inspect image: %s\n", err)
		}
		os.Exit(1)
	}

	// Delete the image
	_, err = cli.ImageRemove(context.Background(), imageInspect.ID, types.ImageRemoveOptions{})
	if err != nil {
		fmt.Printf("Failed to remove image: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Image '%s' deleted successfully.\n", imageName)

	return nil
}
