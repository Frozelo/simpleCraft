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

	if len(g.shapes) != 2 || g.shapes[0].Type != ShapeSquare || g.shapes[0].Team != TeamPlayer || g.shapes[0].Pos != squarePos ||
		g.shapes[1].Type != ShapeCircle || g.shapes[1].Team != TeamAI || g.shapes[1].Pos != circlePos {
		t.Fatalf("shapes = %#v, want figures with their assigned teams and positions", g.shapes)
	}
}

func TestTeamColor(t *testing.T) {
	if teamColor[TeamPlayer] != rl.Green || teamColor[TeamAI] != rl.Blue {
		t.Fatalf("team colors: player=%v AI=%v, want green and blue", teamColor[TeamPlayer], teamColor[TeamAI])
	}
}

func TestShapeGeometry(t *testing.T) {
	position := rl.Vector2{X: 20, Y: 20}
	corner := rl.Vector2{X: 29, Y: 29}
	if !shapeContainsPoint(Shape{Type: ShapeSquare, Pos: position}, corner) {
		t.Fatal("square should contain a point near its corner")
	}
	if shapeContainsPoint(Shape{Type: ShapeCircle, Pos: position}, corner) {
		t.Fatal("circle should not contain a point outside its radius")
	}
}

func TestClickSelectsTopmostShapeAndEmptyClickClearsSelection(t *testing.T) {
	position := rl.Vector2{X: 25, Y: 25}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Team: TeamPlayer, Pos: position},
		{Type: ShapeCircle, Pos: position, Team: TeamPlayer},
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

func TestEnemyShapesCannotBeSelectedMovedOrDeleted(t *testing.T) {
	position := rl.Vector2{X: 20, Y: 20}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Team: TeamPlayer, Pos: position},
		{Type: ShapeCircle, Team: TeamAI, Pos: position},
		{Type: ShapeSquare, Team: TeamAI, Pos: rl.Vector2{X: 60, Y: 60}},
	}}

	releaseSelection(g, position, position, 1)
	if !g.shapes[0].Selected || g.shapes[1].Selected {
		t.Fatalf("click selection = %#v, want only player shape selected", g.shapes)
	}

	g.selectShapes(rl.Rectangle{X: 0, Y: 0, Width: 100, Height: 100})
	if !g.shapes[0].Selected || g.shapes[1].Selected || g.shapes[2].Selected {
		t.Fatalf("drag selection = %#v, want only player shape selected", g.shapes)
	}

	g.selectAllShapesOfType(ShapeSquare)
	if !g.shapes[0].Selected || g.shapes[2].Selected {
		t.Fatalf("type selection = %#v, want only player square selected", g.shapes)
	}

	// Even inconsistent selection state must not allow commands to affect an enemy.
	g.shapes[1].Selected = true
	g.shapes[2].Selected = true
	target := rl.Vector2{X: 100, Y: 100}
	g.moveSelectedShapes(target)
	g.advanceShapes(0.25)
	if g.shapes[1].Target != (rl.Vector2{}) || g.shapes[1].Pos != position ||
		g.shapes[2].Target != (rl.Vector2{}) || g.shapes[2].Pos != (rl.Vector2{X: 60, Y: 60}) {
		t.Fatalf("enemy shapes moved: %#v", g.shapes)
	}

	g.deleteSelectedShapes()
	if len(g.shapes) != 2 || g.shapes[0].Team != TeamAI || g.shapes[1].Team != TeamAI {
		t.Fatalf("after deletion = %#v, want both enemy shapes left on field", g.shapes)
	}
}

func TestDoubleClickSelectsAllShapesOfSameType(t *testing.T) {
	position := rl.Vector2{X: 20, Y: 20}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Team: TeamPlayer, Pos: position},
		{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 60, Y: 60}},
		{Type: ShapeCircle, Team: TeamAI, Pos: rl.Vector2{X: 100, Y: 100}},
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
		{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 20, Y: 20}},
		{Type: ShapeCircle, Team: TeamPlayer, Pos: rl.Vector2{X: 45, Y: 45}},
		{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 100, Y: 100}, Selected: true},
	}, hasLastClick: true}

	releaseSelection(g, rl.Vector2{X: 50, Y: 50}, rl.Vector2{X: 10, Y: 10}, 1)
	if !g.shapes[0].Selected || !g.shapes[1].Selected || g.shapes[2].Selected || g.hasLastClick || g.selecting {
		t.Fatalf("after drag: game = %#v, want intersecting shapes selected", g)
	}
}

func TestMovementAtThresholdRemainsClick(t *testing.T) {
	g := &Game{shapes: []Shape{{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 10, Y: 10}}}}
	releaseSelection(g, rl.Vector2{X: 10, Y: 10}, rl.Vector2{X: 15, Y: 10}, 1)
	if !g.shapes[0].Selected || !g.hasLastClick {
		t.Fatalf("game = %#v, want a click at the drag threshold", g)
	}
}

func TestDeleteSelectedShapes(t *testing.T) {
	kept := Shape{Type: ShapeCircle, Team: TeamPlayer, Pos: rl.Vector2{X: 30, Y: 30}}
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Team: TeamPlayer, Selected: true},
		kept,
		{Type: ShapeCircle, Team: TeamPlayer, Selected: true},
	}, hasLastClick: true}

	g.deleteSelectedShapes()
	if len(g.shapes) != 1 || g.shapes[0] != kept || g.hasLastClick {
		t.Fatalf("game = %#v, want only the unselected shape", g)
	}
}

func TestMoveSelectedShapesSetsIndividualTargets(t *testing.T) {
	g := &Game{shapes: []Shape{
		{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 10}, Selected: true},
		{Type: ShapeCircle, Team: TeamPlayer, Pos: rl.Vector2{X: 20}, Selected: true},
		{Type: ShapeSquare, Team: TeamPlayer, Pos: rl.Vector2{X: 30}},
	}}
	target := rl.Vector2{X: 100, Y: 50}
	g.moveSelectedShapes(target)

	for i := 0; i < 2; i++ {
		if g.shapes[i].Target != target || !g.shapes[i].IsMoving || g.shapes[i].Pos.X != float32((i+1)*10) {
			t.Fatalf("shape %d = %#v, want target assigned without immediate movement", i, g.shapes[i])
		}
	}
	if g.shapes[2].Target != (rl.Vector2{}) || g.shapes[2].IsMoving {
		t.Fatalf("unselected shape = %#v, want no movement", g.shapes[2])
	}

	g.shapes[0].Selected = false
	g.shapes[1].Selected = false
	g.moveSelectedShapes(rl.Vector2{X: 200})
	if g.shapes[0].Target != target || g.shapes[1].Target != target {
		t.Fatal("command without a selection changed moving shapes")
	}

	g.shapes[1].Selected = true
	newTarget := rl.Vector2{X: 200}
	g.moveSelectedShapes(newTarget)
	if g.shapes[0].Target != target || g.shapes[1].Target != newTarget {
		t.Fatalf("shapes = %#v, want only the selected circle retargeted", g.shapes)
	}
}

func TestAdvanceShapesUsesFrameTimeAndStopsAtTarget(t *testing.T) {
	g := &Game{shapes: []Shape{{Type: ShapeSquare, Team: TeamPlayer, Target: rl.Vector2{X: 200}, IsMoving: true}}}

	g.advanceShapes(0.25)
	if g.shapes[0].Pos != (rl.Vector2{X: 50}) || !g.shapes[0].IsMoving {
		t.Fatalf("after 0.25s: shape = %#v, want x=50 and still moving", g.shapes[0])
	}
	g.advanceShapes(0.5)
	if g.shapes[0].Pos != (rl.Vector2{X: 150}) || !g.shapes[0].IsMoving {
		t.Fatalf("after 0.75s: shape = %#v, want x=150 and still moving", g.shapes[0])
	}
	g.advanceShapes(1)
	if g.shapes[0].Pos != g.shapes[0].Target || g.shapes[0].IsMoving {
		t.Fatalf("after arrival: shape = %#v, want exact target and stopped", g.shapes[0])
	}
	g.advanceShapes(1)
	if g.shapes[0].Pos != g.shapes[0].Target {
		t.Fatalf("after another frame: shape = %#v, want to stay at target", g.shapes[0])
	}
}

func TestAdvanceShapesHasSameSpeedDiagonally(t *testing.T) {
	g := &Game{shapes: []Shape{{Type: ShapeCircle, Team: TeamPlayer, Target: rl.Vector2{X: 72, Y: 96}, IsMoving: true}}}
	g.advanceShapes(0.5)
	if g.shapes[0].Pos != (rl.Vector2{X: 36, Y: 48}) || !g.shapes[0].IsMoving {
		t.Fatalf("shape = %#v, want 60 pixels of diagonal movement", g.shapes[0])
	}
}

func TestMoveSelectedShapeToItsCurrentPositionStopsIt(t *testing.T) {
	position := rl.Vector2{X: 10, Y: 20}
	g := &Game{shapes: []Shape{{Team: TeamPlayer, Pos: position, Target: rl.Vector2{X: 100}, Selected: true, IsMoving: true}}}
	g.moveSelectedShapes(position)
	if g.shapes[0].Pos != position || g.shapes[0].Target != position || g.shapes[0].IsMoving {
		t.Fatalf("shape = %#v, want movement canceled at current position", g.shapes[0])
	}
}

func TestSelectionRectForClickUsesCursorPoint(t *testing.T) {
	position := rl.Vector2{X: 12, Y: 34}
	rect := selectionRect(position, position)

	if rect.X != position.X || rect.Y != position.Y || rect.Width != 1 || rect.Height != 1 {
		t.Fatalf("selectionRect() = %#v, want a 1x1 rectangle at %#v", rect, position)
	}
}

func releaseSelection(g *Game, start, end rl.Vector2, now float64) {
	g.selectionStart = start
	g.selecting = true
	g.finishSelection(end, now)
}
