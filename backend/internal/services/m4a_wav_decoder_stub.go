//go:build !cgo || windows

package services

import "fmt"

func convertM4AToWAVInGo(audioFilePath string) ([]byte, error) {
	return nil, fmt.Errorf("go-fdkaac M4A decoder is unavailable on this build; configure AUDIO_CONVERTER_PATH to an ffmpeg-compatible binary")
}
