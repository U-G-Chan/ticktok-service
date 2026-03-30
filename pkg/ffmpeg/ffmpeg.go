package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"ticktok-service/pkg/config"
)

// ExtractCover extracts the first frame (or at specific second) of a video as a cover image.
// videoURL can be a local file path or an HTTP/HTTPS URL.
// outputPath is the local file path where the image will be saved.
func ExtractCover(videoURL string, outputPath string) error {
	ffmpegPath := config.Config.Media.FfmpegPath
	if ffmpegPath == "" {
		ffmpegPath = "ffmpeg" // Default to global command
	}

	cmd := exec.Command(ffmpegPath,
		"-i", videoURL,
		"-ss", "00:00:01",
		"-vframes", "1",
		"-y", // overwrite output file if it exists
		outputPath,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg execution failed: %v, stderr: %s", err, stderr.String())
	}

	return nil
}
