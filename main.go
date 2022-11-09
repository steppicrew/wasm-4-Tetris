package main

import (
	"cart/w4"
	"math/rand"
)

var rnd = rand.New(rand.NewSource(int64(1))).Intn

var game_over bool = false

const row_count = 20
const col_count = 10

const preview_row_count = 2
const preview_col_count = 4

const first_key_pause = 15
const key_pause = 8

const softdrop_pause = 2

var board [row_count * col_count]byte
var preview_board [preview_row_count * preview_col_count]byte

// 1: rechts, 2: hoch, 3: links, 4: runter
var stones = [7][4]byte{
	{1, 1, 1, 0},
	{4, 1, 1, 0},
	{1, 1, 2, 0},
	{1, 2, 3, 0},
	{1, 2, 1, 0},
	{1, 4, 1, 0},
	{1, 2, 4, 1},
}

var stone_rotaion_fixes = [7][3][2]int{
	{{1, 1}, {3, 0}, {1, -2}},
	{{0, 2}, {2, 2}, {2, 0}},
	{{1, 1}, {2, 0}, {1, -1}},
	{{1, 0}, {1, -1}, {0, -1}},
	{{1, 1}, {2, -1}, {0, -1}},
	{{0, 1}, {2, 1}, {1, -1}},
	{{1, 1}, {2, -1}, {1, -1}},
}

var width byte = 160 / col_count
var height byte = 160 / row_count

var initialized bool = false

var this_stone byte
var this_color byte

var next_stone byte
var next_color byte

var x byte = 4
var y byte = 4
var rotation byte

var speed byte = 30
var level byte = 0
var score uint16 = 0

var scale uint
var x_offset int
var y_offset int
var preview_x_offset int
var preview_y_offset int
var remaining_key_pause byte = 0
var remaining_softdrop_pause byte = 0

func _num2str(x uint) string {
	var digit = x % 10
	x = x / 10
	if x > 0 {
		return _num2str(x) + string('0'+digit)
	}
	return string('0' + digit)
}
func num2str(x int) string {
	if x < 0 {
		return "-" + _num2str(uint(-x))
	}
	return _num2str(uint(x))
}

func _sound(freq1, freq2, volume uint, attack, decay, sustain, release, channel, mode byte) {
	w4.Tone(freq1|freq2<<16, uint(attack)<<24|uint(decay)<<16|uint(sustain)|uint(release)<<8, volume, uint(channel)|uint(mode)<<2)
}

func sound_drop() {
	_sound(470, 0, 8, 0, 0, 0, 4, 3, 0)
}

func new_stone(init bool) {
	/*
		var my_x byte = byte(rand.Intn(int(col_count)))
		var my_y byte = 0
		rotation = byte(rand.Intn(4))
		this_stone = byte(rand.Intn(len(stones)))
	*/
	var my_x byte = col_count/2 - 1
	var my_y byte = 0

	rotation = 0

	if init {
		this_stone = byte(rnd(len(stones)))
		this_color = byte(rnd(4)) + 1
	} else {
		this_stone = next_stone
		this_color = next_color
	}

	next_stone = byte(rnd(len(stones)))
	next_color = byte(rnd(4)) + 1
	for i := 0; i < preview_row_count*preview_col_count; i++ {
		preview_board[i] = 0
	}

	if this_stone > 1 {
		my_y = 1
	}

	if _set_stone(preview_board[:], preview_col_count, preview_row_count, next_stone, 0, 0, 0, 0, false, false) == 0 {
		_set_stone(preview_board[:], preview_col_count, preview_row_count, next_stone, next_color, 0, 0, 0, true, false)
	} else {
		_set_stone(preview_board[:], preview_col_count, preview_row_count, next_stone, next_color, 0, 1, 0, true, false)
	}

	if this_stone == 0 {
		my_x--
	}

	for true {
		switch set_stone(my_x, my_y, false, false) {
		case 0:
			x, y = my_x, my_y
			return
		case 1:
			game_over = true
			this_stone = 99
			return
		case 2:
			my_x++
		case 3:
			my_x--
		case 4:
			my_y++
		}
	}

}

// 0: ok, 1: already used, 2: left out, 3: right out, 4: top out
func _set_stone(board []byte, width, height, _stone, _color byte, x, y, rotation byte, apply bool, clear bool) byte {
	if int(this_stone) > len(stones) {
		return 1
	}
	var stone [4]byte = stones[_stone]
	var my_x int8 = int8(x)
	var my_y int8 = int8(y)

	var i = int(my_y)*int(width) + int(my_x)

	if apply {
		if my_x >= 0 && my_x < int8(width) && my_y >= 0 && my_y < int8(height) {
			if clear {
				board[i] = 0
			} else {
				board[i] = _color
			}
		}
	} else {
		if my_x < 0 || my_x >= int8(width) || my_y >= int8(height) {
			return 1
		}
		if board[i] != 0 {
			return 1
		}
	}
	for _, stone_dir := range stone {
		if stone_dir == 0 {
			break
		}
		var dx, dy int8 = 0, 0
		switch stone_dir {
		case 1:
			dx = 1
		case 2:
			dy = -1
		case 3:
			dx = -1
		case 4:
			dy = 1
		}
		switch rotation {
		case 1:
			dx, dy = dy, -dx
		case 2:
			dx, dy = -dx, -dy
		case 3:
			dx, dy = -dy, dx
		}
		my_x += dx
		my_y += dy
		if my_x < 0 {
			return 2
		}
		if my_x >= int8(width) {
			return 3
		}
		if my_y < 0 {
			if apply {
				continue
			}
			return 4
		}
		if my_y >= int8(height) {
			return 1
		}
		var i = int(my_y)*int(width) + int(my_x)
		if apply {
			if clear {
				board[i] = 0
			} else {
				board[i] = _color
			}
		} else {
			if board[i] != 0 {
				return 1
			}
		}
	}
	return 0
}

// 0: ok, 1: already used, 2: left out, 3: right out, 4: top out
func set_stone(x, y byte, apply bool, clear bool) byte {
	return _set_stone(board[:], col_count, row_count, this_stone, this_color, x, y, rotation, apply, clear)
}

func fix_rotation_coordinates(old_rotation, new_rotation byte) {
	var new_x = int(x)
	var new_y = int(y)
	if old_rotation > 0 {
		var offset = stone_rotaion_fixes[this_stone][old_rotation-1]
		new_x = new_x - offset[0]
		new_y = new_y - offset[1]
	}
	if new_rotation > 0 {
		var offset = stone_rotaion_fixes[this_stone][new_rotation-1]
		new_x = new_x + offset[0]
		new_y = new_y + offset[1]
	}
	if new_x < 0 {
		new_x = 0
	} else if new_x >= int(col_count) {
		new_x = int(col_count) - 1
	}
	if new_y < 0 {
		new_y = 0
	} else if new_y >= int(row_count) {
		new_y = int(row_count) - 1
	}
	x = byte(new_x)
	y = byte(new_y)
}

func remove_row(del_row byte) {
	for row := del_row; row > 0; row-- {
		for col := byte(0); col < col_count; col++ {
			var i = row*col_count + col
			board[i] = board[i-col_count]
		}
	}
	for col := byte(0); col < col_count; col++ {
		board[col] = 0
	}
}
func check_for_completed_rows() {
	var removed_rows byte = 0
	for row := byte(0); row < row_count; row++ {
		var completed = true
		for col := byte(0); col < col_count; col++ {
			if board[row*col_count+col] == 0 {
				completed = false
				break
			}
		}
		if completed {
			remove_row(row)
			removed_rows++
		}
	}
	switch removed_rows {
	case 1:
		score += 40 * uint16(level+1)
	case 2:
		score += 100 * uint16(level+1)
	case 3:
		score += 300 * uint16(level+1)
	case 4:
		score += 1200 * uint16(level+1)
	}
}

var tick byte = 0
var last_gamepad uint8 = 0

func _update() {
	var gamepad = *w4.GAMEPAD1
	var changed_gamepad = gamepad & (last_gamepad ^ 0xFF)
	var softdrop bool = false

	if changed_gamepad == 0 {
		remaining_key_pause--
		if remaining_key_pause == 0 {
			changed_gamepad = gamepad
			remaining_key_pause = key_pause
		}
		remaining_softdrop_pause--
		if remaining_softdrop_pause == 0 {
			softdrop = gamepad&w4.BUTTON_DOWN != 0
			remaining_softdrop_pause = softdrop_pause
		}
	} else {
		remaining_key_pause = first_key_pause
		remaining_softdrop_pause = softdrop_pause
	}

	if changed_gamepad&w4.BUTTON_2 != 0 {
		_init()
	}
	if game_over {
		return
	}

	// clear current stone
	set_stone(x, y, true, true)

	if changed_gamepad&w4.BUTTON_LEFT != 0 && x > 0 && set_stone(x-1, y, false, false) == 0 {
		x--
	}
	if changed_gamepad&w4.BUTTON_RIGHT != 0 && x < col_count-1 && set_stone(x+1, y, false, false) == 0 {
		x++
	}
	if changed_gamepad&w4.BUTTON_UP != 0 {
		var old_rotation = rotation
		rotation++
		if rotation >= 4 {
			rotation = 0
		}
		fix_rotation_coordinates(old_rotation, rotation)

		var loop bool = true
		for loop {
			switch set_stone(x, y, false, false) {
			case 0:
				loop = false
			case 1:
				{
					rotation = old_rotation
					loop = false
				}
			case 2:
				x++
			case 3:
				x--
			case 4:
				loop = false
			}
		}
	}
	if softdrop {
		if set_stone(x, y+1, false, false) == 0 {
			y++
			score++
		}
	}
	last_gamepad = gamepad

	tick++
	if tick%speed == 0 {
		tick = 0

		if set_stone(x, y+1, false, false) == 0 {
			y++
			sound_drop()
		} else {
			set_stone(x, y, true, false)
			check_for_completed_rows()
			new_stone(false)
		}
	}
	set_stone(x, y, true, false)
}

func _render() {
	for row := 0; row < int(row_count); row++ {
		for col := 0; col < int(col_count); col++ {
			var i = int(row)*int(col_count) + int(col)
			if board[i] != 0 {
				var border_color = 5 - board[i]
				if border_color == 1 {
					border_color = 2
				}
				*w4.DRAW_COLORS = uint16(board[i] | border_color<<4)
				w4.Rect(x_offset+col*int(scale), y_offset+row*int(scale), scale, scale)
			}
		}
	}
	for row := 0; row < int(preview_row_count); row++ {
		for col := 0; col < int(preview_col_count); col++ {
			var i = int(row)*int(preview_col_count) + int(col)
			if preview_board[i] != 0 {
				var border_color = 5 - preview_board[i]
				if border_color == 1 {
					border_color = 2
				}
				*w4.DRAW_COLORS = uint16(preview_board[i] | border_color<<4)
				w4.Rect(preview_x_offset+col*int(scale), preview_y_offset+row*int(scale), scale, scale)
			}
		}
	}
	*w4.DRAW_COLORS = uint16(0x14)
	w4.Text("Score:", int(col_count*scale)+2, 20)
	var score_text = num2str(int(score))
	w4.Text(score_text, 158-8*len(score_text), 30)
	w4.VLine(int(col_count*scale), 0, 160)
	if game_over {
		*w4.DRAW_COLORS = uint16(0x41)
		w4.Rect(40, 46, 80, 16)
		*w4.DRAW_COLORS = uint16(0x14)
		w4.Text("Game Over", 44, 50)
		return
	}
}

func _init() {
	speed = 30
	score = 0
	game_over = false

	for i := 0; i < row_count*col_count; i++ {
		board[i] = 0
	}
	scale = uint(width)
	if width > height {
		scale = uint(height)
	}
	x_offset = 0 //80 - int(scale*uint(col_count)>>1)
	y_offset = 160 - int(scale*uint(row_count))

	preview_x_offset = 150 - int(scale*uint(preview_col_count))
	preview_y_offset = 80

	new_stone(true)
	set_stone(x, y, true, false)
	initialized = true
}

//go:export update
func update() {
	if !initialized {
		_init()
	}
	_update()
	_render()
}
