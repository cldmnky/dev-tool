package image

import (
	"context"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

// BuildImage builds a docker image
func BuildImage(dockerFilePath string, buildContextPath string, tag string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(300)*time.Second)
	defer cancel()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	buildOpts := types.ImageBuildOptions{
		Dockerfile: dockerFilePath,
		Tags:       []string{tag},
	}

	buildCtx, _ := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})

	resp, err := cli.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)

	imagePushOpts := types.ImagePushOptions{} // TODO add auth creds
	pushResp, err := cli.ImagePush(ctx, tag, imagePushOpts)
	if err != nil {
		return err
	}
	defer pushResp.Close()
	return nil
}
