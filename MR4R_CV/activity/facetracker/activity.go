/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package facetracker

import (
	"errors"
	"image"
	"image/color"
	"math"
	"sync"

	"gocv.io/x/gocv"

	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/P-f1/LC/labs-mr4r/lib/util"
)

var log = logger.GetLogger("activity-mr4r-facetracker")

const (
	cModel       = "Model"
	cConfig      = "Config"
	cOutputImage = "OutputImage"
	cExtension   = "Extension"
	cInput       = "VideoFrame"
	cFrameWidth  = "FrameWidth"
	cFrameHeight = "FrameHeight"
)

type FaceTracker struct {
	mux      sync.Mutex
	metadata *activity.Metadata
	dectors  map[string]*FaceDector
}

func NewActivity(metadata *activity.Metadata) activity.Activity {
	tracker := FaceTracker{
		metadata: metadata,
		dectors:  make(map[string]*FaceDector),
	}

	return &tracker
}

func (a *FaceTracker) Metadata() *activity.Metadata {
	return a.metadata
}

func (a *FaceTracker) Eval(ctx activity.Context) (done bool, err error) {
	log.Info("(FaceTracker.Eval) Entering ...")
	defer log.Info("(FaceTracker.Eval) Exit ...")

	faceDector, err := a.getFaceDector(ctx)

	if nil != err {
		return false, err
	}

	image, ok := ctx.GetInput(cInput).(string)
	if !ok {
		return false, errors.New("Invalid video frame ! ")
	}

	frameWidth, ok := ctx.GetInput(cFrameWidth).(int)
	if !ok {
		return false, errors.New("Invalid video frame width! ")
	}

	frameHeight, ok := ctx.GetInput(cFrameHeight).(int)
	if !ok {
		return false, errors.New("Invalid video frame height! ")
	}

	img, _ := gocv.NewMatFromBytes(frameHeight, frameWidth, gocv.MatTypeCV8UC3, []byte(image))
	if img.Empty() {
		logger.Warn("(FaceTracker.Eval) img.Empty()")
		return false, nil
	}

	if detection := faceDector.detectFace(&img); nil != detection {
		log.Info("Detection.location : ", detection["location"])
		log.Info("         .delta    : ", detection["delta"])
		complexObj := &data.ComplexObject{Metadata: "Detection", Value: detection}
		ctx.SetOutput("Detection", complexObj)
		return true, nil
	}

	return false, nil
}

func (a *FaceTracker) getFaceDector(context activity.Context) (*FaceDector, error) {

	myId := util.ActivityId(context)
	dector := a.dectors[myId]
	log.Debug("%%%%%%%%%%%%%%%%%%%%%% getOutputFile : myId = ", myId, ", dector = ", dector)

	if nil == dector {
		a.mux.Lock()
		defer a.mux.Unlock()
		dector = a.dectors[myId]
		if nil == dector {
			log.Info("Initializing FaceDector start ...")

			model, _ := context.GetSetting(cModel)
			config, _ := context.GetSetting(cConfig)
			outputImage, _ := context.GetSetting(cOutputImage)
			imageExt, _ := context.GetSetting(cExtension)

			dector = NewFaceDector(model.(string), config.(string), outputImage.(bool), imageExt.(string))
			a.dectors[myId] = dector

			log.Info("Initializing FaceDector end ...")
		}
	}

	return dector, nil
}

func NewFaceDector(model string, config string, outputImage bool, ext string) *FaceDector {
	backend := gocv.NetBackendDefault
	target := gocv.NetTargetCPU

	n := gocv.ReadNet(model, config)
	if n.Empty() {
		logger.Errorf("Error reading network model from : %v %v\n", model, config)
		return nil
	}
	//	defer n.Close()
	n.SetPreferableBackend(gocv.NetBackendType(backend))
	n.SetPreferableTarget(gocv.NetTargetType(target))

	dector := FaceDector{
		imageExt:    ext,
		outputImage: outputImage,
		net:         &n,
		tracking:    true,
		detected:    false,
		detectSize:  true,
		green:       color.RGBA{0, 255, 0, 0},
	}

	return &dector
}

type FaceDector struct {
	metadata *activity.Metadata
	net      *gocv.Net

	// tracking
	outputImage              bool
	imageExt                 string
	tracking                 bool
	detected                 bool
	detectSize               bool
	green                    color.RGBA
	refDistance              float64
	left, top, right, bottom float64
}

func (a *FaceDector) detectFace(frame *gocv.Mat) map[string]interface{} {
	log.Info("(FaceTracker.trackFace) enter, frameWidth = ", frame.Cols(), ", frameHeight = ", frame.Rows())
	W := float64(frame.Cols())
	H := float64(frame.Rows())
	distTolerance := 0.05 * dist(0, 0, W, H)

	blob := gocv.BlobFromImage(*frame, 1.0, image.Pt(300, 300), gocv.NewScalar(104, 177, 123, 0), false, false)
	defer blob.Close()

	a.net.SetInput(blob, "data")

	detBlob := a.net.Forward("detection_out")
	defer detBlob.Close()

	detections := gocv.GetBlobChannel(detBlob, 0, 0)
	defer detections.Close()

	for r := 0; r < detections.Rows(); r++ {
		confidence := detections.GetFloatAt(r, 2)
		if confidence < 0.5 {
			continue
		}

		a.left = float64(detections.GetFloatAt(r, 3)) * W
		a.top = float64(detections.GetFloatAt(r, 4)) * H
		a.right = float64(detections.GetFloatAt(r, 5)) * W
		a.bottom = float64(detections.GetFloatAt(r, 6)) * H

		a.left = math.Min(math.Max(0.0, a.left), W-1.0)
		a.right = math.Min(math.Max(0.0, a.right), W-1.0)
		a.bottom = math.Min(math.Max(0.0, a.bottom), H-1.0)
		a.top = math.Min(math.Max(0.0, a.top), H-1.0)

		a.detected = true
		rect := image.Rect(int(a.left), int(a.top), int(a.right), int(a.bottom))
		gocv.Rectangle(frame, rect, a.green, 3)
	}

	if !a.tracking || !a.detected {
		return nil
	}

	if a.detectSize {
		a.detectSize = false
		a.refDistance = dist(a.left, a.top, a.right, a.bottom)
	}

	distance := dist(a.left, a.top, a.right, a.bottom)

	// x axis
	var deltaX float64
	switch {
	case a.right < W/2:
		deltaX = a.right - W/2
	case a.left > W/2:
		deltaX = a.left - W/2
	default:
		deltaX = 0
	}

	// y axis
	var deltaY float64
	switch {
	case a.top < H/10:
		deltaY = a.top - H/10
	case a.bottom > H-H/10:
		deltaY = a.bottom - (H - H/10)
	default:
		deltaY = 0
	}

	// z axis
	var deltaZ float64
	switch {
	case distance < a.refDistance-distTolerance:
		deltaZ = distance - (a.refDistance - distTolerance)
	case distance > a.refDistance+distTolerance:
		deltaZ = distance - (a.refDistance + distTolerance)
	default:
		deltaZ = 0
	}

	var image string
	if a.outputImage {
		bytes, _ := gocv.IMEncode(a.getExtension(a.imageExt), *frame)
		image = string(bytes)
	}

	return map[string]interface{}{
		"frame": map[string]interface{}{
			"image":  image,
			"height": H,
			"width":  W,
		},
		"location": map[string]interface{}{
			"left":     a.left,
			"top":      a.top,
			"right":    a.right,
			"bottom":   a.bottom,
			"distance": a.refDistance,
		},
		"delta": map[string]interface{}{
			"x": deltaX,
			"y": deltaY,
			"z": deltaZ,
		},
	}
}

func (a *FaceDector) getExtension(ext string) gocv.FileExt {
	switch ext {
	case ".png":
		return gocv.PNGFileExt
	case ".jpg":
		return gocv.JPEGFileExt
	case ".gif":
		return gocv.GIFFileExt
	}
	return gocv.JPEGFileExt
}

func dist(x1, y1, x2, y2 float64) float64 {
	return math.Sqrt((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1))
}
