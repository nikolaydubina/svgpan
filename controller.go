package svgpan

import (
	"fmt"
	"math"
	"strings"
	"syscall/js"

	"gonum.org/v1/gonum/mat"
)

type state string

const (
	drag      state = "drag"
	notActive state = ""
)

// SVGPanZoomer does zooming and panning of SVG using JS events.
// This is port of https://github.com/aleofreddi/svgpan
// Which is also as of 2021-09-19 inlined in https://github.com/google/pprof/blob/master/third_party/svgpan/svgpan.go
type SVGPanZoomer struct {
	svgID               string
	rootID              string
	zoomScale           float64
	state               state
	transform           *mat.Dense
	transformBeforeDrag *mat.Dense
	origin              *mat.Dense
}

func NewSVGPanZoomer(
	svgID string,
	rootID string,
	zoomScale float64,
) *SVGPanZoomer {
	return &SVGPanZoomer{
		svgID:     svgID,
		rootID:    rootID,
		zoomScale: zoomScale,
		transform: identity(),
		state:     notActive,
	}
}

func (p *SVGPanZoomer) SetupHandlers() {
	container := js.Global().Get("document").Call("getElementById", p.svgID)
	container.Call("addEventListener", "mouseup", js.FuncOf(p.handleMouseUp))
	container.Call("addEventListener", "mousedown", js.FuncOf(p.handleMouseDown))
	container.Call("addEventListener", "mousemove", js.FuncOf(p.handleMouseMove))
	container.Call("addEventListener", "mousepout", js.FuncOf(p.handleMouseUp))

	userAgent := js.Global().Get("navigator").Get("userAgent").String()
	if strings.Contains(strings.ToLower(userAgent), "webkit") {
		// Chrome/Safari
		container.Call("addEventListener", "mousewheel", js.FuncOf(p.handleMouseWheel), false)
	} else {
		// Firefox/Other
		container.Call("addEventListener", "DOMMouseScroll", js.FuncOf(p.handleMouseWheel), false)
	}
}

func getEventPoint(event js.Value) *mat.Dense {
	return mat.NewDense(3, 1, []float64{
		event.Get("clientX").Float(),
		event.Get("clientY").Float(),
		1,
	})
}

// setRootTranslation updates translation matrix of root svg element
func (p *SVGPanZoomer) SetRootTranslation() {
	a := p.transform.At(0, 0)
	b := p.transform.At(1, 0)
	c := p.transform.At(0, 1)
	d := p.transform.At(1, 1)
	e := p.transform.At(0, 2)
	f := p.transform.At(1, 2)

	s := fmt.Sprintf("matrix(%f,%f,%f,%f,%f,%f)", a, b, c, d, e, f)
	js.Global().Get("document").Call("getElementById", p.rootID).Call("setAttribute", "transform", s)
}

// will make zoom at currently pointing at mouse
func (p *SVGPanZoomer) handleMouseWheel(this js.Value, args []js.Value) interface{} {
	event := args[0]
	delta := 0.0
	if event.Get("wheelDelta").Truthy() {
		// Chrome/Safari
		delta = event.Get("wheelDelta").Float() / 360
	} else {
		// Mozilla
		delta = event.Get("detail").Float() / -9
	}
	var z = math.Pow(1+p.zoomScale, delta)

	if p.origin == nil {
		// when no origing yet, wait a bit for first event to fillout
		return nil
	}

	x := p.origin.At(0, 0)
	y := p.origin.At(1, 0)

	// move to center, apply zoom, move back
	k := identity()
	k.Mul(translate(-x, -y), k)
	k.Mul(scale(z), k)
	k.Mul(translate(x, y), k)

	p.transform.Mul(k, p.transform)
	p.SetRootTranslation()
	return nil
}

func (p *SVGPanZoomer) handleMouseMove(_ js.Value, args []js.Value) interface{} {
	event := args[0]
	point := getEventPoint(event)

	if p.state != drag {
		p.origin = point
		return nil
	}

	x := point.At(0, 0)
	y := point.At(1, 0)
	ox := p.origin.At(0, 0)
	oy := p.origin.At(1, 0)

	// revert to original scaling, compute translation, move back to new scaling
	var k mat.Dense
	k.Inverse(p.transformBeforeDrag)
	k.Mul(&k, translate(x-ox, y-oy))
	k.Mul(p.transformBeforeDrag, &k)

	// add new scaling
	p.transform.Mul(&k, p.transformBeforeDrag)
	p.SetRootTranslation()
	return nil
}

func (p *SVGPanZoomer) handleMouseDown(_ js.Value, args []js.Value) interface{} {
	p.state = drag
	event := args[0]
	p.origin = getEventPoint(event)
	p.transformBeforeDrag = mat.DenseCopyOf(p.transform)
	return nil
}

func (p *SVGPanZoomer) handleMouseUp(_ js.Value, _ []js.Value) interface{} {
	p.state = notActive
	return nil
}

func (p *SVGPanZoomer) Reset() *SVGPanZoomer {
	p.transform = identity()
	p.transformBeforeDrag = nil
	p.origin = nil
	return p
}

func (p *SVGPanZoomer) Translate(dx, dy float64) *SVGPanZoomer {
	p.transform.Mul(p.transform, translate(dx, dy))
	return p
}

func (p *SVGPanZoomer) ScaleAt(z, x, y float64) *SVGPanZoomer {
	k := identity()
	k.Mul(translate(-x, -y), k)
	k.Mul(scale(z), k)
	k.Mul(translate(x, y), k)

	p.transform.Mul(k, p.transform)
	return p
}

// CenterBox makes given bounding box in the center and fill the screen.
func (p *SVGPanZoomer) CenterBox(minx, miny, maxx, maxy float64) *SVGPanZoomer {
	wScreen := js.Global().Get("window").Get("innerWidth").Float()
	hScreen := js.Global().Get("window").Get("innerHeight").Float()

	dx, dy, zoom := centerBox(wScreen, hScreen, minx, miny, maxx, maxy)
	p.Reset().Translate(dx, dy).ScaleAt(zoom, wScreen/2, hScreen/2)
	return p
}
