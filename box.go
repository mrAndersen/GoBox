package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math/rand"
	"time"
)

const sw int32 = 1910
const sh int32 = 820

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

type Object struct {
	id int64

	speed        float64
	acceleration float64

	frameTimeMs   float64
	lastUpdatedMs float64
	lastDelta     float64

	x int32
	y int32
	w int32
	h int32

	rect  sdl.Rect
	start time.Time

	color [4]uint8
}

type Mew struct {
	lastObjectId int64
}

func (m *Mew) object(x int32, y int32, w int32, h int32) *Object {
	o := Object{x: x, y: y, w: w, h: h}

	o.start = time.Now()
	o.speed = 1.5

	o.lastUpdatedMs = float64(time.Now().UnixNano() / 1000)
	o.frameTimeMs = 0

	o.lastDelta = 0
	o.color = peekColor()

	o.id = m.lastObjectId
	m.lastObjectId++

	return &o
}

func (o *Object) render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(o.color[0], o.color[1], o.color[2], o.color[3])
	renderer.FillRect(&o.rect)
}

func (o *Object) update([]*Object) {
	life := time.Now().Sub(o.start).Seconds()
	o.frameTimeMs = float64(time.Now().UnixNano()/1000) - o.lastUpdatedMs

	deltaPx := o.speed * o.frameTimeMs / 10000
	o.lastDelta += deltaPx

	if o.lastDelta >= 1 {
		o.y += 1
		o.lastDelta = 0
	}

	o.speed = o.speed + (o.speed * life / 10)

	if o.y+o.h >= sh {
		o.speed = 0
	}

	o.rect.W = int32(o.w)
	o.rect.H = int32(o.h)
	o.rect.X = int32(o.x)
	o.rect.Y = int32(o.y)

	o.lastUpdatedMs = float64(time.Now().UnixNano() / 1000)
}

func (o *Object) isDead() bool {
	lifeMs := time.Now().Sub(o.start).Seconds() * 1000

	if lifeMs >= 50000 {
		return true
	}

	return false
}

func createWindow(width int32, height int32) *sdl.Window {
	handleError(sdl.Init(sdl.INIT_EVERYTHING))

	displayMode, err := sdl.GetCurrentDisplayMode(0)
	handleError(err)

	top := displayMode.H/2 - height/2
	left := displayMode.W/2 - width/2

	window, err := sdl.CreateWindow("Boxes", left, top, width, height, sdl.WINDOW_SHOWN)
	handleError(err)

	return window
}

func peekColor() [4]uint8 {
	c := [4]uint8{uint8(rand.Intn(255)), uint8(rand.Int31n(255)), uint8(rand.Int31n(255)), uint8(rand.Int31n(255))}
	return c
}

func main() {
	window := createWindow(sw, sh)

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	handleError(err)

	running := true
	mousePressed := false

	var frames int64 = 0
	var start = time.Now()

	mew := Mew{lastObjectId: 1}

	container := make([]*Object, 0)
	delContainer := make([]int, 0)
	lastSpawn := time.Now().UnixNano()

	for running {
		event := sdl.PollEvent()

		//clear
		renderer.Clear()
		renderer.SetDrawColor(102, 113, 132, 200)
		renderer.FillRect(&sdl.Rect{0, 0, sw, sh})

		for i := range delContainer {
			container = append(container[:i], container[i+1:]...)
		}

		delContainer = make([]int, 0)

		if event != nil {
			eType := event.GetType()

			switch eType {
			case sdl.QUIT:
				running = false

			case sdl.MOUSEBUTTONUP:
				mousePressed = false

			case sdl.MOUSEBUTTONDOWN:
				mousePressed = true
			}
		}

		if mousePressed && time.Now().UnixNano()-lastSpawn >= 500000000 {
			mx, my, _ := sdl.GetMouseState()
			size := rand.Int31n(50) + 20

			x := mx - size/2
			y := my - size/2

			o := mew.object(x, y, size, size)
			container = append(container, o)

			lastSpawn = time.Now().UnixNano()
		}

		if time.Now().Sub(start).Seconds() >= 1 {
			window.SetTitle(fmt.Sprintf("Boxes, fps=%d, objects=%d", frames, len(container)))

			start = time.Now()
			frames = 0
		}

		for k, o := range container {
			if o.isDead() {
				delContainer = append(delContainer, k)
			} else {
				o.update(container)
				o.render(renderer)
			}
		}

		renderer.Present()
		frames++
	}
}
