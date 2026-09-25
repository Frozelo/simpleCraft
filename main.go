package main

import (
	"log/slog"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth          = 1280
	windowHeight         = 720
	windowTitle          = "SimpleCraft Sim"
	targetFPS            = 60
	squareSize           = 20
	squareToggleCooldown = 0.05
)

type Game struct {
	squares      []rl.Vector2
	nextToggleAt float64
}

func (g *Game) Init() {
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	rl.SetTargetFPS(targetFPS)
}

func (g *Game) ToggleSquare(position rl.Vector2) {
	if index := slices.Index(g.squares, position); index >= 0 {
		g.squares = slices.Delete(g.squares, index, index+1)
		slog.Info("square removed", "x", position.X, "y", position.Y)
		return
	}

	g.squares = append(g.squares, position)
	slog.Info("square created", "x", position.X, "y", position.Y)
}

func (g *Game) TryToggleSquare(position rl.Vector2, now float64) bool {
	if now < g.nextToggleAt {
		return false
	}

	g.ToggleSquare(position)
	g.nextToggleAt = now + squareToggleCooldown
	return true
}

func main() {
	g := &Game{}
	g.Init()
	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		if rl.IsKeyDown(rl.KeyA) {
			g.TryToggleSquare(rl.GetMousePosition(), rl.GetTime())
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)
		for _, square := range g.squares {
			rl.DrawRectangle(int32(square.X)-squareSize/2, int32(square.Y)-squareSize/2, squareSize, squareSize, rl.RayWhite)
		}
		rl.EndDrawing()
	}
}
