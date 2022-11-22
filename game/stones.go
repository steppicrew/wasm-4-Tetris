package game

var atoms = [][7]byte{
	{
		0b01111111,
		0b00111110,
		0b11111111,
		0b11111111,
		0b11111111,
		0b11111111,
		0b10000000,
	},
	{
		0b01111111,
		0b00000110,
		0b00001100,
		0b00011000,
		0b00110000,
		0b01111111,
		0b10000000,
	},
	{
		0b11111111,
		0b00000110,
		0b00001100,
		0b00011000,
		0b00110000,
		0b01111111,
		0b10000000,
	},
}

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

// fixes x/y-positions for rotated stones to appear being rotated on the spot
var stone_rotation_fixes = [7][3][2]int8{
	{{1, 1}, {3, 0}, {1, -2}},
	{{0, 2}, {2, 2}, {2, 0}},
	{{1, 1}, {2, 0}, {1, -1}},
	{{1, 0}, {1, -1}, {0, -1}},
	{{1, 1}, {2, -1}, {0, -1}},
	{{0, 1}, {2, 1}, {1, -1}},
	{{1, 1}, {2, -1}, {1, -1}},
}

var stone_styles = [7]byte{
	0b010110,
	0b000110,
	0b000111,
	0b100110,
	0b000110,
	0b000111,
	0b010110,
}

func get_stone_color(style byte) uint16 {
	var color1 = style & 0b0011
	var color2 = (style >> 2) & 0b0011
	return uint16((color1+1)<<4 | (color2 + 1))
}
func get_stone_atom(style byte) *byte {
	var atom_type = style >> 4
	return &atoms[atom_type][0]
}

// create a new stone and reset position/rotation
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
		this_stone_style = stone_styles[this_stone]
	} else {
		this_stone = next_stone
		this_stone_style = next_stone_style
	}

	next_stone = byte(rnd(len(stones)))
	next_stone_style = stone_styles[next_stone]
	for i := 0; i < preview_row_count*preview_col_count; i++ {
		preview_board[i] = 0
	}

	// stones greater than 1 should be placed in the second row
	if this_stone > 1 && this_stone != 5 {
		my_y = 1
	}

	if next_stone > 1 && next_stone != 5 {
		set_preview_stone(1)
	} else {
		set_preview_stone(0)
	}

	// stone 0 (4-stone-bar) should be placed one stone left
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
		if my_y >= 0 && board[i] != 0 {
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
			continue
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
	return _set_stone(board[:], col_count, row_count, this_stone, this_stone_style, x, y, rotation, apply, clear)
}

func set_preview_stone(y byte) byte {
	return _set_stone(preview_board[:], preview_col_count, preview_row_count, next_stone, next_stone_style, 0, y, 0, true, false)
}

// fix rotation to appear rotated on the spot
func fix_rotation_coordinates(old_rotation, new_rotation byte) {
	var new_x = int8(x)
	var new_y = int8(y)
	if old_rotation > 0 {
		var offset = stone_rotation_fixes[this_stone][old_rotation-1]
		new_x = new_x - offset[0]
		new_y = new_y - offset[1]
	}
	if new_rotation > 0 {
		var offset = stone_rotation_fixes[this_stone][new_rotation-1]
		new_x = new_x + offset[0]
		new_y = new_y + offset[1]
	}
	if new_x < 0 {
		new_x = 0
	} else if new_x >= int8(col_count) {
		new_x = int8(col_count) - 1
	}
	if new_y < 0 {
		new_y = 0
	} else if new_y >= int8(row_count) {
		new_y = int8(row_count) - 1
	}
	x = byte(new_x)
	y = byte(new_y)
}
