package game

import "core:fmt"
import "core:log"
import rl "vendor:raylib"

WINDOW_WIDTH :: 1280
WINDOW_HEIGHT :: 720
WINDOW_TITLE :: "SimpleCraft Sim"
TARGET_FPS :: 60
DOUBLE_CLICK_INTERVAL :: 0.35
SELECTION_DRAG_THRESHOLD :: 5.0
NOTICE_DURATION :: 2.0
NOTICE_FADE :: 1.2
PLACEMENT_BLOCKED_TEXT :: "Cannot place shape here"

Shape_Type :: enum u16 {
	Square,
	Circle,
}

Team :: enum u8 {
	Player,
	AI,
}

Shape_Geometry :: enum u8 {
	Square,
	Circle,
}

Shape_Definition :: struct {
	name:         string,
	geometry:     Shape_Geometry,
	size:         f32,
	move_speed:   f32,
	default_team: Team,
	spawn_key:    rl.KeyboardKey,
}

TEAM_COLOR := [Team]rl.Color {
	.Player = rl.GREEN,
	.AI     = rl.BLUE,
}

SHAPE_DEFINITIONS := [Shape_Type]Shape_Definition {
	.Square = {
		name = "Squares",
		geometry = .Square,
		size = 20,
		move_speed = 200,
		default_team = .Player,
		spawn_key = .S,
	},
	.Circle = {
		name = "Circles",
		geometry = .Circle,
		size = 20,
		move_speed = 120,
		default_team = .AI,
		spawn_key = .C,
	},
}

Shape :: struct {
	shape_type: Shape_Type,
	team:       Team,
	pos:        rl.Vector2,
	target:     rl.Vector2,
	selected:   bool,
	is_moving:  bool,
}

Game :: struct {
	shapes: [dynamic]Shape,

	selection_start: rl.Vector2,
	selecting:       bool,

	last_click_at:          f64,
	last_click_shape_index: int,
	has_last_click:         bool,

	notice_remaining: f32,
}

main :: proc() {
	game: Game
	defer delete(game.shapes)
	run(&game)
}

run :: proc(game: ^Game) {
	context.logger = log.create_console_logger()
	defer log.destroy_console_logger(context.logger)

	rl.SetConfigFlags({.WINDOW_RESIZABLE})
	rl.InitWindow(WINDOW_WIDTH, WINDOW_HEIGHT, WINDOW_TITLE)
	defer rl.CloseWindow()

	rl.SetTargetFPS(TARGET_FPS)

	for !rl.WindowShouldClose() {
		update(game, rl.GetFrameTime())

		rl.BeginDrawing()
		rl.ClearBackground(rl.BLACK)
		draw(game)
		rl.EndDrawing()

		free_all(context.temp_allocator)
	}
}

update :: proc(game: ^Game, dt: f32) {
	for shape_type in Shape_Type {
		definition := SHAPE_DEFINITIONS[shape_type]
		if definition.spawn_key != .KEY_NULL && rl.IsKeyPressed(definition.spawn_key) {
			add_shape(game, shape_type, rl.GetMousePosition())
		}
	}

	if rl.IsKeyPressed(.BACKSPACE) {
		delete_selected_shapes(game)
	}

	// Start tracking a selection when the left mouse button is pressed.
	if rl.IsMouseButtonPressed(.LEFT) {
		game.selecting = true
		game.selection_start = rl.GetMousePosition()
	}

	if rl.IsMouseButtonReleased(.LEFT) && game.selecting {
		finish_selection(game, rl.GetMousePosition(), rl.GetTime())
	}

	if rl.IsMouseButtonPressed(.RIGHT) {
		move_selected_shapes(game, rl.GetMousePosition())
	}
	advance_shapes(game, dt)
	advance_notice(game, dt)
}

finish_selection :: proc(game: ^Game, end: rl.Vector2, now: f64) {
	dx := end.x - game.selection_start.x
	dy := end.y - game.selection_start.y
	// Compare squared distances so the threshold does not require a square root.
	if dx * dx + dy * dy > SELECTION_DRAG_THRESHOLD * SELECTION_DRAG_THRESHOLD {
		select_shapes(game, selection_rect(game.selection_start, end))
		game.has_last_click = false
	} else {
		select_clicked_shape(game, end, now)
	}
	game.selecting = false
}

select_clicked_shape :: proc(game: ^Game, position: rl.Vector2, now: f64) {
	index := shape_at(game, position)
	if index < 0 {
		clear_selection(game)
		game.has_last_click = false
		return
	}

	shape := game.shapes[index]
	is_double_click := game.has_last_click &&
		index == game.last_click_shape_index &&
		now >= game.last_click_at &&
		now - game.last_click_at <= DOUBLE_CLICK_INTERVAL
	if is_double_click {
		select_all_shapes_of_type(game, shape.shape_type)
		game.has_last_click = false
		return
	}

	select_shape(game, index)
	game.last_click_at = now
	game.last_click_shape_index = index
	game.has_last_click = true
}

draw :: proc(game: ^Game) {
	selection: rl.Rectangle
	if game.selecting {
		selection = selection_rect(game.selection_start, rl.GetMousePosition())
		rl.DrawRectangleRec(selection, rl.Fade(rl.SKYBLUE, 0.2))
	}

	selected := 0
	pre_selected := 0
	player_count := 0
	enemy_count := 0
	counts: [Shape_Type]int
	for shape in game.shapes {
		counts[shape.shape_type] += 1

		color := TEAM_COLOR[shape.team]
		if shape.team == .Player {
			player_count += 1
		} else {
			enemy_count += 1
		}
		if shape.selected {
			selected += 1
		}
		preview := game.selecting && shape.team == .Player && shape_intersects_rect(shape, selection)
		if preview {
			pre_selected += 1
		}

		draw_shape(shape, color)
		if preview {
			draw_shape_outline(shape, rl.ORANGE)
		} else if shape.team == .Player && shape.selected {
			draw_shape_outline(shape, rl.GOLD)
		}
	}

	if game.selecting {
		rl.DrawRectangleLinesEx(selection, 1, rl.SKYBLUE)
	}

	for shape_type in Shape_Type {
		definition := SHAPE_DEFINITIONS[shape_type]
		rl.DrawText(
			fmt.ctprintf("%s: %d", definition.name, counts[shape_type]),
			10,
			i32(50 + 20 * int(shape_type)),
			20,
			rl.LIGHTGRAY,
		)
	}

	rl.DrawText(
		fmt.ctprintf("Player: %d", player_count),
		10,
		i32(50 + 20 * len(Shape_Type)),
		20,
		TEAM_COLOR[.Player],
	)
	rl.DrawText(
		fmt.ctprintf("Enemy: %d", enemy_count),
		10,
		i32(70 + 20 * len(Shape_Type)),
		20,
		TEAM_COLOR[.AI],
	)
	rl.DrawText(fmt.ctprintf("Selected: %d", selected), 10, 10, 20, rl.GREEN)
	rl.DrawText(fmt.ctprintf("PreSelected: %d", pre_selected), 10, 30, 20, rl.GRAY)
	draw_notice(game)
}

draw_notice :: proc(game: ^Game) {
	alpha := notice_alpha(game.notice_remaining)
	if alpha <= 0 {
		return
	}

	text := fmt.ctprintf("%s", PLACEMENT_BLOCKED_TEXT)
	font_size: i32 = 20
	width := rl.MeasureText(text, font_size)
	x := (rl.GetScreenWidth() - width) / 2
	y := rl.GetScreenHeight() - font_size - 24
	rl.DrawText(text, x, y, font_size, rl.Fade(rl.RED, alpha))
}

notice_alpha :: proc(remaining: f32) -> f32 {
	if remaining <= 0 {
		return 0
	}
	if remaining >= NOTICE_FADE {
		return 1
	}
	return remaining / NOTICE_FADE
}

advance_notice :: proc(game: ^Game, dt: f32) {
	if dt <= 0 || game.notice_remaining <= 0 {
		return
	}
	game.notice_remaining = max(0, game.notice_remaining - dt)
}

show_placement_notice :: proc(game: ^Game) {
	game.notice_remaining = NOTICE_DURATION
}

add_shape :: proc(game: ^Game, shape_type: Shape_Type, position: rl.Vector2) {
	definition := SHAPE_DEFINITIONS[shape_type]
	shape := Shape {
		shape_type = shape_type,
		team = definition.default_team,
		pos = position,
	}
	if overlaps_any_shape(game, shape) {
		show_placement_notice(game)
		log.infof(
			"shape placement blocked type=%v x=%v y=%v",
			shape_type,
			position.x,
			position.y,
		)
		return
	}

	append(&game.shapes, shape)
	log.infof(
		"shape created type=%v team=%v x=%v y=%v",
		shape_type,
		definition.default_team,
		position.x,
		position.y,
	)
}

overlaps_any_shape :: proc(game: ^Game, candidate: Shape) -> bool {
	for existing in game.shapes {
		if shapes_overlap(candidate, existing) {
			return true
		}
	}
	return false
}

delete_selected_shapes :: proc(game: ^Game) {
	kept := 0
	for shape, i in game.shapes {
		if shape.team == .Player && shape.selected {
			continue
		}
		if kept != i {
			game.shapes[kept] = shape
		}
		kept += 1
	}
	resize(&game.shapes, kept)
	// Compaction can change indices used to recognize a double click.
	game.has_last_click = false
}

select_shapes :: proc(game: ^Game, rect: rl.Rectangle) {
	for &shape in game.shapes {
		shape.selected = shape.team == .Player && shape_intersects_rect(shape, rect)
	}
}

shape_at :: proc(game: ^Game, position: rl.Vector2) -> int {
	for i := len(game.shapes) - 1; i >= 0; i -= 1 {
		shape := game.shapes[i]
		if shape.team == .Player && shape_contains_point(shape, position) {
			return i
		}
	}
	return -1
}

select_shape :: proc(game: ^Game, index: int) {
	for &shape in game.shapes {
		shape.selected = false
	}
	game.shapes[index].selected = game.shapes[index].team == .Player
}

select_all_shapes_of_type :: proc(game: ^Game, shape_type: Shape_Type) {
	for &shape in game.shapes {
		shape.selected = shape.team == .Player && shape.shape_type == shape_type
	}
}

clear_selection :: proc(game: ^Game) {
	for &shape in game.shapes {
		shape.selected = false
	}
}

move_selected_shapes :: proc(game: ^Game, target: rl.Vector2) {
	for &shape in game.shapes {
		if shape.team == .Player && shape.selected {
			shape.target = target
			shape.is_moving = shape.pos != target
		}
	}
}

advance_shapes :: proc(game: ^Game, dt: f32) {
	if dt <= 0 {
		return
	}

	for &shape in game.shapes {
		if shape.team != .Player || !shape.is_moving {
			continue
		}

		step := SHAPE_DEFINITIONS[shape.shape_type].move_speed * dt

		dx := shape.target.x - shape.pos.x
		dy := shape.target.y - shape.pos.y
		distance := rl.Vector2Distance(shape.pos, shape.target)
		if distance <= step {
			shape.pos = shape.target
			shape.is_moving = false
			continue
		}

		shape.pos.x += dx / distance * step
		shape.pos.y += dy / distance * step
	}
}

draw_shape :: proc(shape: Shape, color: rl.Color) {
	definition := SHAPE_DEFINITIONS[shape.shape_type]
	switch definition.geometry {
	case .Square:
		rl.DrawRectangleRec(square_rect(shape.pos, definition.size), color)
	case .Circle:
		rl.DrawCircleV(shape.pos, definition.size / 2, color)
	}
}

draw_shape_outline :: proc(shape: Shape, color: rl.Color) {
	definition := SHAPE_DEFINITIONS[shape.shape_type]
	switch definition.geometry {
	case .Square:
		rl.DrawRectangleLinesEx(square_rect(shape.pos, definition.size + 4), 2, color)
	case .Circle:
		rl.DrawCircleLinesV(shape.pos, definition.size / 2 + 3, color)
	}
}

shape_contains_point :: proc(shape: Shape, point: rl.Vector2) -> bool {
	definition := SHAPE_DEFINITIONS[shape.shape_type]
	switch definition.geometry {
	case .Square:
		return rl.CheckCollisionPointRec(point, square_rect(shape.pos, definition.size))
	case .Circle:
		return rl.CheckCollisionPointCircle(point, shape.pos, definition.size / 2)
	}
	return false
}

shapes_overlap :: proc(a, b: Shape) -> bool {
	a_definition := SHAPE_DEFINITIONS[a.shape_type]
	b_definition := SHAPE_DEFINITIONS[b.shape_type]
	switch a_definition.geometry {
	case .Square:
		a_rect := square_rect(a.pos, a_definition.size)
		switch b_definition.geometry {
		case .Square:
			return rl.CheckCollisionRecs(a_rect, square_rect(b.pos, b_definition.size))
		case .Circle:
			return rl.CheckCollisionCircleRec(b.pos, b_definition.size / 2, a_rect)
		}
	case .Circle:
		a_radius := a_definition.size / 2
		switch b_definition.geometry {
		case .Square:
			return rl.CheckCollisionCircleRec(a.pos, a_radius, square_rect(b.pos, b_definition.size))
		case .Circle:
			return rl.CheckCollisionCircles(a.pos, a_radius, b.pos, b_definition.size / 2)
		}
	}
	return false
}

shape_intersects_rect :: proc(shape: Shape, rect: rl.Rectangle) -> bool {
	definition := SHAPE_DEFINITIONS[shape.shape_type]
	switch definition.geometry {
	case .Square:
		return rl.CheckCollisionRecs(square_rect(shape.pos, definition.size), rect)
	case .Circle:
		return rl.CheckCollisionCircleRec(shape.pos, definition.size / 2, rect)
	}
	return false
}

square_rect :: proc(position: rl.Vector2, size: f32) -> rl.Rectangle {
	return {
		x = position.x - size / 2,
		y = position.y - size / 2,
		width = size,
		height = size,
	}
}

selection_rect :: proc(start, end: rl.Vector2) -> rl.Rectangle {
	if start == end {
		return {x = start.x, y = start.y, width = 1, height = 1}
	}

	s := start
	e := end
	if s.x > e.x {
		s.x, e.x = e.x, s.x
	}
	if s.y > e.y {
		s.y, e.y = e.y, s.y
	}

	return {
		x = s.x,
		y = s.y,
		width = e.x - s.x,
		height = e.y - s.y,
	}
}
