/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package videoplayer

import (
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/logger"
)

var log = logger.GetLogger("activity-mr4r-videoplayer")

const (
	cConnection     = "Connection"
	cConnectionName = "name"
	input           = "Data"
)

type VideoPlayer struct {
	metadata    *activity.Metadata
	imgplayerIn io.WriteCloser
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	//imgplayer := exec.Command(
	//	"/Users/steven/Applications/MR4R/TelloFlogo/src/github.com/P-f1/LC/labs-mr4r/lib/imgplayer/imgplayer",
	//	strconv.Itoa(frameX), strconv.Itoa(frameY),
	//)
	imgplayer := exec.Command("/Applications/MPlayer OSX Extended.app/Contents/Resources/Binaries/mpextended.mpBinaries/Contents/MacOS/mplayer", "-fps", "60", "-")
	imgplayerIn, _ := imgplayer.StdinPipe()
	if err := imgplayer.Start(); err != nil {
		fmt.Println(err)
		return nil
	}

	imgProcessor := VideoPlayer{
		metadata:    metadata,
		imgplayerIn: imgplayerIn,
	}

	return &imgProcessor
}

func (a *VideoPlayer) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *VideoPlayer) Eval(ctx activity.Context) (done bool, err error) {
	log.Info("(VideoPlayer.Eval) Entering ...")
	defer log.Info("(VideoPlayer.Eval) Exit ...")

	video, ok := ctx.GetInput("Video").(string)
	if !ok {
		return false, errors.New("Invalid video ! ")
	}

	buf := []byte(video)

	if _, err := a.imgplayerIn.Write(buf); err != nil {
		fmt.Println(err)
	}

	return true, nil
}
