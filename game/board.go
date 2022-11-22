package game

const row_count = 20
const col_count = 10

const preview_row_count = 2
const preview_col_count = 4

var board [row_count * col_count]byte
var preview_board [preview_row_count * preview_col_count]byte
var remove_lines [row_count]byte

var width byte = 160 / col_count
var height byte = 160 / row_count
