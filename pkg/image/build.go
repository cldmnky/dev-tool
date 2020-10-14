package image

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

// BuildImage builds a docker image
func BuildImage(dockerFilePath string, buildContextPath string, tags []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerFilePath,
		Tags:       tags,
	}

	buildCtx, _ := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
	return nil
}

//writes from the build response to the log
func writeToLog(reader io.ReadCloser) error {
	defer reader.Close()
	rd := bufio.NewReader(reader)
	for {
		n, _, err := rd.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		log.Println(string(n))
	}
	return nil
}
