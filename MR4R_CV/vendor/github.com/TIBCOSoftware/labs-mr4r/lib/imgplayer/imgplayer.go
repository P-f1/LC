package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"gocv.io/x/gocv"
)

func main() {

	frameX, _ := strconv.Atoi(os.Args[1])
	frameY, _ := strconv.Atoi(os.Args[2])
	frameSize := frameX * frameY * 3

	window := gocv.NewWindow("Tello")
	for {
		fmt.Println("Beginning of the loop!!")
		buf := make([]byte, frameSize)
		if _, err := io.ReadFull(os.Stdin, buf); err != nil {
			fmt.Println(err)
			continue
		}
		img, _ := gocv.NewMatFromBytes(frameX, frameY, gocv.MatTypeCV8UC3, buf)
		if !img.Empty() {
			continue
		}
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}
}
