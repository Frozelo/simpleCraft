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
	squareToggleCooldown   = 0.05
	doubleClickInterval    = 0.35
	selectionDragThreshold = 5.0
)

type Square struct {
	Pos      rl.Vector2
	Selected bool
}

type Game struct {
	squares        []Square
	nextToggleAt   float64
	selectionStart rl.Vector2
	selecting      bool

	lastClickAt     float64
	lastClickSquare rl.Vector2
	hasLastClick    bool
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
	// Keep the keyboard shortcut for creating or removing squares.
	if rl.IsKeyDown(rl.KeyA) {
		g.toggleSquare(rl.GetMousePosition(), rl.GetTime())
	}

	// Start tracking a selection when the left mouse button is pressed.
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		g.selecting = true
		g.selectionStart = rl.GetMousePosition()
	}

	// On release, distinguish a drag selection from a click on a square.
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && g.selecting {
		end := rl.GetMousePosition()
		dx := end.X - g.selectionStart.X
		dy := end.Y - g.selectionStart.Y
		// Treat the input as a drag when the cursor moved farther than the threshold
		// between the press and release positions. Compare squared distances to avoid
		// computing a square root: dx² + dy² > threshold².
		isDrag := dx*dx+dy*dy > selectionDragThreshold*selectionDragThreshold

		if isDrag {
			// Dragging selects every square inside the selection rectangle.
			g.selectSquares(selectionRect(g.selectionStart, end))
			g.hasLastClick = false
		} else if index := g.squareAt(end); index >= 0 {
			squarePos := g.squares[index].Pos
			now := rl.GetTime()
			isDoubleClick := g.hasLastClick &&
				squarePos == g.lastClickSquare &&
				now-g.lastClickAt <= doubleClickInterval

			if isDoubleClick {
				// A second click on the same square selects all squares.
				g.selectAllSquares()
				g.hasLastClick = false
			} else {
				// A single click selects only this square and starts the double-click timer.
				g.selectSquare(index)
				g.lastClickAt = now
				g.lastClickSquare = squarePos
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
		g.deleteSelectedSquares()
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
	for _, square := range g.squares {
		rect := squareRect(square.Pos)
		color := rl.RayWhite
		if square.Selected {
			color = rl.Gold
			selected++
		}
		if g.selecting && rl.CheckCollisionRecs(rect, selection) {
			color = rl.Orange
			preSelected++
		}

		rl.DrawRectangleRec(rect, color)
	}

	if g.selecting {
		rl.DrawRectangleLinesEx(selection, 1, rl.SkyBlue)
	}

	rl.DrawText("Selected: "+strconv.Itoa(selected), 10, 10, 20, rl.Green)
	rl.DrawText("PreSelected: "+strconv.Itoa(preSelected), 10, 30, 20, rl.Gray)
}

func (g *Game) toggleSquare(position rl.Vector2, now float64) bool {
	if now < g.nextToggleAt {
		return false
	}

	if index := slices.IndexFunc(g.squares, func(square Square) bool {
		return square.Pos == position
	}); index >= 0 {
		g.squares = slices.Delete(g.squares, index, index+1)
		slog.Info("square removed", "x", position.X, "y", position.Y)
	} else {
		g.squares = append(g.squares, Square{Pos: position})
		slog.Info("square created", "x", position.X, "y", position.Y)
	}

	g.nextToggleAt = now + squareToggleCooldown
	return true
}

func (g *Game) deleteSelectedSquares() {
	for i := len(g.squares) - 1; i >= 0; i-- {
		if g.squares[i].Selected {
			g.squares = slices.Delete(g.squares, i, i+1)
		}
	}
}

func (g *Game) selectSquares(rect rl.Rectangle) {
	for i := range g.squares {
		g.squares[i].Selected = rl.CheckCollisionRecs(squareRect(g.squares[i].Pos), rect)
	}
}

func (g *Game) squareAt(position rl.Vector2) int {
	for i := len(g.squares) - 1; i >= 0; i-- {
		if rl.CheckCollisionPointRec(position, squareRect(g.squares[i].Pos)) {
			return i
		}
	}

	return -1
}

func (g *Game) selectSquare(index int) {
	for i := range g.squares {
		g.squares[i].Selected = false
	}

	g.squares[index].Selected = true
}

func (g *Game) selectAllSquares() {
	for i := range g.squares {
		g.squares[i].Selected = true
	}
}

func (g *Game) clearSelection() {
	for i := range g.squares {
		g.squares[i].Selected = false
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
