package internal

import (
	"embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"unicode/utf8"
)

//
//// Type alias to create an unexported alias of etxt.Renderer.
//// This is quite irrelevant for this example, but it's useful in
//// practical scenarios to avoid leaking a public internal field.
//type renderer = etxt.Renderer
//
//// Wrapper type for etxt.Renderer. Since this type embeds etxt.Renderer
//// it will preserve all its methods, and we can additionally add our own
//// new DrawInBox() method.
//type TextBoxRenderer struct{ renderer }
//
//// The new method for TextBoxRenderer. It draws the given text within the
//// given bounds, performing basic line wrapping on space " " characters.
//// This is only meant as a reference: this method doesn't split on "-",
//// very long words will overflow the box when a single word is longer
//// than the width of the box, \r\n will be considered two line breaks
//// instead of one, etc. In many practical scenarios you will want to
//// further customize the behavior of this function. For more complex
//// examples of Feed usages, see examples/ebiten/typewriter, which also
//// has a typewriter effect, multiple colors, bold, italics and more.
//// Otherwise, if you only needed really basic line wrapping, feel free
//// to copy this function and use it directly. If you don't want a custom
//// TextBoxRenderer type, it's trivial to adapt the function to receive
//// a standard *etxt.Renderer as an argument instead.
////
//// Notice that this function relies on the renderer's alignment being
//// (etxt.Top, etxt.Left).
//func (r *TextBoxRenderer) DrawInBox(text string, bounds image.Rectangle) {
//	// helper function
//	var getNextWord = func(str string, index int) string {
//		start := index
//		for index < len(str) {
//			codePoint, size := utf8.DecodeRuneInString(str[index:])
//			if codePoint <= ' ' {
//				return str[start:index]
//			}
//			index += size
//		}
//		return str[start:index]
//	}
//
//	// create Feed and iterate each rune / word
//	feed := r.renderer.NewFeed(fixed.P(bounds.Min.X, bounds.Min.Y))
//	index := 0
//	for index < len(text) {
//		switch text[index] {
//		case ' ': // handle spaces with Advance() instead of Draw()
//			feed.Advance(' ')
//			index += 1
//		case '\n', '\r': // \r\n line breaks *not* handled as single line breaks
//			feed.LineBreak()
//			index += 1
//		default:
//			// get next word and measure it to see if it fits
//			word := getNextWord(text, index)
//			width := r.renderer.SelectionRect(word).Width
//			if (feed.Position.X + width).Ceil() > bounds.Max.X {
//				feed.LineBreak() // didn't fit, jump to next line before drawing
//			}
//
//			// abort if we are going beyond the vertical working area
//			if feed.Position.Y.Floor() >= bounds.Max.Y {
//				return
//			}
//
//			// draw the word and increase index
//			for _, codePoint := range word {
//				feed.Draw(codePoint) // you may want to cut this earlier if the word is too long
//			}
//			index += len(word)
//		}
//	}
//}

//go:embed gamedata
var gamedata embed.FS

var renderer *etxt.Renderer
var fontLib *etxt.FontLibrary

const dpi = 72
const fontSize = 12
const padding = 30

var fontColor = color.White

func init() {
	fontLib = etxt.NewFontLibrary()
	_, _, err := fontLib.ParseEmbedDirFonts("gamedata/fonts", gamedata)
	if err != nil {
		panic(err)
	}
	renderer = etxt.NewStdRenderer()
	renderer.SetFont(fontLib.GetFont("Press Start Regular"))
	renderer.SetSizePx(fontSize)
	renderer.SetColor(fontColor)
}

type TextScene struct {
}

func NewTextScene() *TextScene {
	return &TextScene{}
}

func (s *TextScene) Update() error {
	return nil
}

func (s *TextScene) Draw(screen *ebiten.Image) {
	s.LayoutNode(screen, Node{
		Prompt: "There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc.",
	})
}

type Node struct {
	Prompt  string
	Choices []string
}

func (s *TextScene) LayoutNode(screen *ebiten.Image, n Node) {
	renderer.SetTarget(screen)
	DrawInBox(n.Prompt, image.Rect(500, 12+padding, 800-padding, 600))
}

func DrawInBox(text string, bounds image.Rectangle) {
	// helper function
	var getNextWord = func(str string, index int) string {
		start := index
		for index < len(str) {
			codePoint, size := utf8.DecodeRuneInString(str[index:])
			if codePoint <= ' ' {
				return str[start:index]
			}
			index += size
		}
		return str[start:index]
	}

	// create Feed and iterate each rune / word
	feed := renderer.NewFeed(fixed.P(bounds.Min.X, bounds.Min.Y))
	index := 0
	for index < len(text) {
		switch text[index] {
		case ' ': // handle spaces with Advance() instead of Draw()
			feed.Advance(' ')
			index += 1
		case '\n', '\r': // \r\n line breaks *not* handled as single line breaks
			feed.LineBreak()
			index += 1
		default:
			// get next word and measure it to see if it fits
			word := getNextWord(text, index)
			width := renderer.SelectionRect(word).Width
			if (feed.Position.X + width).Ceil() > bounds.Max.X {
				feed.LineBreak() // didn't fit, jump to next line before drawing
			}

			// abort if we are going beyond the vertical working area
			if feed.Position.Y.Floor() >= bounds.Max.Y {
				return
			}

			// draw the word and increase index
			for _, codePoint := range word {
				feed.Draw(codePoint) // you may want to cut this earlier if the word is too long
			}
			index += len(word)
		}
	}
}

func (s *TextScene) Layout(w, h int) (int, int) {
	return 800, 600
}
