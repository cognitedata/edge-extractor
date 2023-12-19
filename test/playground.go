package main

import (
	"fmt"
	"time"

	"github.com/cognitedata/edge-extractor/pkg/ffmpeg"
)

func main() {

	// Prints "Hello, playground"
	println("Hello, playground")
	camera, err := ffmpeg.NewLocalCamera(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	println("Codec :", camera.Codec())
	// defer camera.Close()
	// cam, err := ffmpeg.NewRtspCamera("1", "admin", "kiborgi", "192.168.86.230:554/Streaming/Channels/1/")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// camera := cam.Camera()
	fmt.Println("Codec  :", camera.Codec())
	fmt.Println("Width  :", camera.Width())
	fmt.Println("Height :", camera.Height())
	fmt.Println("Depth :", camera.Depth())
	defer camera.Close()
	err = ffmpeg.InitLocalCamera(camera)
	// err = cam.InitCamera()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Stream the webcam
	go func() {
		for camera.Read() {
			fmt.Println("1) New frame received ")
			frame := camera.FrameBuffer()
			fmt.Println("1) New frame received ", len(frame)/1024*1024, " MB")
			err := ffmpeg.Write("cam1.jpg", camera.Width(), camera.Height(), frame)
			if err != nil {
				fmt.Println(err)
			}
		}
		// Video processing here...
	}()

	time.Sleep(time.Second * 5)

	// cam2, err := ffmpeg.NewRtspCamera("2", "admin", "kiborgi", "192.168.86.230:554/Streaming/Channels/2/")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// camera2 := cam2.Camera()
	// fmt.Println("Codec2  :", camera2.Codec())
	// fmt.Println("Width2  :", camera2.Width())
	// fmt.Println("Height2 :", camera2.Height())
	// fmt.Println("Depth2 :", camera2.Depth())
	// defer camera2.Close()

	// err = cam2.InitCamera()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// // Stream the webcam
	// for camera2.Read() {
	// 	frame := camera2.FrameBuffer()
	// 	fmt.Println("2) New frame received ", len(frame))
	// 	err := ffmpeg.Write("cam2.jpg", camera2.Width(), camera2.Height(), frame)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	// Video processing here...
	// }

}
