package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestToggleSquare(t *testing.T) {
	g := &Game{}
	position := rl.Vector2{X: 12, Y: 34}

	g.toggleSquare(position, 1)
	if len(g.squares) != 1 || g.squares[0].Pos != position {
		t.Fatalf("squares = %#v, want [%#v]", g.squares, position)
	}

	g.toggleSquare(position, 1+squareToggleCooldown)
	if len(g.squares) != 0 {
		t.Fatalf("squares = %#v, want empty", g.squares)
	}
}

func TestSelectSquares(t *testing.T) {
	g := &Game{squares: []Square{
		{Pos: rl.Vector2{X: 25, Y: 10}},
		{Pos: rl.Vector2{X: 50, Y: 50}, Selected: true},
	}}

	g.selectSquares(rl.Rectangle{X: 0, Y: 0, Width: 20, Height: 20})

	if !g.squares[0].Selected || g.squares[1].Selected {
		t.Fatalf("selected = %#v, want [true false]", g.squares)
	}
}

func TestSelectionRectForClickUsesCursorPoint(t *testing.T) {
	position := rl.Vector2{X: 12, Y: 34}
	rect := selectionRect(position, position)

	if rect.X != position.X || rect.Y != position.Y || rect.Width != 1 || rect.Height != 1 {
		t.Fatalf("selectionRect() = %#v, want a 1x1 rectangle at %#v", rect, position)
	}
}

func TestTryToggleSquareRespectsCooldown(t *testing.T) {
	g := &Game{}
	position := rl.Vector2{X: 12, Y: 34}

	if !g.toggleSquare(position, 1) {
		t.Fatal("first toggle was ignored")
	}
	if g.toggleSquare(position, 1+squareToggleCooldown/2) {
		t.Fatal("toggle during cooldown succeeded")
	}
	if !g.toggleSquare(position, 1+squareToggleCooldown) {
		t.Fatal("toggle after cooldown was ignored")
	}
	if len(g.squares) != 0 {
		t.Fatalf("squares = %#v, want empty", g.squares)
	}
}
