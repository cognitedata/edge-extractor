package ffmpeg

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type Camera struct {
	name        string         // Camera device name.
	width       int            // Camera frame width.
	height      int            // Camera frame height.
	depth       int            // Camera frame depth.
	fps         float64        // Camera frame rate.
	codec       string         // Camera codec.
	framebuffer []byte         // Raw frame data.
	pipe        *io.ReadCloser // Stdout pipe for ffmpeg process streaming webcam.
	cmd         *exec.Cmd      // ffmpeg command.
	framerate   string         // Framerate string for ffmpeg command.
}

func (camera *Camera) Name() string {
	return camera.name
}

func (camera *Camera) Width() int {
	return camera.width
}

func (camera *Camera) Height() int {
	return camera.height
}

func (camera *Camera) Depth() int {
	return camera.depth
}

func (camera *Camera) FPS() float64 {
	return camera.fps
}

func (camera *Camera) Codec() string {
	return camera.codec
}

func (camera *Camera) FrameBuffer() []byte {
	return camera.framebuffer
}

func (camera *Camera) SetFramerate(framerate string) {
	camera.framerate = framerate
}

// Sets the framebuffer to the given byte array. Note that "buffer" must be large enough
// to store one frame of video data which is width*height*3.
func (camera *Camera) SetFrameBuffer(buffer []byte) {
	camera.framebuffer = buffer
}

// Reads the next frame from the webcam and stores in the framebuffer.
func (camera *Camera) Read() bool {

	total := 0
	for total < camera.width*camera.height*camera.depth {
		if camera.pipe == nil {
			fmt.Println("Pipe is nil")
			return false
		}
		n, _ := (*camera.pipe).Read(camera.framebuffer[total:])
		total += n
		fmt.Println("Read ", n, " bytes")
	}
	return true
}

// Closes the pipe and stops the ffmpeg process.
func (camera *Camera) Close() {
	if camera.pipe != nil {
		(*camera.pipe).Close()
	}
	if camera.cmd != nil {
		camera.cmd.Process.Kill()
	}
}

// Stops the "cmd" process running when the user presses Ctrl+C.
// https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe.
func (camera *Camera) cleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if camera.pipe != nil {
			(*camera.pipe).Close()
		}
		if camera.cmd != nil {
			camera.cmd.Process.Kill()
		}
		os.Exit(1)
	}()
}
