/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package vfencoder

import (
	"errors"
	"sync"

	"gocv.io/x/gocv"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/logger"
)

var log = logger.GetLogger("activity-mr4r-vfencoder")

const (
	extension    = "Extension"
	input        = "RawVideoFrame"
	cFrameWidth  = "FrameWidth"
	cFrameHeight = "FrameHeight"
	output       = "EncodedVideoFrame"
)

type VideoFrameEncoder struct {
	metadata *activity.Metadata
	mux      sync.Mutex
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	aCSVParserActivity := &VideoFrameEncoder{
		metadata: metadata,
	}
	return aCSVParserActivity
}

func (a *VideoFrameEncoder) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *VideoFrameEncoder) Eval(ctx activity.Context) (done bool, err error) {

	ext, err := a.getExtension(ctx)
	if nil != err {
		return false, err
	}

	data, ok := ctx.GetInput(input).(string)
	if !ok {
		log.Warn("Invalid data, expecting string but get : ", ctx.GetInput("Video"))
	}

	frameWidth, ok := ctx.GetInput(cFrameWidth).(int)
	if !ok {
		return false, errors.New("Invalid video frame width! ")
	}

	frameHeight, ok := ctx.GetInput(cFrameHeight).(int)
	if !ok {
		return false, errors.New("Invalid video frame height! ")
	}

	img, _ := gocv.NewMatFromBytes(frameHeight, frameWidth, gocv.MatTypeCV8UC3, []byte(data))
	if img.Empty() {
		logger.Warn("(VideoFrameEncoder.Eval) img.Empty()")
		return false, nil
	}

	bytes, _ := gocv.IMEncode(ext, img)

	ctx.SetOutput(output, string(bytes))

	return true, nil
}

func (a *VideoFrameEncoder) getExtension(context activity.Context) (gocv.FileExt, error) {
	ext, exists := context.GetSetting(extension)
	if !exists {
		return "", errors.New("No image extension defined!")
	}

	switch ext {
	case ".png":
		return gocv.PNGFileExt, nil
	case ".jpg":
		return gocv.JPEGFileExt, nil
	case ".gif":
		return gocv.GIFFileExt, nil
	}
	return gocv.JPEGFileExt, nil
}
