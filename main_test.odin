package game

import "core:testing"
import rl "vendor:raylib"

@(test)
test_add_shape :: proc(t: ^testing.T) {
	game: Game
	defer delete(game.shapes)

	square_pos := rl.Vector2{12, 34}
	circle_pos := rl.Vector2{56, 78}
	add_shape(&game, .Square, square_pos)
	add_shape(&game, .Circle, circle_pos)

	if !testing.expect_value(t, len(game.shapes), 2) {
		return
	}
	testing.expect_value(t, game.shapes[0].shape_type, Shape_Type.Square)
	testing.expect_value(t, game.shapes[0].team, Team.Player)
	testing.expect_value(t, game.shapes[0].pos, square_pos)
	testing.expect_value(t, game.shapes[1].shape_type, Shape_Type.Circle)
	testing.expect_value(t, game.shapes[1].team, Team.AI)
	testing.expect_value(t, game.shapes[1].pos, circle_pos)
}

@(test)
test_team_color :: proc(t: ^testing.T) {
	testing.expect_value(t, TEAM_COLOR[.Player], rl.GREEN)
	testing.expect_value(t, TEAM_COLOR[.AI], rl.BLUE)
}

@(test)
test_shape_geometry :: proc(t: ^testing.T) {
	position := rl.Vector2{20, 20}
	corner := rl.Vector2{29, 29}
	testing.expect(
		t,
		shape_contains_point(Shape{shape_type = .Square, pos = position}, corner),
		"square should contain a point near its corner",
	)
	testing.expect(
		t,
		!shape_contains_point(Shape{shape_type = .Circle, pos = position}, corner),
		"circle should not contain a point outside its radius",
	)
}

@(test)
test_click_selects_topmost_shape_and_empty_click_clears_selection :: proc(t: ^testing.T) {
	position := rl.Vector2{25, 25}
	game := Game {
		shapes = make([dynamic]Shape, 2),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = position}
	game.shapes[1] = {shape_type = .Circle, team = .Player, pos = position}

	release_selection(&game, position, position, 1)
	testing.expect(t, !game.shapes[0].selected)
	testing.expect(t, game.shapes[1].selected)
	testing.expect(t, game.has_last_click)
	testing.expect(t, !game.selecting)

	empty := rl.Vector2{100, 100}
	release_selection(&game, empty, empty, 2)
	testing.expect(t, !game.shapes[0].selected)
	testing.expect(t, !game.shapes[1].selected)
	testing.expect(t, !game.has_last_click)
}

@(test)
test_enemy_shapes_cannot_be_selected_moved_or_deleted :: proc(t: ^testing.T) {
	position := rl.Vector2{20, 20}
	game := Game {
		shapes = make([dynamic]Shape, 3),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = position}
	game.shapes[1] = {shape_type = .Circle, team = .AI, pos = position}
	game.shapes[2] = {shape_type = .Square, team = .AI, pos = {60, 60}}

	release_selection(&game, position, position, 1)
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, !game.shapes[1].selected)

	select_shapes(&game, {x = 0, y = 0, width = 100, height = 100})
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, !game.shapes[1].selected)
	testing.expect(t, !game.shapes[2].selected)

	select_all_shapes_of_type(&game, .Square)
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, !game.shapes[2].selected)

	// Even inconsistent selection state must not allow commands to affect an enemy.
	game.shapes[1].selected = true
	game.shapes[2].selected = true
	target := rl.Vector2{100, 100}
	move_selected_shapes(&game, target)
	advance_shapes(&game, 0.25)
	testing.expect_value(t, game.shapes[1].target, rl.Vector2{})
	testing.expect_value(t, game.shapes[1].pos, position)
	testing.expect_value(t, game.shapes[2].target, rl.Vector2{})
	testing.expect_value(t, game.shapes[2].pos, rl.Vector2{60, 60})

	delete_selected_shapes(&game)
	if !testing.expect_value(t, len(game.shapes), 2) {
		return
	}
	testing.expect_value(t, game.shapes[0].team, Team.AI)
	testing.expect_value(t, game.shapes[1].team, Team.AI)
}

@(test)
test_double_click_selects_all_shapes_of_same_type :: proc(t: ^testing.T) {
	position := rl.Vector2{20, 20}
	game := Game {
		shapes = make([dynamic]Shape, 3),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = position}
	game.shapes[1] = {shape_type = .Square, team = .Player, pos = {60, 60}}
	game.shapes[2] = {shape_type = .Circle, team = .AI, pos = {100, 100}}

	release_selection(&game, position, position, 1)
	release_selection(&game, position, position, 1 + DOUBLE_CLICK_INTERVAL / 2)
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, game.shapes[1].selected)
	testing.expect(t, !game.shapes[2].selected)
	testing.expect(t, !game.has_last_click)

	release_selection(&game, position, position, 2)
	testing.expect(t, game.has_last_click)
	testing.expect(t, !game.shapes[1].selected)
}

@(test)
test_drag_selects_intersecting_shapes_and_resets_click :: proc(t: ^testing.T) {
	game := Game {
		shapes = make([dynamic]Shape, 3),
		has_last_click = true,
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = {20, 20}}
	game.shapes[1] = {shape_type = .Circle, team = .Player, pos = {45, 45}}
	game.shapes[2] = {shape_type = .Square, team = .Player, pos = {100, 100}, selected = true}

	release_selection(&game, {50, 50}, {10, 10}, 1)
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, game.shapes[1].selected)
	testing.expect(t, !game.shapes[2].selected)
	testing.expect(t, !game.has_last_click)
	testing.expect(t, !game.selecting)
}

@(test)
test_movement_at_threshold_remains_click :: proc(t: ^testing.T) {
	game := Game {
		shapes = make([dynamic]Shape, 1),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = {10, 10}}

	release_selection(&game, {10, 10}, {15, 10}, 1)
	testing.expect(t, game.shapes[0].selected)
	testing.expect(t, game.has_last_click)
}

@(test)
test_delete_selected_shapes :: proc(t: ^testing.T) {
	kept := Shape{shape_type = .Circle, team = .Player, pos = {30, 30}}
	game := Game {
		shapes = make([dynamic]Shape, 3),
		has_last_click = true,
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, selected = true}
	game.shapes[1] = kept
	game.shapes[2] = {shape_type = .Circle, team = .Player, selected = true}

	delete_selected_shapes(&game)
	if !testing.expect_value(t, len(game.shapes), 1) {
		return
	}
	testing.expect_value(t, game.shapes[0], kept)
	testing.expect(t, !game.has_last_click)
}

@(test)
test_move_selected_shapes_sets_individual_targets :: proc(t: ^testing.T) {
	game := Game {
		shapes = make([dynamic]Shape, 3),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, pos = {10, 0}, selected = true}
	game.shapes[1] = {shape_type = .Circle, team = .Player, pos = {20, 0}, selected = true}
	game.shapes[2] = {shape_type = .Square, team = .Player, pos = {30, 0}}

	target := rl.Vector2{100, 50}
	move_selected_shapes(&game, target)

	for i in 0 ..< 2 {
		testing.expect_value(t, game.shapes[i].target, target)
		testing.expect(t, game.shapes[i].is_moving)
		testing.expect_value(t, game.shapes[i].pos.x, f32((i + 1) * 10))
	}
	testing.expect_value(t, game.shapes[2].target, rl.Vector2{})
	testing.expect(t, !game.shapes[2].is_moving)

	game.shapes[0].selected = false
	game.shapes[1].selected = false
	move_selected_shapes(&game, {200, 0})
	testing.expect_value(t, game.shapes[0].target, target)
	testing.expect_value(t, game.shapes[1].target, target)

	game.shapes[1].selected = true
	new_target := rl.Vector2{200, 0}
	move_selected_shapes(&game, new_target)
	testing.expect_value(t, game.shapes[0].target, target)
	testing.expect_value(t, game.shapes[1].target, new_target)
}

@(test)
test_advance_shapes_uses_frame_time_and_stops_at_target :: proc(t: ^testing.T) {
	game := Game {
		shapes = make([dynamic]Shape, 1),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Square, team = .Player, target = {200, 0}, is_moving = true}

	advance_shapes(&game, 0.25)
	testing.expect_value(t, game.shapes[0].pos, rl.Vector2{50, 0})
	testing.expect(t, game.shapes[0].is_moving)

	advance_shapes(&game, 0.5)
	testing.expect_value(t, game.shapes[0].pos, rl.Vector2{150, 0})
	testing.expect(t, game.shapes[0].is_moving)

	advance_shapes(&game, 1)
	testing.expect_value(t, game.shapes[0].pos, game.shapes[0].target)
	testing.expect(t, !game.shapes[0].is_moving)

	advance_shapes(&game, 1)
	testing.expect_value(t, game.shapes[0].pos, game.shapes[0].target)
}

@(test)
test_advance_shapes_has_same_speed_diagonally :: proc(t: ^testing.T) {
	game := Game {
		shapes = make([dynamic]Shape, 1),
	}
	defer delete(game.shapes)
	game.shapes[0] = {shape_type = .Circle, team = .Player, target = {72, 96}, is_moving = true}

	advance_shapes(&game, 0.5)
	testing.expect_value(t, game.shapes[0].pos, rl.Vector2{36, 48})
	testing.expect(t, game.shapes[0].is_moving)
}

@(test)
test_move_selected_shape_to_its_current_position_stops_it :: proc(t: ^testing.T) {
	position := rl.Vector2{10, 20}
	game := Game {
		shapes = make([dynamic]Shape, 1),
	}
	defer delete(game.shapes)
	game.shapes[0] = {
		team = .Player,
		pos = position,
		target = {100, 0},
		selected = true,
		is_moving = true,
	}

	move_selected_shapes(&game, position)
	testing.expect_value(t, game.shapes[0].pos, position)
	testing.expect_value(t, game.shapes[0].target, position)
	testing.expect(t, !game.shapes[0].is_moving)
}

@(test)
test_selection_rect_for_click_uses_cursor_point :: proc(t: ^testing.T) {
	position := rl.Vector2{12, 34}
	rect := selection_rect(position, position)
	testing.expect_value(t, rect.x, position.x)
	testing.expect_value(t, rect.y, position.y)
	testing.expect_value(t, rect.width, f32(1))
	testing.expect_value(t, rect.height, f32(1))
}

@(test)
test_double_click_tracks_moving_shape :: proc(t: ^testing.T) {
	game := Game{shapes = make([dynamic]Shape, 2)}
	defer delete(game.shapes)
	game.shapes[0] = {
		team = .Player,
		pos = {20, 20},
		target = {200, 20},
		is_moving = true,
	}
	game.shapes[1] = {team = .Player, pos = {100, 100}}

	select_clicked_shape(&game, game.shapes[0].pos, 1)
	advance_shapes(&game, 0.1)
	select_clicked_shape(&game, game.shapes[0].pos, 1.1)
	testing.expect(t, game.shapes[1].selected, "double click should track the same moving shape")
}

@(test)
test_clicking_different_shape_at_same_position_is_not_double_click :: proc(t: ^testing.T) {
	game := Game{shapes = make([dynamic]Shape, 2)}
	defer delete(game.shapes)
	game.shapes[0] = {team = .Player, pos = {20, 20}}
	game.shapes[1] = {team = .Player, pos = {100, 100}}

	select_clicked_shape(&game, {20, 20}, 1)
	game.shapes[0].pos = {60, 60}
	game.shapes[1].pos = {20, 20}
	select_clicked_shape(&game, {20, 20}, 1.1)
	testing.expect(t, !game.shapes[0].selected)
	testing.expect(t, game.shapes[1].selected)
	testing.expect(t, game.has_last_click)
}

@(test)
test_delete_selected_shapes_preserves_survivor_order :: proc(t: ^testing.T) {
	game := Game{shapes = make([dynamic]Shape, 6)}
	defer delete(game.shapes)
	for &shape, i in game.shapes {
		shape = {team = .Player, pos = {f32(i), 0}, selected = i % 2 == 0}
	}
	game.shapes[2].team = .AI
	kept := [4]Shape{game.shapes[1], game.shapes[2], game.shapes[3], game.shapes[5]}

	delete_selected_shapes(&game)
	if !testing.expect_value(t, len(game.shapes), len(kept)) {
		return
	}
	for shape, i in game.shapes {
		testing.expect_value(t, shape, kept[i])
	}
	clear_selection(&game)
	delete_selected_shapes(&game)
	testing.expect_value(t, len(game.shapes), len(kept))

	for &shape in game.shapes {
		shape.team = .Player
		shape.selected = true
	}
	delete_selected_shapes(&game)
	testing.expect_value(t, len(game.shapes), 0)
	delete_selected_shapes(&game)
	testing.expect_value(t, len(game.shapes), 0)
}

release_selection :: proc(game: ^Game, start, end: rl.Vector2, now: f64) {
	game.selection_start = start
	game.selecting = true
	finish_selection(game, end, now)
}
