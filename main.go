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
	doubleClickInterval    = 0.35
	selectionDragThreshold = 5.0
)

type ShapeType uint16

const (
	ShapeSquare ShapeType = iota
	ShapeCircle
	shapeTypeCount
)

type Team uint8

const (
	TeamPlayer Team = iota
	TeamAI
)

type ShapeGeometry uint8

const (
	GeometrySquare ShapeGeometry = iota
	GeometryCircle
)

var teamColor = map[Team]rl.Color{
	TeamPlayer: rl.Green,
	TeamAI:     rl.Blue,
}

type ShapeDefinition struct {
	Name        string
	Geometry    ShapeGeometry
	Size        float32
	MoveSpeed   float32
	DefaultTeam Team
	SpawnKey    int32
}

var shapeDefinitions = [shapeTypeCount]ShapeDefinition{
	ShapeSquare: {Name: "Squares", Geometry: GeometrySquare, Size: 20, MoveSpeed: 200, DefaultTeam: TeamPlayer, SpawnKey: rl.KeyS},
	ShapeCircle: {Name: "Circles", Geometry: GeometryCircle, Size: 20, MoveSpeed: 120, DefaultTeam: TeamAI, SpawnKey: rl.KeyC},
}

type Shape struct {
	Type   ShapeType
	Team   Team
	Pos    rl.Vector2
	Target rl.Vector2

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
	for shapeType, definition := range shapeDefinitions {
		if definition.SpawnKey != 0 && rl.IsKeyPressed(definition.SpawnKey) {
			g.addShape(ShapeType(shapeType), rl.GetMousePosition())
		}
	}

	if rl.IsKeyPressed(rl.KeyBackspace) {
		g.deleteSelectedShapes()
	}

	// Start tracking a selection when the left mouse button is pressed.
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		g.selecting = true
		g.selectionStart = rl.GetMousePosition()
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && g.selecting {
		g.finishSelection(rl.GetMousePosition(), rl.GetTime())
	}

	if rl.IsMouseButtonPressed(rl.MouseRightButton) {
		g.moveSelectedShapes(rl.GetMousePosition())
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
	playerCount := 0
	enemyCount := 0
	var counts [shapeTypeCount]int
	for _, shape := range g.shapes {
		counts[shape.Type]++

		color := teamColor[shape.Team]
		if shape.Team == TeamPlayer {
			playerCount++
		} else {
			enemyCount++
		}
		if shape.Selected {
			selected++
		}
		preview := g.selecting && shape.Team == TeamPlayer && shapeIntersectsRect(shape, selection)
		if preview {
			preSelected++
		}

		drawShape(shape, color)
		if preview {
			drawShapeOutline(shape, rl.Orange)
		} else if shape.Team == TeamPlayer && shape.Selected {
			drawShapeOutline(shape, rl.Gold)
		}
	}

	if g.selecting {
		rl.DrawRectangleLinesEx(selection, 1, rl.SkyBlue)
	}

	for shapeType, definition := range shapeDefinitions {
		rl.DrawText(definition.Name+": "+strconv.Itoa(counts[shapeType]), 10, int32(50+20*shapeType), 20, rl.LightGray)
	}

	rl.DrawText("Player: "+strconv.Itoa(playerCount), 10, int32(50+20*len(shapeDefinitions)), 20, teamColor[TeamPlayer])
	rl.DrawText("Enemy: "+strconv.Itoa(enemyCount), 10, int32(70+20*len(shapeDefinitions)), 20, teamColor[TeamAI])
	rl.DrawText("Selected: "+strconv.Itoa(selected), 10, 10, 20, rl.Green)
	rl.DrawText("PreSelected: "+strconv.Itoa(preSelected), 10, 30, 20, rl.Gray)
}

func (g *Game) addShape(shapeType ShapeType, position rl.Vector2) {
	definition := shapeDefinitions[shapeType]
	g.shapes = append(g.shapes, Shape{Type: shapeType, Team: definition.DefaultTeam, Pos: position})
	slog.Info("shape created", "type", shapeType, "team", definition.DefaultTeam, "x", position.X, "y", position.Y)
}

func (g *Game) deleteSelectedShapes() {
	g.shapes = slices.DeleteFunc(g.shapes, func(shape Shape) bool { return shape.Team == TeamPlayer && shape.Selected })
	g.hasLastClick = false
}

func (g *Game) selectShapes(rect rl.Rectangle) {
	for i := range g.shapes {
		shape := &g.shapes[i]
		shape.Selected = shape.Team == TeamPlayer && shapeIntersectsRect(*shape, rect)
	}
}

func (g *Game) shapeAt(position rl.Vector2) int {
	for i := len(g.shapes) - 1; i >= 0; i-- {
		shape := g.shapes[i]
		if shape.Team == TeamPlayer && shapeContainsPoint(shape, position) {
			return i
		}
	}

	return -1
}

func (g *Game) selectShape(index int) {
	for i := range g.shapes {
		g.shapes[i].Selected = false
	}

	g.shapes[index].Selected = g.shapes[index].Team == TeamPlayer
}

func (g *Game) selectAllShapesOfType(shapeType ShapeType) {
	for i := range g.shapes {
		shape := &g.shapes[i]
		shape.Selected = shape.Team == TeamPlayer && shape.Type == shapeType
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
		if shape.Team == TeamPlayer && shape.Selected {
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
		if shape.Team != TeamPlayer || !shape.IsMoving {
			continue
		}

		step := shapeDefinitions[shape.Type].MoveSpeed * dt

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
	definition := shapeDefinitions[shape.Type]
	switch definition.Geometry {
	case GeometrySquare:
		rl.DrawRectangleRec(squareRect(shape.Pos, definition.Size), color)
	case GeometryCircle:
		rl.DrawCircleV(shape.Pos, definition.Size/2, color)
	}
}

func drawShapeOutline(shape Shape, color rl.Color) {
	definition := shapeDefinitions[shape.Type]
	switch definition.Geometry {
	case GeometrySquare:
		rl.DrawRectangleLinesEx(squareRect(shape.Pos, definition.Size+4), 2, color)
	case GeometryCircle:
		rl.DrawCircleLinesV(shape.Pos, definition.Size/2+3, color)
	}
}

func shapeContainsPoint(shape Shape, point rl.Vector2) bool {
	definition := shapeDefinitions[shape.Type]
	switch definition.Geometry {
	case GeometrySquare:
		return rl.CheckCollisionPointRec(point, squareRect(shape.Pos, definition.Size))
	case GeometryCircle:
		return rl.CheckCollisionPointCircle(point, shape.Pos, definition.Size/2)
	default:
		return false
	}
}

func shapeIntersectsRect(shape Shape, rect rl.Rectangle) bool {
	definition := shapeDefinitions[shape.Type]
	switch definition.Geometry {
	case GeometrySquare:
		return rl.CheckCollisionRecs(squareRect(shape.Pos, definition.Size), rect)
	case GeometryCircle:
		return rl.CheckCollisionCircleRec(shape.Pos, definition.Size/2, rect)
	default:
		return false
	}
}

func squareRect(position rl.Vector2, size float32) rl.Rectangle {
	return rl.Rectangle{
		X:      position.X - size/2,
		Y:      position.Y - size/2,
		Width:  size,
		Height: size,
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
