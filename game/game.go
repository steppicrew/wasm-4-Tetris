package game

import (
	"cart/w4"
	"math/rand"
)

var rnd func(int) int

var game_over bool = false

const initial_das = 16
const das = 6

const softdrop_pause = 2

var initialized bool = false

// Style:
//
//	Bit 0-1 border color
//	Bit 2-4 inner color (<<2)
//	Bit 5-6 stone atom (<<4)
var this_stone byte
var this_stone_style byte

var next_stone byte
var next_stone_style byte

var x byte = 4
var y byte = 4
var rotation byte

var are byte = 0
var lines_removed byte = 0
var score uint16 = 0

var remove_lines_frame byte = 0

const remove_lines_frames_count = 5

var scale uint
var x_offset int
var y_offset int
var preview_x_offset int
var preview_y_offset int
var remaining_das byte = 0
var remaining_softdrop_pause byte = 0

var framecount int16 = 0

func _num2str(x uint) string {
	var digit = x % 10
	x = x / 10
	if x > 0 {
		return _num2str(x) + string(rune('0'+digit))
	}
	return string(rune('0' + digit))
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
	_sound(670, 0, 8, 0, 0, 0, 1, 3, 0)
}

func get_level() byte {
	return lines_removed / 10
}

func get_speed() byte {
	var level = get_level()
	switch {
	case level <= 8:
		return 48 - level*5
	case level == 9:
		return 6
	case level >= 10 && level <= 12:
		return 5
	case level >= 13 && level <= 15:
		return 4
	case level >= 16 && level <= 18:
		return 3
	case level >= 19 && level <= 28:
		return 2
	default:
		return 1
	}
}

func get_are(drop_row byte) byte {
	return 10 + ((drop_row+2)/4)*2
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
	var last_removed_row byte = 0
	for row := byte(0); row < row_count; row++ {
		var completed = true
		for col := byte(0); col < col_count; col++ {
			if board[row*col_count+col] == 0 {
				completed = false
				break
			}
		}
		if completed {
			remove_lines[row] = 1
			removed_rows++
			last_removed_row = row
		}
	}
	if removed_rows > 0 {
		var level = get_level()
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
		lines_removed += removed_rows
		remove_lines_frame = remove_lines_frames_count
		are = get_are(row_count - last_removed_row - 1)
	}
}

func center_border_text(text string, y int) {
	var x int = (80 - len(text)<<2)
	*w4.DRAW_COLORS = uint16(0x41)
	w4.Rect(x-4, y-8, uint(len(text)+1)<<3, 16)
	*w4.DRAW_COLORS = uint16(0x14)
	w4.Text(text, x, y-4)

}

var tick byte = 0
var last_gamepad uint8 = 0

func Update() {
	if title_active {
		return
	}
	if remove_lines_frame > 0 {
		return
	}

	if are > 0 {
		are--
		return
	}

	var gamepad = *w4.GAMEPAD1
	var changed_gamepad = gamepad & (last_gamepad ^ 0xFF)
	var softdrop bool = false

	if changed_gamepad&w4.BUTTON_1 != 0 {
		initialize()
	}
	if game_over {
		return
	}

	if changed_gamepad == 0 {
		remaining_das--
		if remaining_das == 0 {
			changed_gamepad = gamepad
			remaining_das = das
		}
		remaining_softdrop_pause--
		if remaining_softdrop_pause == 0 {
			softdrop = gamepad&w4.BUTTON_DOWN != 0
			remaining_softdrop_pause = softdrop_pause
		}
	} else {
		remaining_das = initial_das
		remaining_softdrop_pause = softdrop_pause
	}
	last_gamepad = gamepad

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
			sound_drop()
			score++
		}
	}

	tick++
	if tick%get_speed() == 0 {
		tick = 0

		if set_stone(x, y+1, false, false) == 0 {
			y++
			sound_drop()
		} else {
			set_stone(x, y, true, false)
			check_for_completed_rows()
			if remove_lines_frame == 0 {
				new_stone(false)
			}
		}
	}
	set_stone(x, y, true, false)
}

func Render() {
	framecount++
	if title_active {
		render_title()
		return
	}
	for row := 0; row < int(row_count); row++ {
		for col := 0; col < int(col_count); col++ {
			var i = int(row)*int(col_count) + int(col)
			if board[i] != 0 {
				*w4.DRAW_COLORS = get_stone_color(board[i])
				w4.Blit(get_stone_atom(board[i]), x_offset+col*int(scale), y_offset+row*int(scale)+1, 7, 7, w4.BLIT_1BPP)
				//w4.Rect(x_offset+col*int(scale), y_offset+row*int(scale), scale, scale)
			}
		}
	}
	for row := 0; row < int(preview_row_count); row++ {
		for col := 0; col < int(preview_col_count); col++ {
			var i = int(row)*int(preview_col_count) + int(col)
			if preview_board[i] != 0 {
				*w4.DRAW_COLORS = get_stone_color(preview_board[i])
				w4.Blit(get_stone_atom(preview_board[i]), preview_x_offset+col*int(scale), preview_y_offset+row*int(scale)+1, 7, 7, w4.BLIT_1BPP)
				//w4.Rect(preview_x_offset+col*int(scale), preview_y_offset+row*int(scale), scale, scale)
			}
		}
	}

	if remove_lines_frame > 0 {
		if framecount%4 == 0 {
			remove_lines_frame--
		}
		var board_width = byte(col_count * scale)
		var remove_width = byte(board_width / (remove_lines_frame + 1))
		*w4.DRAW_COLORS = uint16(0x11)
		for row := 0; row < row_count; row++ {
			if remove_lines[row] > 0 {
				w4.Rect((x_offset + int(board_width-remove_width)/2), y_offset+row*int(scale), uint(remove_width), scale)
				if remove_lines_frame == 0 {
					remove_row(byte(row))
					remove_lines[row] = 0
				}
			}
		}
		if remove_lines_frame == 0 {
			new_stone(false)
		}
	}

	*w4.DRAW_COLORS = uint16(0x14)

	w4.Text("Score:", int(col_count*scale)+2, 20)
	var text = num2str(int(score))
	w4.Text(text, 158-8*len(text), 30)

	w4.Text("Level:", int(col_count*scale)+2, 40)
	text = num2str(int(get_level()))
	w4.Text(text, 158-8*len(text), 50)

	w4.Text("Lines:", int(col_count*scale)+2, 60)
	text = num2str(int(lines_removed))
	w4.Text(text, 158-8*len(text), 70)

	w4.VLine(int(col_count*scale), 0, 160)
	if game_over {
		center_border_text("Game Over!", 80)
	}
}

func initialize() {
	rnd = rand.New(rand.NewSource(int64(framecount))).Intn
	w4.Trace("init: " + num2str(int(framecount)))

	are = get_are(0)
	score = 0
	lines_removed = 0
	game_over = false

	for i := 0; i < row_count*col_count; i++ {
		board[i] = 0
	}
	for i := 0; i < row_count; i++ {
		remove_lines[i] = 0
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
