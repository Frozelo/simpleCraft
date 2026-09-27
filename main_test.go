package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAddShape(t *testing.T) {
	g := &Game{}
	squarePos := rl.Vector2{X: 12, Y: 34}
	circlePos := rl.Vector2{X: 56, Y: 78}

	g.addShape(ShapeSquare, squarePos)
	g.addShape(ShapeCircle, circlePos)

	if len(g.shapes) != 2 || g.shapes[0].Type != ShapeSquare || g.shapes[0].Pos != squarePos || g.shapes[0].Color != squareColor ||
		g.shapes[1].Type != ShapeCircle || g.shapes[1].Pos != circlePos || g.shapes[1].Color != circleColor {
		t.Fatalf("shapes = %#v, want a square and a circle at their requested positions", g.shapes)
	}
}

func TestClickSelectsTopmostShapeAndEmptyClickClearsSelection(t *testing.T) {
	position := rl.Vector2{X: 25, Y: 25}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Pos: position},
		{Type: ShapeCircle, Pos: position},
	}}

	releaseSelection(g, position, position, 1)
	if g.shapes[0].Selected || !g.shapes[1].Selected || !g.hasLastClick || g.selecting {
		t.Fatalf("after click: game = %#v, want only topmost shape selected", g)
	}

	empty := rl.Vector2{X: 100, Y: 100}
	releaseSelection(g, empty, empty, 2)
	if g.shapes[0].Selected || g.shapes[1].Selected || g.hasLastClick {
		t.Fatalf("after empty click: game = %#v, want no selection or pending click", g)
	}
}

func TestDoubleClickSelectsAllShapesOfSameType(t *testing.T) {
	position := rl.Vector2{X: 20, Y: 20}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Pos: position},
		{Type: ShapeSquare, Pos: rl.Vector2{X: 60, Y: 60}},
		{Type: ShapeCircle, Pos: rl.Vector2{X: 100, Y: 100}},
	}}

	releaseSelection(g, position, position, 1)
	releaseSelection(g, position, position, 1+doubleClickInterval/2)
	if !g.shapes[0].Selected || !g.shapes[1].Selected || g.shapes[2].Selected || g.hasLastClick {
		t.Fatalf("after double click: game = %#v, want both squares selected", g)
	}

	releaseSelection(g, position, position, 2)
	if !g.hasLastClick || g.shapes[1].Selected {
		t.Fatalf("after next click: game = %#v, want a new single-click sequence", g)
	}
}

func TestDragSelectsIntersectingShapesAndResetsClick(t *testing.T) {
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Pos: rl.Vector2{X: 20, Y: 20}},
		{Type: ShapeCircle, Pos: rl.Vector2{X: 45, Y: 45}},
		{Type: ShapeSquare, Pos: rl.Vector2{X: 100, Y: 100}, Selected: true},
	}, hasLastClick: true}

	releaseSelection(g, rl.Vector2{X: 50, Y: 50}, rl.Vector2{X: 10, Y: 10}, 1)
	if !g.shapes[0].Selected || !g.shapes[1].Selected || g.shapes[2].Selected || g.hasLastClick || g.selecting {
		t.Fatalf("after drag: game = %#v, want intersecting shapes selected", g)
	}
}

func TestMovementAtThresholdRemainsClick(t *testing.T) {
	g := &Game{shapes: []Shape{{Type: ShapeSquare, Pos: rl.Vector2{X: 10, Y: 10}}}}
	releaseSelection(g, rl.Vector2{X: 10, Y: 10}, rl.Vector2{X: 15, Y: 10}, 1)
	if !g.shapes[0].Selected || !g.hasLastClick {
		t.Fatalf("game = %#v, want a click at the drag threshold", g)
	}
}

func TestDeleteSelectedShapes(t *testing.T) {
	kept := Shape{Type: ShapeCircle, Pos: rl.Vector2{X: 30, Y: 30}}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Selected: true},
		kept,
		{Type: ShapeCircle, Selected: true},
	}, hasLastClick: true}

	g.deleteSelectedShapes()
	if len(g.shapes) != 1 || g.shapes[0] != kept || g.hasLastClick {
		t.Fatalf("game = %#v, want only the unselected shape", g)
	}
}

func TestSelectionRectForClickUsesCursorPoint(t *testing.T) {
	position := rl.Vector2{X: 12, Y: 34}
	rect := selectionRect(position, position)

	if rect.X != position.X || rect.Y != position.Y || rect.Width != 1 || rect.Height != 1 {
		t.Fatalf("selectionRect() = %#v, want a 1x1 rectangle at %#v", rect, position)
	}
}

func TestMoveSelectedShapes(t *testing.T) {
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Pos: rl.Vector2{X: 10}, Selected: true},
		{Type: ShapeCircle, Pos: rl.Vector2{X: 20}, Selected: true},
		{Type: ShapeSquare, Pos: rl.Vector2{X: 30}},
	}}
	target := rl.Vector2{X: 100, Y: 50}
	g.moveSelectedShapes(target)
	for i := 0; i < 2; i++ {
		if g.shapes[i].Target != target || !g.shapes[i].IsMoving || g.shapes[i].Pos.X != float32((i+1)*10) {
			t.Fatalf("shape %d = %#v, want target without teleport", i, g.shapes[i])
		}
	}
	if g.shapes[2].IsMoving || g.shapes[2].Target != (rl.Vector2{}) {
		t.Fatalf("unselected shape = %#v, want unchanged", g.shapes[2])
	}
}

func TestAdvanceShapesUsesTypeSpeedAndStopsAtTarget(t *testing.T) {
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Target: rl.Vector2{X: 200}, IsMoving: true},
		{Type: ShapeCircle, Target: rl.Vector2{X: 120}, IsMoving: true},
	}}
	g.advanceShapes(0.25)
	if g.shapes[0].Pos.X != 50 || g.shapes[1].Pos.X != 30 {
		t.Fatalf("positions = %v and %v, want 50 and 30", g.shapes[0].Pos.X, g.shapes[1].Pos.X)
	}
	g.advanceShapes(1)
	for _, shape := range g.shapes {
		if shape.Pos != shape.Target || shape.IsMoving {
			t.Fatalf("shape = %#v, want exact target and stopped", shape)
		}
	}
}

func releaseSelection(g *Game, start, end rl.Vector2, now float64) {
	g.selectionStart = start
	g.selecting = true
	g.finishSelection(end, now)
}
