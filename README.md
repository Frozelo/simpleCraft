# SimpleCraft Sim

A minimalist RTS project written in Odin for experimenting with game simulation, unit control, pathfinding, and performance optimization.

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

Odin.

The core game logic should be written in Odin and remain independent from the renderer.

## Rendering / Input

The initial implementation uses Odin's `vendor:raylib` bindings.

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

- SDL;
- OpenGL;
- Vulkan;
- another renderer;

without significantly modifying the simulation core.

Build and test:

```text
odin run .
odin test .
```

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

```odin
World :: struct {
    units: [dynamic]Unit,
}
```

A unit:

```odin
Unit :: struct {
    id: int,

    position: [2]f32,
    target:   [2]f32,
    speed:    f32,
}
```

World update:

```odin
update :: proc(world: ^World, dt: f32) {
    for &unit in world.units {
        update_unit(&unit, dt)
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

```odin
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
├── src/
│   ├── main.odin
│   ├── sim/
│   │   ├── world.odin
│   │   ├── unit.odin
│   │   └── command.odin
│   ├── input/
│   │   └── input.odin
│   └── render/
│       └── render.odin
└── README.md
```

During the earliest stage, the project can remain as simple as:

```text
main.odin
main_test.odin
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

```odin
unit.target.x = mouse_x
```

use a command:

```odin
Command :: struct {
    unit_ids: [dynamic]int,
    target:   [2]f32,
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
dynamic arrays
structs
preallocated buffers
buffer reuse
simple loops
```

Avoid unnecessary:

```text
allocation per unit per tick
thread per unit
virtual calls between every subsystem
reflection
```

---

# Concurrency

Odin makes data-oriented code the default, but that does not mean every unit should run on its own thread.

Avoid:

```odin
for &unit in units {
    thread.create_and_start_with_poly_data(&unit, update_unit)
}
```

The simulation should first and foremost remain:

```text
simple
predictable
deterministic
cache-friendly
```

Parallelism should only be introduced later, through `core:thread`, where profiling shows a real benefit.

---

# ECS

The project will not use an ECS architecture initially.

Regular Odin structures are sufficient:

```odin
World :: struct {
    units: [dynamic]Unit,
}
```

This keeps the implementation simple and allows the actual requirements of the project to emerge naturally.

If necessary, the simulation can later move toward a structure-of-arrays layout, which Odin supports directly:

```odin
units: #soa[dynamic]Unit
```

That stores each field in its own array while the update loop still iterates units.

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
