package internal

import (
	"embed"
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/DrJosh9000/yarn/bytecode"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
	"golang.org/x/exp/shiny/materialdesign/colornames"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"strings"
	"sync"
	"unicode/utf8"
)

//go:embed gamedata
var gamedata embed.FS

var renderer *etxt.Renderer
var fontLib *etxt.FontLibrary

const dpi = 72
const fontSize = 12
const padding = 30

var fontColor = color.White

var (
	program     *bytecode.Program
	stringTable *yarn.StringTable
)

var (
	cathedralUI   *ebiten.Image
	barracks      *ebiten.Image
	greatHall     *ebiten.Image
	bedroom       *ebiten.Image
	berthilde     *ebiten.Image
	berthildeSide *ebiten.Image
	melusine      *ebiten.Image
	melusineSide  *ebiten.Image

	currBg   *ebiten.Image
	currChar *ebiten.Image
)

func init() {
	// load up UI and background images
	cathedralUI = loadImg("gamedata/ui-scaled.png")
	barracks = loadImg("gamedata/barracks.png")
	greatHall = loadImg("gamedata/hall_background.png")
	bedroom = loadImg("gamedata/bedroom.png")
	currBg = barracks

	berthilde = loadImg("gamedata/berthilde.png")
	berthildeSide = loadImg("gamedata/berthilde_side.png")

	melusine = loadImg("gamedata/melusine.png")
	melusineSide = loadImg("gamedata/melusineSide.png")

	melusine = loadImg("gamedata/melusine.png")

	// load up fonts
	fontLib = etxt.NewFontLibrary()
	_, _, err := fontLib.ParseEmbedDirFonts("gamedata/fonts", gamedata)
	if err != nil {
		panic(err)
	}
	var fonts []string
	_ = fontLib.EachFont(func(s string, font *etxt.Font) error {
		fonts = append(fonts, s)
		return nil
	})
	fmt.Println("fonts available:", strings.Join(fonts, ","))

	// set up text renderer
	renderer = etxt.NewStdRenderer()
	renderer.SetFont(fontLib.GetFont("DePixel Breit"))
	renderer.SetSizePx(fontSize)
	renderer.SetColor(fontColor)
	renderer.SetLineSpacing(1.15)

	program, stringTable, err = yarn.LoadFiles("internal/gamedata/yarn/Main.yarnc", "internal/gamedata/yarn/Main-Lines.csv", "en-US")
	if err != nil {
		panic(err)
	}
}

func loadImg(path string) *ebiten.Image {
	f, err := gamedata.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	im, _, err := image.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.NewImageFromImage(im)
}

type TextScene struct {
	handler     *DialogueHandler
	choices     []Choice
	lastSeq     int
	highlighted int

	music   *Player
	snoring *Player
}

type Choice struct {
	clickBounds image.Rectangle
	choice      int // integer option in the current dialogue settings.
}

func NewTextScene(aCtx *audio.Context) *TextScene {
	music, err := NewPlayer(aCtx, fluteLoop)
	if err != nil {
		panic(fmt.Errorf("audio context: %w", err))
	}
	snoring, err := NewPlayer(aCtx, snoring)
	if err != nil {
		panic(fmt.Errorf("audio context: %w", err))
	}
	result := &TextScene{
		handler: NewDialogueHandler(),
		music:   music,
		snoring: snoring,
	}
	result.handler.textScene = result
	return result
}

func (s *TextScene) Update() error {
	if err := s.music.update(); err != nil {
		panic(fmt.Errorf("music: %w", err))
	}
	s.music.volume128 = 10
	if err := s.snoring.update(); err != nil {
		panic(fmt.Errorf("snoring: %w", err))
	}
	s.snoring.volume128 = 128
	s.highlighted = -1
	posx, posy := ebiten.CursorPosition()
	for idx, choice := range s.choices {
		if image.Pt(posx, posy).In(choice.clickBounds) {
			s.highlighted = idx
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButton0) {
		posx, posy := ebiten.CursorPosition()
		for _, choice := range s.choices {
			if image.Pt(posx, posy).In(choice.clickBounds) {
				fmt.Println("sending choice:", choice.choice)
				s.handler.choice <- choice.choice // send choice back to dialogue handler
			}
		}
		func() {
			s.handler.click.Broadcast()
		}()
	}
	return nil
}

var DialogueBounds = image.Rect(146+5, 430+12, 146+517, 430+163)

func (s *TextScene) Draw(screen *ebiten.Image) {
	opts := ebiten.DrawImageOptions{}
	opts.GeoM.Translate(141, 0)
	screen.DrawImage(currBg, &opts) // draw current background centered

	if currChar != nil {
		opts.GeoM.Reset()
		opts.GeoM.Translate(float64(400-currChar.Bounds().Dx()/2), float64(300-currChar.Bounds().Dy()/2))
		screen.DrawImage(currChar, &opts)
	}

	screen.DrawImage(cathedralUI, nil) // draw foreground over everything.

	// draw all dialogue text
	renderer.SetTarget(screen)
	s.handler.mut.RLock()
	defer s.handler.mut.RUnlock()

	renderer.SetColor(colornames.White)
	feed := renderer.NewFeed(fixed.P(DialogueBounds.Min.X, DialogueBounds.Min.Y))

	DrawInBox(feed, s.handler.currNode.Prompt, DialogueBounds)
	feed.LineBreak()
	s.choices = make([]Choice, len(s.handler.currOpts))
	for i, choice := range s.handler.currOpts {
		str, err := stringTable.Render(choice.Line)
		if err != nil {
			panic(fmt.Errorf("rendering option: %w", err))
		}
		feed.LineBreak()
		if s.highlighted == i {
			renderer.SetColor(colornames.Cyan500)
		} else {
			renderer.SetColor(colornames.White)
		}
		s.choices[i].clickBounds = DrawInBox(feed, fmt.Sprintf("> %s", str), DialogueBounds)
		s.choices[i].choice = i

		s.choices[i].clickBounds.Max.X = DialogueBounds.Max.X
	}
}

type Node struct {
	Prompt  string
	Choices []string
}

func (s *TextScene) Layout(w, h int) (int, int) {
	return 800, 600
}

type DialogueHandler struct {
	vm        *yarn.VirtualMachine
	mut       sync.RWMutex
	seq       int // ever-increasing sequence number
	currNode  Node
	choice    chan int
	textScene *TextScene

	click    *sync.Cond
	clickMut *sync.Mutex

	currOpts []yarn.Option
}

func NewDialogueHandler() *DialogueHandler {
	result := &DialogueHandler{
		choice:   make(chan int),
		clickMut: &sync.Mutex{},
	}
	result.click = sync.NewCond(result.clickMut)
	result.vm = &yarn.VirtualMachine{
		Program: program,
		Handler: result,
		Vars:    make(yarn.MapVariableStorage),
	}
	go func() {
		fmt.Println("running dialogue handler")
		if err := result.vm.Run("Start"); err != nil {
			panic(fmt.Errorf("error from Yarn VM: %w", err))
		}
	}()
	return result
}

func (h *DialogueHandler) NodeStart(nodeName string) error {
	fmt.Println("dialogue handler: start", nodeName)
	return nil
}

func (h *DialogueHandler) PrepareForLines(lineIDs []string) error {
	return nil
}

func (h *DialogueHandler) Line(line yarn.Line) error {
	str, err := stringTable.Render(line)
	if err != nil {
		return err
	}
	h.mut.Lock()
	func() {
		defer h.mut.Unlock()
		if h.currNode.Prompt != "" {
			h.currNode.Prompt += "\n"
		}
		h.currNode.Prompt += str.String()
	}()
	return nil
}

func (h *DialogueHandler) Update() error {
	fmt.Println("dialogue handler: update")
	return nil
}

func (h *DialogueHandler) Options(opts []yarn.Option) (int, error) {
	var choices []string
	for _, opt := range opts {
		str, err := stringTable.Render(opt.Line)
		if err != nil {
			return 0, err
		}
		choices = append(choices, str.String())
	}
	h.mut.Lock()
	func() {
		defer h.mut.Unlock()
		h.currNode.Choices = choices
		h.currOpts = opts
		h.seq++
	}()
	fmt.Println("dialogue handler: blocking on selection")
	choice := <-h.choice
	fmt.Printf("dialogue handler: selected choice [%d]\n", choice)
	h.currNode.Prompt = "" // HACK!
	return choice, nil
}

func (h *DialogueHandler) NodeComplete(name string) error {
	fmt.Println("node complete")
	return nil
}

func (h *DialogueHandler) DialogueComplete() error {
	fmt.Println("dialogue complete")
	go func() { // MAJOR END OF GAME HAX!!
		for ch := range h.choice {
			fmt.Println("chosen", ch)
		}
	}()
	return nil
}

func (h *DialogueHandler) Command(cmd string) error {
	tokens := strings.Split(cmd, " ")
	switch tokens[0] {
	case "background":
		return h.background(tokens[1])
	case "char":
		return h.character(tokens[1])
	case "wait":
		return h.wait()
	case "loopStart":
		return h.loopStart(tokens[1])
	case "loopStop":
		return h.loopStop(tokens[1])
	default:
		return fmt.Errorf("unknown commmand '%s'", tokens[0])
	}
}

func (h *DialogueHandler) loopStart(name string) error {
	switch name {
	case "snoring":
		h.textScene.snoring.Start()
	case "music":
		h.textScene.music.Start()
	}
	return nil
}
func (h *DialogueHandler) loopStop(name string) error {
	switch name {
	case "snoring":
		h.textScene.snoring.Stop()
	case "music":
		h.textScene.music.Stop()
	}
	return nil
}

func (h *DialogueHandler) background(img string) error {
	switch img {
	case "great_hall":
		currBg = greatHall
	case "barracks":
		currBg = barracks
	case "bedroom":
		currBg = bedroom
	default:
		return fmt.Errorf("unknown background: '%s'", img)
	}
	return nil
}

func (h *DialogueHandler) character(name string) error {
	switch name {
	case "berthilde":
		setCharacter(berthilde)
	case "berthildeSide":
		setCharacter(berthildeSide)
	case "melusine":
		setCharacter(melusine)
	case "melusineSide":
		setCharacter(melusineSide)
	case "none":
		unsetCharacter()
	default:
		return fmt.Errorf("unknown character: '%s'", name)
	}
	return nil
}

func (h *DialogueHandler) wait() error {
	h.currOpts = nil
	h.clickMut.Lock()
	defer h.clickMut.Unlock()
	h.click.Wait() // block the current thread
	// clear the dialogue
	h.currNode.Prompt = "" // MOAR HACKS
	return nil
}

func setCharacter(img *ebiten.Image) {
	currChar = img
}

func unsetCharacter() {
	currChar = nil
}

// DrawInBox draws the provided string staying within the provided bounds. The rectangle taken up by the text written is
// returned.
func DrawInBox(feed *etxt.Feed, text string, bounds image.Rectangle) image.Rectangle {
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
	used := image.Rectangle{}
	used.Min.X = feed.Position.X.Ceil()
	used.Min.Y = feed.Position.Y.Floor()
	used.Max = used.Min

	// create Feed and iterate each rune / word
	if feed == nil {
		feed = renderer.NewFeed(fixed.P(bounds.Min.X, bounds.Min.Y))
	}
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
				return used
			}
			used.Max.X = max(used.Max.X, (feed.Position.X + width).Ceil())
			used.Max.Y = max(used.Max.Y, feed.Position.Y.Floor()+14)

			// draw the word and increase index
			for _, codePoint := range word {
				feed.Draw(codePoint) // you may want to cut this earlier if the word is too long
			}
			index += len(word)
		}
	}
	return used
}
