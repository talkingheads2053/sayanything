package say

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/tosone/minimp3"
)

// Player plays audio on the default output device. oto allows one context per
// process, so the device opens once and later audio is converted to its format.
type Player struct {
	mu           sync.Mutex
	ctx          *oto.Context
	sampleRate   int
	channelCount int
	format       string
}

func NewPlayer(format string) *Player {
	return &Player{format: format}
}

func (p *Player) Close() {}

func (p *Player) Say(b []byte) error {
	return p.Play(context.Background(), b, p.format, nil)
}

// Play decodes b as mp3 or wav and blocks until it finishes or ctx is canceled.
// started is called when playback begins.
func (p *Player) Play(ctx context.Context, b []byte, format string, started func()) error {
	if len(b) == 0 {
		return nil
	}

	var pcm []int16
	var rate, chans int
	var err error
	switch format {
	case "mp3":
		pcm, rate, chans, err = decodeMP3(b)
	case "wav":
		pcm, rate, chans, err = decodeWAV(b)
	default:
		return errors.New("unsupported format")
	}
	if err != nil {
		return err
	}

	octx, err := p.open(rate, chans)
	if err != nil {
		return err
	}
	pcm = Convert(pcm, rate, chans, p.sampleRate, p.channelCount)

	player := octx.NewPlayer(bytes.NewReader(int16Bytes(pcm)))
	defer player.Close()
	player.Play()
	if started != nil {
		started()
	}
	for player.IsPlaying() {
		select {
		case <-ctx.Done():
			player.Pause()
			return ctx.Err()
		case <-time.After(time.Millisecond):
		}
	}
	return nil
}

// open creates the oto context on first use and returns it.
func (p *Player) open(sampleRate, channelCount int) (*oto.Context, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.ctx != nil {
		return p.ctx, nil
	}
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       oto.FormatSignedInt16LE,
	})
	if err != nil {
		return nil, err
	}
	<-ready
	p.ctx = ctx
	p.sampleRate = sampleRate
	p.channelCount = channelCount
	return ctx, nil
}

// Drain waits for the hardware audio buffer to finish outputting the last
// played audio. Call this once after the final phrase when not running as a
// server, so the process does not exit before the hardware buffer is empty.
func (p *Player) Drain() {
	time.Sleep(100 * time.Millisecond)
}

// Convert resamples interleaved pcm to another rate with linear interpolation
// and mixes or copies channels to match.
func Convert(pcm []int16, rate, chans, toRate, toChans int) []int16 {
	if chans < 1 || toChans < 1 || rate < 1 || toRate < 1 {
		return pcm
	}
	if rate == toRate && chans == toChans {
		return pcm
	}

	frames := len(pcm) / chans
	mixed := make([]float64, frames*toChans)
	for f := 0; f < frames; f++ {
		src := pcm[f*chans : f*chans+chans]
		for c := 0; c < toChans; c++ {
			if chans == toChans {
				mixed[f*toChans+c] = float64(src[c])
				continue
			}
			var sum float64
			for _, s := range src {
				sum += float64(s)
			}
			mixed[f*toChans+c] = sum / float64(chans)
		}
	}
	if rate == toRate || frames == 0 {
		return toInt16(mixed)
	}

	outFrames := int(int64(frames) * int64(toRate) / int64(rate))
	out := make([]float64, outFrames*toChans)
	step := float64(rate) / float64(toRate)
	for f := 0; f < outFrames; f++ {
		pos := float64(f) * step
		i := int(pos)
		frac := pos - float64(i)
		j := min(i+1, frames-1)
		for c := 0; c < toChans; c++ {
			a, b := mixed[i*toChans+c], mixed[j*toChans+c]
			out[f*toChans+c] = a + (b-a)*frac
		}
	}
	return toInt16(out)
}

func toInt16(in []float64) []int16 {
	out := make([]int16, len(in))
	for i, v := range in {
		out[i] = int16(max(-32768, min(32767, v)))
	}
	return out
}

func int16Bytes(pcm []int16) []byte {
	out := make([]byte, len(pcm)*2)
	for i, s := range pcm {
		out[i*2] = byte(s)
		out[i*2+1] = byte(s >> 8)
	}
	return out
}

func decodeMP3(b []byte) ([]int16, int, int, error) {
	dec, data, err := minimp3.DecodeFull(b)
	if err != nil {
		return nil, 0, 0, err
	}
	pcm := make([]int16, len(data)/2)
	for i := range pcm {
		pcm[i] = int16(uint16(data[i*2]) | uint16(data[i*2+1])<<8)
	}
	return pcm, dec.SampleRate, dec.Channels, nil
}

func decodeWAV(b []byte) ([]int16, int, int, error) {
	dec := wav.NewDecoder(bytes.NewReader(b))
	if dec == nil {
		return nil, 0, 0, errors.New("failed to create wav decoder")
	}
	dec.ReadInfo()
	if dec.SampleRate == 0 {
		return nil, 0, 0, errors.New("failed to read wav info")
	}

	pcmBuf := &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  int(dec.SampleRate),
			NumChannels: int(dec.NumChans),
		},
		Data: make([]int, 4096),
	}

	// oto v3's FormatUnsignedInt8 has a Go byte-arithmetic bug (v8-(1<<7)
	// wraps unsigned, producing inverted output). Always decode to signed 16-bit.
	var pcm []int16
	for {
		n, err := dec.PCMBuffer(pcmBuf)
		if n == 0 {
			break
		}
		for _, v := range pcmBuf.Data[:n] {
			switch dec.BitDepth {
			case 8:
				pcm = append(pcm, int16((v-128)*256))
			case 24:
				pcm = append(pcm, int16(v>>8))
			case 32:
				pcm = append(pcm, int16(v>>16))
			default:
				pcm = append(pcm, int16(v))
			}
		}
		if err != nil {
			break
		}
	}
	return pcm, int(dec.SampleRate), int(dec.NumChans), nil
}
