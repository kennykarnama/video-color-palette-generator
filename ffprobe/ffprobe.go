package ffprobe

import (
	"context"
	"strconv"
	"strings"
	"time"

	goffprobe "gopkg.in/vansante/go-ffprobe.v2"
)

type Ffprobe interface {
	GetBitrate() string
	GetDuration() time.Duration
	GetSize() int64
	GetVideoHeight() int
	GetVideoWidth() int
	GetVideoFps() int
}

type ffprobe struct {
	data *goffprobe.ProbeData
}

func NewFfprobe(mediaPath string) (*ffprobe, error) {
	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFn()

	probeData, err := goffprobe.ProbeURL(ctx, mediaPath)
	if err != nil {
		return nil, err
	}
	return &ffprobe{data: probeData}, nil
}

func (f *ffprobe) GetBitrate() string {
	return f.data.Format.BitRate
}

func (f *ffprobe) GetDuration() time.Duration {
	return f.data.Format.Duration()
}

func (f *ffprobe) GetSize() int64 {
	size, _ := strconv.ParseInt(f.data.Format.Size, 10, 64)
	return size
}

func (f *ffprobe) GetVideoHeight() int {
	videoStream := f.data.FirstVideoStream()
	return videoStream.Height
}

func (f *ffprobe) GetVideoWidth() int {
	videoStream := f.data.FirstVideoStream()
	return videoStream.Width
}

func (f *ffprobe) GetVideoFps() float64 {
	videoStream := f.data.FirstVideoStream()
	fpsString := videoStream.RFrameRate
	fps := float64(0)
	if strings.Contains(fpsString, "/") {
		fpsSlice := strings.Split(fpsString, "/")
		enum, _ := strconv.Atoi(fpsSlice[0])
		denum, _ := strconv.Atoi(fpsSlice[1])
		fps = float64(enum) / float64(denum)
	} else {
		fps, _ = strconv.ParseFloat(fpsString, 64)
	}
	return fps
}

func (f *ffprobe) HasAudio() bool {
	return f.data.FirstAudioStream() != nil
}

func (f *ffprobe) GetAudioVideoCodecs() Codecs {
	var codecs Codecs
	for _, stream := range f.data.Streams {
		if strings.EqualFold(stream.CodecType, "audio") || strings.EqualFold(stream.CodecType, "video") {
			codecs = append(codecs, &Codec{
				Name:    stream.CodecName,
				Tag:     stream.CodecTag,
				Profile: stream.Profile,
			})
		}
	}
	return codecs
}
