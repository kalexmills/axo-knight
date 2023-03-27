package internal

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"image"
	"io"
	"time"
)

const (
	sampleRate = 48000
)

// Player represents the current audio state.
type Player struct {
	game           *Game
	audioContext   *audio.Context
	audioPlayer    *audio.Player
	current        time.Duration
	total          time.Duration
	totalSampleLen time.Duration
	seBytes        []byte
	volume128      int

	fadeInTime *time.Duration
	fadeStart  time.Time

	playButtonPosition  image.Point
	alertButtonPosition image.Point
}

const fluteLoop = "anotherfluteloopwithsometamborine.mp3"
const snoring = "snoring.mp3"

func NewPlayer(audioContext *audio.Context, file string) (*Player, error) {
	type audioStream interface {
		io.ReadSeeker
		Length() int64
	}

	const bytesPerSample = 4 // TODO: This should be defined in audio package

	var s audioStream

	f, err := gamedata.Open("gamedata/music/" + file)
	if err != nil {
		panic(err)
	}

	s, err = mp3.DecodeWithoutResampling(f)
	if err != nil {
		return nil, err
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return nil, err
	}
	fadein := 5 * time.Second
	player := &Player{
		audioContext:   audioContext,
		audioPlayer:    p,
		totalSampleLen: time.Second * time.Duration(s.Length()) / bytesPerSample / sampleRate,
		volume128:      64,
		fadeInTime:     &fadein,
		fadeStart:      time.Now(),
	}
	if player.totalSampleLen == 0 {
		player.totalSampleLen = 1
	}

	return player, nil
}

func (p *Player) Close() error {
	return p.audioPlayer.Close()
}

func (p *Player) Stop() {
	p.audioPlayer.Pause()
}

func (p *Player) Start() {
	p.audioPlayer.Play()
}

func (p *Player) update() error {
	if p.audioPlayer.IsPlaying() {
		p.current = p.audioPlayer.Current()
	} else {
		p.audioPlayer.Play()
	}
	if p.totalSampleLen-p.current <= 19*time.Millisecond { // TODO: this is a hack.
		p.total += p.current
		if err := p.audioPlayer.Rewind(); err != nil {
			panic(fmt.Errorf("audio player panicked on rewind: %w", err))
		}
		p.audioPlayer.Play()
	}
	p.updateVolumeIfNeeded()

	return nil
}

func (p *Player) totalTime() time.Duration {
	return p.total + p.current
}

func (p *Player) updateVolumeIfNeeded() {
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		p.volume128--
	}
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		p.volume128++
	}
	if p.volume128 < 0 {
		p.volume128 = 0
	}
	if 128 < p.volume128 {
		p.volume128 = 128
	}
	if p.fadeInTime != nil { // HACKS!
		dt := float64(time.Now().Sub(p.fadeStart)) / float64(*p.fadeInTime)
		if dt > 1 {
			dt = 1
			p.fadeStart = time.Time{}
			p.fadeInTime = nil
		}
		p.audioPlayer.SetVolume(dt * float64(p.volume128) / 128)
	} else {
		p.audioPlayer.SetVolume(float64(p.volume128) / 128)
	}
}
