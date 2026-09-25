# SimpleCraft Sim

A minimalist RTS project written in Go for experimenting with game simulation, unit control, pathfinding, and performance optimization.

The main goal is not to build a full StarCraft II clone, but to explore the technical principles behind a responsive and well-engineered RTS:

- precise and predictable input;
- fixed-step simulation;
- stable update loop;
- controlling large numbers of units;
- pathfinding;
- local avoidance;
- formations;
- collision and separation;
- stable frametimes;
- minimal coupling between game logic and rendering.

## Idea

At the first stage, the game is simply a field containing squares that represent units.

There are intentionally no:

- textures;
- animations;
- complex UI;
- buildings;
- economy;
- AI;
- multiplayer;
- full-featured game engine.

The primary goal is to build a high-quality RTS simulation layer.

Example:

```text
┌──────────────────────────────────────────┐
│                                          │
│     ■       ■ ■                          │
│                                          │
│                 ■                        │
│                              ■           │
│                                          │
│         ■                                │
│                                          │
└──────────────────────────────────────────┘
```

The player should be able to select units and issue movement commands.

Eventually, the simulation should support hundreds of units running simultaneously.

---

# Tech Stack

## Language

Go.

The core game logic should be written in pure Go and remain independent from the renderer.

## Rendering / Input

The initial implementation uses:

```text
raylib-go
```

Raylib acts only as a thin platform layer responsible for:

```text
window
input
mouse
keyboard
drawing
```

It should not contain the actual game logic.

The architecture should allow replacing raylib in the future with:

- Ebitengine;
- SDL;
- OpenGL;
- Vulkan;
- another renderer;

without significantly modifying the simulation core.

---

# Architecture

The main architectural principle is:

```text
Input
  │
  ▼
Commands
  │
  ▼
Simulation
  │
  ▼
World State
  │
  ▼
Renderer
```

The simulation is the source of truth.

The renderer only visualizes the current state of the world.

For example:

```text
Mouse Click
    │
    ▼
MoveCommand
    │
    ▼
Simulation
    │
    ▼
Unit Position
    │
    ▼
Renderer
```

The renderer should never directly modify unit positions.

---

# Simulation

Basic world structure:

```go
type World struct {
    Units []Unit
}
```

A unit:

```go
type Unit struct {
    ID int

    X float32
    Y float32

    TargetX float32
    TargetY float32

    Speed float32
}
```

World update:

```go
func (w *World) Update(dt float32) {
    for i := range w.Units {
        w.Units[i].Update(dt)
    }
}
```

The basic flow is:

```text
update state
    ↓
render state
```

Game logic should never live inside the rendering code.

---

# Game Loop

Initial game loop:

```go
for !rl.WindowShouldClose() {
    update()
    draw()
}
```

Conceptually:

```text
┌──────────────────┐
│      INPUT       │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│      UPDATE      │
│                  │
│   simulation     │
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│       DRAW       │
│                  │
│     renderer     │
└──────────────────┘
```

Eventually, the simulation should use a fixed timestep.

For example:

```text
Simulation: 60 ticks/sec
Rendering:   independent from simulation
```

This makes the simulation more predictable and provides a foundation for:

- deterministic simulation;
- replays;
- networking;
- automated testing.

---

# Project Structure

Planned structure:

```text
rts-sim/
├── cmd/
│   └── sandbox/
│       └── main.go
│
├── internal/
│   ├── sim/
│   │   ├── world.go
│   │   ├── unit.go
│   │   └── command.go
│   │
│   ├── input/
│   │   └── input.go
│   │
│   └── render/
│       └── render.go
│
├── go.mod
└── README.md
```

During the earliest stage, the project can remain as simple as:

```text
main.go
go.mod
```

The code can be separated into packages once the project grows enough to justify it.

---

# First MVP

The first technical MVP:

```text
1. Open a window
2. Draw a square
3. Create a Unit
4. Render the Unit
5. Select the Unit with the mouse
6. Use RMB to set a target
7. Move the Unit toward the target
8. Add multiple Units
9. Add box selection
10. Issue commands to a group of Units
```

After this point, the project starts moving into actual RTS-specific engineering.

---

# Roadmap

## Phase 0 — Sandbox

Create the window and basic game loop.

```text
window
update()
draw()
FPS
```

## Phase 1 — Unit

A square becomes a unit.

```text
Unit
 ├── position
 ├── speed
 └── target
```

Right-clicking sets the destination.

---

## Phase 2 — Multiple Units

Add:

```text
10 units
50 units
100 units
```

All units are updated by the simulation.

---

## Phase 3 — Selection

Implement:

```text
LMB
 └── select unit

LMB drag
 └── box selection
```

---

## Phase 4 — Commands

Input should not directly modify a unit.

Instead of:

```go
unit.TargetX = mouseX
```

use a command:

```go
type Command struct {
    UnitIDs []int

    TargetX float32
    TargetY float32
}
```

Flow:

```text
Input
  ↓
Command
  ↓
Simulation
```

---

## Phase 5 — Group Movement

When multiple units are sent to the same position, they should not overlap.

The simulation should eventually support:

```text
target
  ↓
group destination
  ↓
individual unit destinations
```

---

## Phase 6 — Spatial Partitioning

Avoid comparing every unit against every other unit:

```text
O(n²)
```

Instead, use a spatial grid or spatial hash:

```text
┌─────┬─────┬─────┐
│     │ ■ ■ │     │
├─────┼─────┼─────┤
│  ■  │ ■   │     │
├─────┼─────┼─────┤
│     │     │ ■   │
└─────┴─────┴─────┘
```

Each unit only needs to inspect nearby cells.

---

## Phase 7 — Separation

Units should maintain a minimum distance from each other.

```text
desired velocity
      +
separation
      =
final velocity
```

---

## Phase 8 — Map

Add a tile/grid-based map.

Example:

```text
. . . . . . .
. . X X . . .
. . X X . . .
. . . . . . .
```

Where:

```text
. = walkable
X = blocked
```

---

## Phase 9 — Pathfinding

Initial implementation:

```text
A*
```

Flow:

```text
Unit
 ↓
Target
 ↓
A*
 ↓
Path
 ↓
Waypoints
 ↓
Movement
```

---

## Phase 10 — Local Avoidance

Pathfinding handles the global route.

Local avoidance handles short-range movement and interactions between units.

```text
Global Path
      +
Local Avoidance
      +
Separation
      =
Final Movement
```

---

# Performance Principles

For an RTS, average performance is not enough.

Stable frametimes are especially important.

Bad:

```text
8ms
9ms
8ms
35ms
8ms
```

Good:

```text
10ms
10ms
10ms
10ms
10ms
```

The simulation hot path should therefore remain simple.

Prefer:

```text
slices
structs
preallocated buffers
buffer reuse
simple loops
```

Avoid unnecessary:

```text
allocation per unit per tick
goroutine per unit
channels between every subsystem
complex interface hierarchies
reflection
```

---

# Concurrency

Go provides a powerful concurrency model, but that does not mean every unit should run inside its own goroutine.

Avoid:

```go
for _, unit := range units {
    go unit.Update()
}
```

The simulation should first and foremost remain:

```text
simple
predictable
deterministic
cache-friendly
```

Parallelism should only be introduced later where profiling shows a real benefit.

---

# ECS

The project will not use an ECS architecture initially.

Regular Go structures are sufficient:

```go
type World struct {
    Units []Unit
}
```

This keeps the implementation simple and allows the actual requirements of the project to emerge naturally.

If necessary, the simulation can later move toward a more data-oriented representation:

```go
positions  []Vec2
velocities []Vec2
health     []int
states     []State
```

---

# Long-Term Experiments

Once the core simulation becomes stable, the project can experiment with:

- combat;
- attack move;
- projectiles;
- health;
- buildings;
- workers;
- resources;
- unit production;
- fog of war;
- enemy AI;
- formations;
- flow fields;
- hierarchical pathfinding;
- replays;
- deterministic multiplayer;
- lockstep networking.

---

# Core Goal

The main goal is to achieve the following interaction:

```text
select units
      ↓
give command
      ↓
immediate response
      ↓
predictable movement
      ↓
stable simulation
```

The priority is not the number of gameplay mechanics.

The priority is the quality of the fundamental unit-control experience.

The first significant milestone is:

> 100–200 square units can be selected and ordered to move to a destination while moving smoothly and predictably, avoiding obstacles and each other without noticeable input delay.
