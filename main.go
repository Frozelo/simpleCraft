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
	Color    rl.Color
	Selected bool
}

type Game struct {
	shapes []Shape

	squaresCount int
	circlesCount int

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
		g.update()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		g.draw()
		rl.EndDrawing()
	}
}

func (g *Game) update() {
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

	// On release, distinguish a drag selection from a click on a shape.
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && g.selecting {
		end := rl.GetMousePosition()
		dx := end.X - g.selectionStart.X
		dy := end.Y - g.selectionStart.Y
		// Treat the input as a drag when the cursor moved farther than the threshold
		// between the press and release positions. Compare squared distances to avoid
		// computing a square root: dx² + dy² > threshold².
		isDrag := dx*dx+dy*dy > selectionDragThreshold*selectionDragThreshold

		if isDrag {
			// Dragging selects every shape inside the selection rectangle.
			g.selectShapes(selectionRect(g.selectionStart, end))
			g.hasLastClick = false
		} else if index := g.shapeAt(end); index >= 0 {
			shape := g.shapes[index]
			now := rl.GetTime()
			isDoubleClick := g.hasLastClick &&
				shape.Pos == g.lastClickPosition &&
				shape.Type == g.lastClickShapeType &&
				now-g.lastClickAt <= doubleClickInterval

			if isDoubleClick {
				// A second click selects every shape of the same type.
				g.selectAllShapesOfType(shape.Type)
				g.hasLastClick = false
			} else {
				// A single click selects only this shape and starts the double-click timer.
				g.selectShape(index)
				g.lastClickAt = now
				g.lastClickPosition = shape.Pos
				g.lastClickShapeType = shape.Type
				g.hasLastClick = true
			}
		} else {
			// Clicking empty space clears the current selection.
			g.clearSelection()
			g.hasLastClick = false
		}

		g.selecting = false
	}

	if rl.IsKeyPressed(rl.KeyBackspace) {
		g.deleteSelectedShapes()
	}
}

func (g *Game) draw() {
	var selection rl.Rectangle
	if g.selecting {
		selection = selectionRect(g.selectionStart, rl.GetMousePosition())
		rl.DrawRectangleRec(selection, rl.Fade(rl.SkyBlue, 0.2))
	}

	selected := 0
	preSelected := 0
	for _, shape := range g.shapes {
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

	rl.DrawText("Squares: "+strconv.Itoa(g.squaresCount), 10, 50, 20, rl.Red)
	rl.DrawText("Circles: "+strconv.Itoa(g.circlesCount), 10, 70, 20, rl.Blue)
	rl.DrawText("Selected: "+strconv.Itoa(selected), 10, 10, 20, rl.Green)
	rl.DrawText("PreSelected: "+strconv.Itoa(preSelected), 10, 30, 20, rl.Gray)
}

func (g *Game) addShape(shapeType ShapeType, position rl.Vector2) {
	color := squareColor
	if shapeType == ShapeCircle {
		g.circlesCount++
		color = circleColor
	} else {
		g.squaresCount++
	}

	g.shapes = append(g.shapes, Shape{Type: shapeType, Pos: position, Color: color})
	slog.Info("shape created", "type", shapeType, "x", position.X, "y", position.Y)
}

func (g *Game) deleteSelectedShapes() {
	for i := len(g.shapes) - 1; i >= 0; i-- {
		if g.shapes[i].Selected {
			switch g.shapes[i].Type {
			case ShapeSquare:
				g.squaresCount--
			case ShapeCircle:
				g.circlesCount--
			}

			g.shapes = slices.Delete(g.shapes, i, i+1)
		}
	}
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
