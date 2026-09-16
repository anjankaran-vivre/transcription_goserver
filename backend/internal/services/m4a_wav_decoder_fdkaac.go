//go:build cgo && !windows

package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	fdkaac "github.com/IzumiSy/go-fdkaac"
	"github.com/sagartechversant/go-m4a-wav-decode/mp4audio"
)

func convertM4AToWAVInGo(audioFilePath string) ([]byte, error) {
	file, err := os.Open(audioFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	mp4Audio, err := mp4audio.New(file)
	if err != nil {
		return nil, err
	}

	ascDescriptor, err := mp4Audio.ASCDescriptor()
	if err != nil {
		return nil, err
	}
	if len(ascDescriptor.Data) == 0 {
		return nil, fmt.Errorf("empty AAC decoder config")
	}

	decoder := fdkaac.NewAacDecoder()
	if err := decoder.InitRaw(ascDescriptor.Data); err != nil {
		return nil, err
	}
	defer decoder.Close()

	frameIterator, err := mp4Audio.Frames()
	if err != nil {
		return nil, err
	}

	var pcm bytes.Buffer
	for {
		nextFrame := frameIterator.Next()
		if nextFrame == nil {
			break
		}

		frame := make([]byte, nextFrame.Size)
		_, err := io.ReadFull(io.NewSectionReader(file, int64(nextFrame.Offset), int64(nextFrame.Size)), frame)
		if err != nil {
			return nil, err
		}

		if err := decoder.Decode(frame, &pcm); err != nil {
			return nil, err
		}
	}
	if err := decoder.Decode(nil, &pcm); err != nil {
		return nil, err
	}
	if pcm.Len() == 0 {
		return nil, fmt.Errorf("decoded PCM was empty")
	}

	sampleRate := decoder.SampleRate()
	channels := decoder.NumChannels()
	bitsPerSample := decoder.SampleBits()
	if sampleRate <= 0 || channels <= 0 || bitsPerSample <= 0 {
		return nil, fmt.Errorf("invalid decoded audio metadata: sampleRate=%d channels=%d bits=%d", sampleRate, channels, bitsPerSample)
	}

	return writePCMToWAV(pcm.Bytes(), sampleRate, channels, bitsPerSample)
}

func writePCMToWAV(pcm []byte, sampleRate, channels, bitsPerSample int) ([]byte, error) {
	if channels <= 0 || sampleRate <= 0 || bitsPerSample <= 0 {
		return nil, fmt.Errorf("invalid wav metadata")
	}
	if len(pcm) > 0xFFFFFFFF-36 {
		return nil, fmt.Errorf("wav data too large")
	}

	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	var out bytes.Buffer
	out.WriteString("RIFF")
	binary.Write(&out, binary.LittleEndian, uint32(36+len(pcm)))
	out.WriteString("WAVE")
	out.WriteString("fmt ")
	binary.Write(&out, binary.LittleEndian, uint32(16))
	binary.Write(&out, binary.LittleEndian, uint16(1))
	binary.Write(&out, binary.LittleEndian, uint16(channels))
	binary.Write(&out, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&out, binary.LittleEndian, uint32(byteRate))
	binary.Write(&out, binary.LittleEndian, uint16(blockAlign))
	binary.Write(&out, binary.LittleEndian, uint16(bitsPerSample))
	out.WriteString("data")
	binary.Write(&out, binary.LittleEndian, uint32(len(pcm)))
	out.Write(pcm)

	return out.Bytes(), nil
}
