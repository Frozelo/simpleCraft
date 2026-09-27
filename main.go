package main

import (
	"log/slog"
	"slices"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth            = 1280
	windowHeight           = 720
	windowTitle            = "SimpleCraft Sim"
	targetFPS              = 60
	squareMoveSpeed        = 200
	circleMoveSpeed        = 120
	squareSize             = 20
	circleRadius           = float32(squareSize) / 2
	doubleClickInterval    = 0.35
	selectionDragThreshold = 5.0
)

var (
	circleColor = rl.Blue
	squareColor = rl.Red
)

type ShapeType uint8

const (
	ShapeSquare ShapeType = iota
	ShapeCircle
)

type Shape struct {
	Type     ShapeType
	Pos      rl.Vector2
	Target   rl.Vector2
	Color    rl.Color
	Selected bool
	IsMoving bool
}

type Game struct {
	shapes []Shape

	selectionStart rl.Vector2
	selecting      bool

	lastClickAt        float64
	lastClickPosition  rl.Vector2
	lastClickShapeType ShapeType
	hasLastClick       bool
}

func main() {
	var game Game
	game.run()
}

func (g *Game) run() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(targetFPS)

	for !rl.WindowShouldClose() {
		g.update(rl.GetFrameTime())

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		g.draw()
		rl.EndDrawing()
	}
}

func (g *Game) update(dt float32) {
	// Create one shape at the cursor position when its key is pressed.
	if rl.IsKeyPressed(rl.KeyS) {
		g.addShape(ShapeSquare, rl.GetMousePosition())
	}
	if rl.IsKeyPressed(rl.KeyC) {
		g.addShape(ShapeCircle, rl.GetMousePosition())
	}

	// Start tracking a selection when the left mouse button is pressed.
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		g.selecting = true
		g.selectionStart = rl.GetMousePosition()
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && g.selecting {
		g.finishSelection(rl.GetMousePosition(), rl.GetTime())
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonRight) {
		g.moveSelectedShapes(rl.GetMousePosition())
	}

	if rl.IsKeyPressed(rl.KeyBackspace) {
		g.deleteSelectedShapes()
	}
	g.advanceShapes(dt)
}

func (g *Game) finishSelection(end rl.Vector2, now float64) {
	dx := end.X - g.selectionStart.X
	dy := end.Y - g.selectionStart.Y
	// Compare squared distances so the threshold does not require a square root.
	if dx*dx+dy*dy > selectionDragThreshold*selectionDragThreshold {
		g.selectShapes(selectionRect(g.selectionStart, end))
		g.hasLastClick = false
	} else {
		g.selectClickedShape(end, now)
	}
	g.selecting = false
}

func (g *Game) selectClickedShape(position rl.Vector2, now float64) {
	index := g.shapeAt(position)
	if index < 0 {
		g.clearSelection()
		g.hasLastClick = false
		return
	}

	shape := g.shapes[index]
	isDoubleClick := g.hasLastClick &&
		shape.Pos == g.lastClickPosition &&
		shape.Type == g.lastClickShapeType &&
		now-g.lastClickAt <= doubleClickInterval
	if isDoubleClick {
		g.selectAllShapesOfType(shape.Type)
		g.hasLastClick = false
		return
	}

	g.selectShape(index)
	g.lastClickAt = now
	g.lastClickPosition = shape.Pos
	g.lastClickShapeType = shape.Type
	g.hasLastClick = true
}

func (g *Game) draw() {
	var selection rl.Rectangle
	if g.selecting {
		selection = selectionRect(g.selectionStart, rl.GetMousePosition())
		rl.DrawRectangleRec(selection, rl.Fade(rl.SkyBlue, 0.2))
	}

	selected := 0
	preSelected := 0
	squares := 0
	circles := 0
	for _, shape := range g.shapes {
		switch shape.Type {
		case ShapeSquare:
			squares++
		case ShapeCircle:
			circles++
		}

		color := shape.Color
		if shape.Selected {
			color = rl.Gold
			selected++
		}
		if g.selecting && shapeIntersectsRect(shape, selection) {
			color = rl.Orange
			preSelected++
		}

		drawShape(shape, color)
	}

	if g.selecting {
		rl.DrawRectangleLinesEx(selection, 1, rl.SkyBlue)
	}

	rl.DrawText("Squares: "+strconv.Itoa(squares), 10, 50, 20, rl.Red)
	rl.DrawText("Circles: "+strconv.Itoa(circles), 10, 70, 20, rl.Blue)
	rl.DrawText("Selected: "+strconv.Itoa(selected), 10, 10, 20, rl.Green)
	rl.DrawText("PreSelected: "+strconv.Itoa(preSelected), 10, 30, 20, rl.Gray)
}

func (g *Game) addShape(shapeType ShapeType, position rl.Vector2) {
	color := squareColor
	if shapeType == ShapeCircle {
		color = circleColor
	}

	g.shapes = append(g.shapes, Shape{Type: shapeType, Pos: position, Color: color})
	slog.Info("shape created", "type", shapeType, "x", position.X, "y", position.Y)
}

func (g *Game) deleteSelectedShapes() {
	g.shapes = slices.DeleteFunc(g.shapes, func(shape Shape) bool { return shape.Selected })
	g.hasLastClick = false
}

func (g *Game) selectShapes(rect rl.Rectangle) {
	for i := range g.shapes {
		g.shapes[i].Selected = shapeIntersectsRect(g.shapes[i], rect)
	}
}

func (g *Game) shapeAt(position rl.Vector2) int {
	for i := len(g.shapes) - 1; i >= 0; i-- {
		shape := g.shapes[i]
		if shapeContainsPoint(shape, position) {
			return i
		}
	}

	return -1
}

func (g *Game) selectShape(index int) {
	for i := range g.shapes {
		g.shapes[i].Selected = false
	}

	g.shapes[index].Selected = true
}

func (g *Game) selectAllShapesOfType(shapeType ShapeType) {
	for i := range g.shapes {
		g.shapes[i].Selected = g.shapes[i].Type == shapeType
	}
}

func (g *Game) clearSelection() {
	for i := range g.shapes {
		g.shapes[i].Selected = false
	}
}

func (g *Game) moveSelectedShapes(target rl.Vector2) {
	for i := range g.shapes {
		shape := &g.shapes[i]
		if shape.Selected {
			shape.Target = target
			shape.IsMoving = shape.Pos != target
		}
	}
}

func (g *Game) advanceShapes(dt float32) {
	if dt <= 0 {
		return
	}
	for i := range g.shapes {
		shape := &g.shapes[i]
		if !shape.IsMoving {
			continue
		}
		step := circleMoveSpeed * dt
		if shape.Type == ShapeSquare {
			step = squareMoveSpeed * dt
		}
		dx := shape.Target.X - shape.Pos.X
		dy := shape.Target.Y - shape.Pos.Y
		distance := rl.Vector2Distance(shape.Pos, shape.Target)
		if distance <= step {
			shape.Pos = shape.Target
			shape.IsMoving = false
			continue
		}
		shape.Pos.X += dx / distance * step
		shape.Pos.Y += dy / distance * step
	}
}

func drawShape(shape Shape, color rl.Color) {
	switch shape.Type {
	case ShapeSquare:
		rl.DrawRectangleRec(squareRect(shape.Pos), color)
	case ShapeCircle:
		rl.DrawCircleV(shape.Pos, circleRadius, color)
	}
}

func shapeContainsPoint(shape Shape, point rl.Vector2) bool {
	switch shape.Type {
	case ShapeSquare:
		return rl.CheckCollisionPointRec(point, squareRect(shape.Pos))
	case ShapeCircle:
		return rl.CheckCollisionPointCircle(point, shape.Pos, circleRadius)
	default:
		return false
	}
}

func shapeIntersectsRect(shape Shape, rect rl.Rectangle) bool {
	switch shape.Type {
	case ShapeSquare:
		return rl.CheckCollisionRecs(squareRect(shape.Pos), rect)
	case ShapeCircle:
		return rl.CheckCollisionCircleRec(shape.Pos, circleRadius, rect)
	default:
		return false
	}
}

func squareRect(position rl.Vector2) rl.Rectangle {
	return rl.Rectangle{
		X:      position.X - squareSize/2,
		Y:      position.Y - squareSize/2,
		Width:  squareSize,
		Height: squareSize,
	}
}

func selectionRect(start, end rl.Vector2) rl.Rectangle {
	if start == end {
		return rl.Rectangle{
			X:      start.X,
			Y:      start.Y,
			Width:  1,
			Height: 1,
		}
	}

	if start.X > end.X {
		start.X, end.X = end.X, start.X
	}

	if start.Y > end.Y {
		start.Y, end.Y = end.Y, start.Y
	}

	return rl.Rectangle{
		X:      start.X,
		Y:      start.Y,
		Width:  end.X - start.X,
		Height: end.Y - start.Y,
	}
}
