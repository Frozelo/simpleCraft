package main

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestToggleSquare(t *testing.T) {
	g := &Game{}
	position := rl.Vector2{X: 12, Y: 34}

	g.ToggleSquare(position)
	if len(g.squares) != 1 || g.squares[0] != position {
		t.Fatalf("squares = %#v, want [%#v]", g.squares, position)
	}

	g.ToggleSquare(position)
	if len(g.squares) != 0 {
		t.Fatalf("squares = %#v, want empty", g.squares)
	}
}

func TestTryToggleSquareRespectsCooldown(t *testing.T) {
	g := &Game{}
	position := rl.Vector2{X: 12, Y: 34}

	if !g.TryToggleSquare(position, 1) {
		t.Fatal("first toggle was ignored")
	}
	if g.TryToggleSquare(position, 1+squareToggleCooldown/2) {
		t.Fatal("toggle during cooldown succeeded")
	}
	if !g.TryToggleSquare(position, 1+squareToggleCooldown) {
		t.Fatal("toggle after cooldown was ignored")
	}
	if len(g.squares) != 0 {
		t.Fatalf("squares = %#v, want empty", g.squares)
	}
}
