package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"image/color"
	"log"
	"time"
)

// N.B.: do NOT create a util package; create a util file and leave common helpers there.
// See https://www.adam-bien.com/roller/abien/entry/util_packages_are_evil for more details

type float interface {
	~float64 | ~float32
}

type number interface {
	~float64 | ~int
}

func max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func min[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func orZero[T float](x T) T {
	if -1e-2 < x && x < 1e-2 {
		return 0
	}
	return x
}

func timeit(operation string, f func()) {
	start := time.Now()
	f()
	duration := time.Now().Sub(start)
	log.Printf("completed %s in %s\n", operation, duration.Round(1*time.Microsecond))
}

func placeholderImage(w, h int, baseColor color.Color) *ebiten.Image {
	var halfSat colorm.ColorM
	halfSat.Scale(0.5, 0.5, 0.5, 1.0)

	result := ebiten.NewImage(w, h)
	// Draw a fake sprite with boundary
	for x := 0; x < h; x++ {
		for y := 0; y < h; y++ {
			c := baseColor
			if x%(w-1) == 0 || y%(h-1) == 0 {
				c = halfSat.Apply(baseColor)
			}
			result.Set(x, y, c)
		}
	}
	return result
}
