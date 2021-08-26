/*
 * Copyright © 2020. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package dockerimagebuilder

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"golang.org/x/net/context"
)

var log = logger.GetLogger("tibco-model-ops-cmdconverter")

var initialized bool = false

const (
	sAPIVersion    = "APIVersion"
	sUseRESTful    = "UseRESTful"
	sDockerHost    = "DockerHost"
	sDockerSocket  = "DockerSocket"
	iCommand       = "Command"
	iImageName     = "ImageName"
	iImageTag      = "ImageTag"
	iDockerFile    = "DockerFile"
	iWorkingFolder = "WorkingFolder"
	oSuccess       = "Success"
	oMessage       = "Message"
	oErrorCode     = "ErrorCode"
	CMD_Build      = "deploy"
	CMD_Push       = "update"
)

type DockerImageBuilderActivity struct {
	metadata   *activity.Metadata
	mux        sync.Mutex
	dockerEnvs map[string]*DockerEnv
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aDockerImageBuilderActivity := &DockerImageBuilderActivity{
		metadata: metadata,
	}

	return aDockerImageBuilderActivity
}

func (a *DockerImageBuilderActivity) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *DockerImageBuilderActivity) Eval(context activity.Context) (done bool, err error) {

	log.Info("[DockerImageBuilderActivity:Eval] entering ........ ")
	APIVersion, ok := context.GetSetting(sAPIVersion)
	if !ok {
		return false, errors.New("Invalid APIVersion ... ")
	}

	command, ok := context.GetInput(iCommand).(string)
	if !ok {
		return false, errors.New("Invalid command ... ")
	}

	imageName, ok := context.GetInput(iImageName).(string)
	if !ok {
		return false, errors.New("Invalid imageName ... ")
	}

	imageTag := context.GetInput(iImageTag)

	dockerFile, ok := context.GetInput(iDockerFile).(string)
	if !ok {
		return false, errors.New("Invalid dockerfile ... ")
	}

	workingFolder, ok := context.GetInput(iWorkingFolder).(string)
	if !ok {
		return false, errors.New("Invalid subfolder ... ")
	}

	dockerEnv, err := a.getDockerEnv(context)
	if nil != err {
		return false, err
	}

	log.Info("command : ", command)
	log.Info("APIVersion : ", APIVersion)
	log.Info("dockerFile : ", dockerFile)
	log.Info("imageName : ", imageName)
	log.Info("imageTag : ", imageTag)
	log.Info("host : ", dockerEnv.Host)
	log.Info("socket : ", dockerEnv.Socket)
	log.Info("workingFolder : ", workingFolder)

	cli, err := getClient(dockerEnv, APIVersion.(string))

	switch command {
	case CMD_Build:
		{
			err = buildImage(
				cli,
				imageName,
				imageTag,
				make(map[string]*string),
				workingFolder)
		}
	case CMD_Push:
		{
			err = pushImage(
				cli,
				imageName,
				imageTag)
		}
	}

	if nil != err {
		return false, nil
	}

	context.SetOutput(iCommand, command)
	log.Info("[DockerImageBuilderActivity:Eval] Exit ........ ")

	return true, nil
}

func (a *DockerImageBuilderActivity) getDockerEnv(ctx activity.Context) (*DockerEnv, error) {
	myId := ActivityId(ctx)
	dockerEnv := a.dockerEnvs[myId]

	if nil == dockerEnv {
		a.mux.Lock()
		defer a.mux.Unlock()
		dockerEnv = a.dockerEnvs[myId]
		if nil == dockerEnv {
			var ok bool

			useRESTful, exists := ctx.GetSetting(sUseRESTful)
			if !exists {
				useRESTful = false
			}
			host, ok := ctx.GetSetting(sDockerHost)
			if !ok && useRESTful.(bool) {
				return nil, errors.New("Invalid docker host ... ")
			}
			socket, ok := ctx.GetSetting(sDockerSocket)
			if !ok && !useRESTful.(bool) {
				return nil, errors.New("Invalid docker socket ... ")
			}
			dockerEnv = &DockerEnv{
				UseRESTful: useRESTful.(bool),
				Host:       host.(string),
				Socket:     socket.(string),
			}
		}
		a.dockerEnvs[myId] = dockerEnv
	}
	return dockerEnv, nil
}

func getClient(
	dockerEnv *DockerEnv,
	version string) (*client.Client, error) {

	var cli *client.Client
	var err error
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	if dockerEnv.UseRESTful {
		// Use RESTful
		httpClient := &http.Client{}
		cli, err = client.NewClient(dockerEnv.Host, version, httpClient, defaultHeaders)
	} else {
		// Use docker socket "unix:///var/run/docker.sock"
		cli, err = client.NewClient(dockerEnv.Socket, version, nil, defaultHeaders)
	}

	if err != nil {
		panic(err)
	}

	fmt.Print(cli.ClientVersion())
	return cli, nil
}

func buildImage(
	cli *client.Client,
	imageName string,
	tagName interface{},
	arguments map[string]*string,
	dockerFolder string) error {

	dockerBuildContext, err := os.Open(fmt.Sprintf("%s/%s.tar", dockerFolder, imageName))
	defer dockerBuildContext.Close()

	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Dockerfile:     fmt.Sprintf("%s/Dockerfile", dockerFolder),
		BuildArgs:      arguments,
	}

	if nil != tagName {
		buildOptions.Tags = []string{tagName.(string)}
	}

	buildResponse, err := cli.ImageBuild(context.Background(), dockerBuildContext, buildOptions)
	if err != nil {
		return err
	}
	defer buildResponse.Body.Close()

	return nil
}

func pushImage(
	cli *client.Client,
	imageName string,
	tagName interface{}) error {

	pushOptions := types.ImagePushOptions{}
	pushResponse, err := cli.ImagePush(context.Background(), fmt.Sprintf("%s:%s", imageName, tagName), pushOptions)
	if err != nil {
		return err
	}
	defer pushResponse.Close()

	return nil
}

/*
// Build the container using the native docker api
func Build(prebuild string, dockerfile string, arguments map[string]*string) error {
	dockerBuildContext, err := os.Open("/tmp/nodejs-distro.tar")
	defer dockerBuildContext.Close()
	cli, _ := client.NewClientWithOpts(client.WithVersion(defaultDockerAPIVersion))
	options := types.ImageBuildOptions{
		SuppressOutput: false,
		Remove:         true,
		ForceRemove:    true,
		PullParent:     true,
		Tags:           getTags(s),
		Dockerfile:     dockerfile,
		BuildArgs:      arguments,
	}
	buildResponse, err := cli.ImageBuild(context.Background(), dockerBuildContext, options)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	defer buildResponse.Body.Close()
	fmt.Printf("********* %s **********\n", buildResponse.OSType)

	termFd, isTerm := term.GetFdInfo(os.Stderr)
	return jsonmessage.DisplayJSONMessagesStream(buildResponse.Body, os.Stderr, termFd, isTerm, nil)
}
*/

func ActivityId(ctx activity.Context) string {
	return fmt.Sprintf("%s_%s", ctx.FlowDetails().Name(), ctx.TaskName())
}

type DockerEnv struct {
	UseRESTful bool
	Host       string
	Socket     string
}
