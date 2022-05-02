package ffmpeg

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

type VideoWriter struct {
	filename   string          // Output filename.
	audio      string          // Audio filename.
	width      int             // Frame width.
	height     int             // Frame height.
	bitrate    int             // Output video bitrate.
	loop       int             // Number of times for GIF to loop.
	delay      int             // Delay of final frame of GIF. Default -1 (same delay as previous frame).
	macro      int             // Macroblock size for determining how to resize frames for codecs.
	fps        float64         // Frames per second for output video. Default 25.
	quality    float64         // Used if bitrate not given. Default 0.5.
	codec      string          // Codec to encode video with. Default libx264.
	audioCodec string          // Codec to encode audio with. Default aac.
	pipe       *io.WriteCloser // Stdout pipe of ffmpeg process.
	cmd        *exec.Cmd       // ffmpeg command.
}

// Optional parameters for VideoWriter.
type Options struct {
	Bitrate    int     // Bitrate.
	Loop       int     // For GIFs only. -1=no loop, 0=infinite loop, >0=number of loops.
	Delay      int     // Delay for final frame of GIFs.
	Macro      int     // Macroblock size for determining how to resize frames for codecs.
	FPS        float64 // Frames per second for output video.
	Quality    float64 // If bitrate not given, use quality instead. Must be between 0 and 1. 0:best, 1:worst.
	Codec      string  // Codec for video.
	Audio      string  // File path for audio. If no audio, audio="".
	AudioCodec string  // Codec for audio.
}

func (writer *VideoWriter) FileName() string {
	return writer.filename
}

func (writer *VideoWriter) Width() int {
	return writer.width
}

func (writer *VideoWriter) Height() int {
	return writer.height
}

func (writer *VideoWriter) Bitrate() int {
	return writer.bitrate
}

func (writer *VideoWriter) Loop() int {
	return writer.loop
}

func (writer *VideoWriter) Delay() int {
	return writer.delay
}

func (writer *VideoWriter) Macro() int {
	return writer.macro
}

func (writer *VideoWriter) FPS() float64 {
	return writer.fps
}

func (writer *VideoWriter) Quality() float64 {
	return writer.quality
}

func (writer *VideoWriter) Codec() string {
	return writer.codec
}

func (writer *VideoWriter) AudioCodec() string {
	return writer.audioCodec
}

// Creates a new VideoWriter struct with default values from the Options struct.
func NewVideoWriter(filename string, width, height int, options *Options) (*VideoWriter, error) {
	// Check if ffmpeg is installed on the users machine.
	if err := checkExists("ffmpeg"); err != nil {
		return nil, err
	}

	writer := VideoWriter{filename: filename}

	writer.width = width
	writer.height = height
	writer.bitrate = options.Bitrate

	// Default Parameter options logic from:
	// https://github.com/imageio/imageio-ffmpeg/blob/master/imageio_ffmpeg/_io.py#L268.

	// GIF settings
	writer.loop = options.Loop // Default to infinite loop.
	if options.Delay == 0 {
		writer.delay = -1 // Default to frame delay of previous frame.
	} else {
		writer.delay = options.Delay
	}

	if options.Macro == 0 {
		writer.macro = 16
	} else {
		writer.macro = options.Macro
	}

	if options.FPS == 0 {
		writer.fps = 25
	} else {
		writer.fps = options.FPS
	}

	if options.Quality == 0 {
		writer.quality = 0.5
	} else {
		writer.quality = math.Max(0, math.Min(options.Quality, 1))
	}

	if options.Codec == "" {
		if strings.HasSuffix(strings.ToLower(filename), ".wmv") {
			writer.codec = "msmpeg4"
		} else if strings.HasSuffix(strings.ToLower(filename), ".gif") {
			writer.codec = "gif"
		} else {
			writer.codec = "libx264"
		}
	} else {
		writer.codec = options.Codec
	}

	if options.Audio != "" {
		if !exists(options.Audio) {
			return nil, errors.New("Audio file " + options.Audio + " does not exist.")
		}

		audioData, err := ffprobe(options.Audio, "a")
		if err != nil {
			return nil, err
		} else if len(audioData) == 0 {
			return nil, errors.New("Given \"audio\" file " + options.Audio + " has no audio.")
		}

		writer.audio = options.Audio

		if options.AudioCodec == "" {
			writer.audioCodec = "aac"
		} else {
			writer.audioCodec = options.AudioCodec
		}
	}

	return &writer, nil
}

// Once the user calls Write() for the first time on a VideoWriter struct,
// the ffmpeg command which is used to write to the video file is started.
func initVideoWriter(writer *VideoWriter) error {
	// If user exits with Ctrl+C, stop ffmpeg process.
	writer.cleanup()
	// ffmpeg command to write to video file. Takes in bytes from Stdin and encodes them.
	command := []string{
		"-y", // overwrite output file if it exists.
		"-loglevel", "quiet",
		"-f", "rawvideo",
		"-vcodec", "rawvideo",
		"-s", fmt.Sprintf("%dx%d", writer.width, writer.height), // frame w x h.
		"-pix_fmt", "rgb24",
		"-r", fmt.Sprintf("%.02f", writer.fps), // frames per second.
		"-i", "-", // The input comes from stdin.
	}

	gif := strings.HasSuffix(strings.ToLower(writer.filename), ".gif")

	if writer.audio == "" || gif {
		command = append(command, "-an") // No audio.
	} else {
		command = append(
			command,
			"-i", writer.audio,
		)
	}

	command = append(
		command,
		"-vcodec", writer.codec,
		"-pix_fmt", "yuv420p", // Output is 8-but RGB, no alpha.
	)

	// Code from the imageio-ffmpeg project.
	// https://github.com/imageio/imageio-ffmpeg/blob/master/imageio_ffmpeg/_io.py#L399.
	// If bitrate not given, use a default.
	if writer.bitrate == 0 {
		if writer.codec == "libx264" {
			// Quality between 0 an 51. 51 is worst.
			command = append(command, "-crf", fmt.Sprintf("%d", int(writer.quality*51)))
		} else {
			// Quality between 1 and 31. 31 is worst.
			command = append(command, "-qscale:v", fmt.Sprintf("%d", int(writer.quality*30)+1))
		}
	} else {
		command = append(command, "-b:v", fmt.Sprintf("%d", writer.bitrate))
	}

	// For GIFs, add looping and delay parameters.
	if gif {
		command = append(
			command,
			"-loop", fmt.Sprintf("%d", writer.loop),
			"-final_delay", fmt.Sprintf("%d", writer.delay),
		)
	}

	// Code from the imageio-ffmpeg project:
	// https://github.com/imageio/imageio-ffmpeg/blob/master/imageio_ffmpeg/_io.py#L415.
	// Resizes the video frames to a size that works with most codecs.
	if writer.macro > 1 {
		if writer.width%writer.macro > 0 || writer.height%writer.macro > 0 {
			width := writer.width
			height := writer.height
			if writer.width%writer.macro > 0 {
				width += writer.macro - (writer.width % writer.macro)
			}
			if writer.height%writer.macro > 0 {
				height += writer.macro - (writer.height % writer.macro)
			}
			command = append(
				command,
				"-vf", fmt.Sprintf("scale=%d:%d", width, height),
			)
		}
	}

	// If audio was included, then specify video and audio channels.
	if writer.audio != "" && !gif {
		command = append(
			command,
			"-acodec", writer.audioCodec,
			"-map", "0:v:0",
			"-map", "1:a:0",
		)
	}

	command = append(command, writer.filename)
	cmd := exec.Command("ffmpeg", command...)
	writer.cmd = cmd

	pipe, err := cmd.StdinPipe()
	if err != nil {
		pipe.Close()
		return err
	}
	writer.pipe = &pipe
	if err := cmd.Start(); err != nil {
		cmd.Process.Kill()
		return err
	}

	return nil
}

// Writes the given frame to the video file.
func (writer *VideoWriter) Write(frame []byte) error {
	// If cmd is nil, video writing has not been set up.
	if writer.cmd == nil {
		if err := initVideoWriter(writer); err != nil {
			return err
		}
	}
	total := 0
	for total < len(frame) {
		n, err := (*writer.pipe).Write(frame[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

// Closes the pipe and stops the ffmpeg process.
func (writer *VideoWriter) Close() {
	if writer.pipe != nil {
		(*writer.pipe).Close()
	}
	if writer.cmd != nil {
		writer.cmd.Wait()
	}
}

// Stops the "cmd" process running when the user presses Ctrl+C.
// https://stackoverflow.com/questions/11268943/is-it-possible-to-capture-a-ctrlc-signal-and-run-a-cleanup-function-in-a-defe.
func (writer *VideoWriter) cleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if writer.pipe != nil {
			(*writer.pipe).Close()
		}
		if writer.cmd != nil {
			writer.cmd.Process.Kill()
		}
		os.Exit(1)
	}()
}
