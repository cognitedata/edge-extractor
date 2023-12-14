package ffmpeg

import (
	"io"
	"os/exec"
)

type RTSPCamera struct {
	id        string
	username  string
	password  string
	streamUri string
	camera    *Camera
}

// Creates a new camera struct that can read from the device with the given stream index.
func NewRtspCamera(id, username, password, streamUri string) (*RTSPCamera, error) {
	// Check if ffmpeg is installed on the users machine.
	if err := checkExists("ffmpeg"); err != nil {
		return nil, err
	}

	camera := Camera{name: "rtsp", depth: 3, framerate: "30"}
	// if err := getCameraData(device, &camera); err != nil {
	// 	return nil, err
	// }
	rtspCamera := &RTSPCamera{id: id, camera: &camera, username: username, password: password, streamUri: streamUri}
	err := rtspCamera.GetCameraData()
	if err != nil {
		return nil, err
	}

	return rtspCamera, nil
}

func (cam *RTSPCamera) Camera() *Camera {
	return cam.camera
}

// Once the user calls Read() for the first time on a Camera struct,
// the ffmpeg command which is used to read the camera device is started.
func (cam *RTSPCamera) InitCamera() error {
	// If user exits with Ctrl+C, stop ffmpeg process.
	cam.camera.cleanup()

	// Use ffmpeg to pipe webcam to stdout.
	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "quiet",
		"-max_delay", "500000",
		"-rtsp_transport", "udp",
		"-i", "rtsp://"+cam.username+":"+cam.password+"@"+cam.streamUri,
		"-f", "image2pipe",
		"-r", "0.1", //  1 = 1HZ or frame per second.
		"-pix_fmt", "rgb24",
		"-vcodec", "rawvideo", "-",
	)

	cam.camera.cmd = cmd
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		pipe.Close()
		return err
	}

	cam.camera.pipe = &pipe
	if err := cmd.Start(); err != nil {
		cmd.Process.Kill()
		return err
	}

	cam.camera.framebuffer = make([]byte, cam.camera.width*cam.camera.height*cam.camera.depth)

	return nil
}

// Get camera meta data such as width, height, fps and codec.
func (cam *RTSPCamera) GetCameraData() error {
	// Run command to get camera data.
	cmd := exec.Command(
		"ffmpeg",
		"-hide_banner",
		"-i", "rtsp://"+cam.username+":"+cam.password+"@"+cam.streamUri,
	)
	// The command will fail since we do not give a file to write to, therefore
	// it will write the meta data to Stderr.
	pipe, err := cmd.StderrPipe()
	if err != nil {
		pipe.Close()
		return err
	}
	// Start the command.
	if err := cmd.Start(); err != nil {
		cmd.Process.Kill()
		return err
	}
	// Read ffmpeg output from Stdout.
	buffer := make([]byte, 2<<11)
	total := 0
	for {
		n, err := pipe.Read(buffer[total:])
		total += n
		if err == io.EOF {
			break
		}
	}
	// Wait for the command to finish.
	cmd.Wait()

	parseWebcamData(buffer[:total], cam.camera)
	return nil
}
