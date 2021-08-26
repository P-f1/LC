/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package imagewriter

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/P-f1/LC/flogo-lib/core/activity"
	"github.com/P-f1/LC/flogo-lib/logger"
	"github.com/P-f1/LC/labs-mr4r/lib/util"
)

var log = logger.GetLogger("activity-mr4r-imagewriter")

const (
	Input        = "Image"
	ToFileFolder = "toFileFolder"
	Filename     = "filename"
	Extension    = "extension"
)

type ImageWriter struct {
	mux         sync.Mutex
	metadata    *activity.Metadata
	outputFiles map[string]string
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	imgProcessor := ImageWriter{
		metadata:    metadata,
		outputFiles: make(map[string]string),
	}

	return &imgProcessor
}

func (a *ImageWriter) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *ImageWriter) Eval(ctx activity.Context) (done bool, err error) {
	log.Info("(ImageWriter.Eval) Entering ...")
	defer log.Info("(ImageWriter.Eval) Exit ...")

	img, ok := ctx.GetInput(Input).(string)
	if !ok {
		return false, errors.New("Invalid image ! ")
	}

	imgByte := []byte(img)
	imgObj, _, _ := image.Decode(bytes.NewReader(imgByte))
	fullFilename, _ := a.getOutputFile(ctx)

	image2File(fullFilename, imgObj)

	return true, nil
}

func (a *ImageWriter) getOutputFile(context activity.Context) (string, error) {

	myId := util.ActivityId(context)
	fullFilename := a.outputFiles[myId]
	log.Warn("(getOutputFile) myId = ", myId, ", fullFilename = ", fullFilename)

	if "" == fullFilename {
		a.mux.Lock()
		defer a.mux.Unlock()
		fullFilename = a.outputFiles[myId]
		if "" == fullFilename {
			log.Info("Initializing output folder start ...")

			folder, _ := context.GetSetting(ToFileFolder)
			filename, _ := context.GetSetting(Filename)
			ext, _ := context.GetSetting(Extension)

			fullFilename = filepath.Join(folder.(string), fmt.Sprintf("%s%s.%s", filename, "%d", ext))
			log.Info("Output fullFilename : ", fullFilename)

			err := os.MkdirAll(folder.(string), os.ModePerm)
			if nil != err {
				log.Error("Unable to create folder : ", err)
				return "", err
			}

			log.Info("Initializing FileWriter Service end ...")
			a.outputFiles[myId] = fullFilename
		}
	}

	return fullFilename, nil
}

func image2File(fullFilename string, img image.Image) error {
	f, err := os.Create(fmt.Sprintf(fullFilename, time.Now().Nanosecond()))
	if err != nil {
		return err
	}
	defer f.Close()

	switch strings.ToLower(fullFilename[strings.LastIndex(fullFilename, ".")+1:]) {
	case "jpeg", "jpg":
		opt := jpeg.Options{}
		err = jpeg.Encode(f, img, &opt)
	case "png":
		err = png.Encode(f, img)
	case "gif":
		opt := gif.Options{}
		err = gif.Encode(f, img, &opt)
	}

	return err
}
