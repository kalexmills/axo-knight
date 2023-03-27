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
	game         *Game
	audioContext *audio.Context
	audioPlayer  *audio.Player
	current      time.Duration
	total        time.Duration
	seBytes      []byte
	volume128    int

	playButtonPosition  image.Point
	alertButtonPosition image.Point
}

func NewPlayer(audioContext *audio.Context) (*Player, error) {
	type audioStream interface {
		io.ReadSeeker
		Length() int64
	}

	const bytesPerSample = 4 // TODO: This should be defined in audio package

	var s audioStream

	f, err := gamedata.Open("gamedata/music/fluteloop.mp3")
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
	player := &Player{
		audioContext: audioContext,
		audioPlayer:  p,
		total:        time.Second * time.Duration(s.Length()) / bytesPerSample / sampleRate,
		volume128:    128,
	}
	if player.total == 0 {
		player.total = 1
	}

	player.audioPlayer.Play()
	return player, nil
}

func (p *Player) Close() error {
	return p.audioPlayer.Close()
}

func (p *Player) update() error {
	if p.audioPlayer.IsPlaying() {
		p.current = p.audioPlayer.Current()
	}
	if p.total-p.current <= 19*time.Millisecond { // TODO: this is a total hack.
		if err := p.audioPlayer.Rewind(); err != nil {
			panic(fmt.Errorf("audio player panicked on rewind: %w", err))
		}
		p.audioPlayer.Play()
	}
	p.updateVolumeIfNeeded()

	return nil
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
	p.audioPlayer.SetVolume(float64(p.volume128) / 128)
}
