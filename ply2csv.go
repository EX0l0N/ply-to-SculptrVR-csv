package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
	REQ_X = iota
	REQ_Y
	REQ_Z
	REQ_RED
	REQ_GREEN
	REQ_BLUE
	REQ_ALPHA
	REQ_VERTEX
	REQ_FORMAT
	REQ_FIELD_LEN
)

type ply_header struct {
	num_vertices int64
	has_alpha    bool
	field_order  []byte
}

type point [3]byte

type coord_point struct {
	x, y, z float32
	p       point
}

type pointcloud map[float32]map[float32]map[float32]point

func parse_header(in io.Reader) ply_header {
	var checked_fields [REQ_FIELD_LEN]bool
	var header ply_header
	header.field_order = make([]byte, 0, 7)

	r := csv.NewReader(in)
	r.Comma = ' '
	r.FieldsPerRecord = -1

	if magic, err := r.Read(); err != nil {
		panic(err)
	} else if len(magic) != 1 || magic[0] != "ply" {
		fmt.Println("Magic line corrupted:", magic)
	}

	for run := true; run; {
		line, err := r.Read()
		if err != nil {
			panic(err)
		}

		switch line[0] {
		case "comment":
			fmt.Println("Ignoring comment:", line[1:])
		case "format":
			if line[1] != "binary_little_endian" || line[2] != "1.0" {
				fmt.Println(`Format needs to be exactly "binary_little_endian 1.0"`)
				fmt.Printf("Got %q instead.", line[1:])
				panic("Can't parse format")
			}
		case "element":
			if line[1] == "vertex" {
				fmt.Print("Parsing vertex count: ")
				if i, err := strconv.ParseInt(line[2], 10, 64); err != nil {
					panic(err)
				} else {
					header.num_vertices = i
				}
				fmt.Printf("Setting up for %d vertices.\n", header.num_vertices)
			} else {
				fmt.Printf("I hope it's ok to totally ignore %q.\n", line)
			}
		case "property":
			switch line[1] {
			case "float":
				switch line[2] {
				case "x":
					checked_fields[REQ_X] = true
					header.field_order = append(header.field_order, REQ_X)
				case "y":
					checked_fields[REQ_Y] = true
					header.field_order = append(header.field_order, REQ_Y)
				case "z":
					checked_fields[REQ_Z] = true
					header.field_order = append(header.field_order, REQ_Z)
				default:
					fmt.Println(line)
					panic("A parsing error at this stage will most likely be caused by normals. Please don't include normals in your ply.")
				}
			case "uchar":
				switch line[2] {
				case "red":
					checked_fields[REQ_RED] = true
					header.field_order = append(header.field_order, REQ_RED)
				case "green":
					checked_fields[REQ_GREEN] = true
					header.field_order = append(header.field_order, REQ_GREEN)
				case "blue":
					checked_fields[REQ_BLUE] = true
					header.field_order = append(header.field_order, REQ_BLUE)
				case "alpha":
					checked_fields[REQ_ALPHA] = true
					header.field_order = append(header.field_order, REQ_ALPHA)
					header.has_alpha = true
					fmt.Println("Please note that alpha values will be ignored.")
				default:
					fmt.Println(line)
					panic("Can't read that.")
				}
			case "list":
				fmt.Printf("Ignoring unknown porperty list %q.\n", line)
			default:
				fmt.Println(line)
				panic("Can't read that.")
			}
		case "end_header":
			fmt.Println("That's it. From now on I expect binary data.")
			run = false
		default:
			fmt.Println(line)
			panic("Can't read that.")
		}
	}

	return header
}

func read_float32(in *bufio.Reader) float32 {
	var f float32

	if err := binary.Read(in, binary.LittleEndian, &f); err != nil {
		panic(fmt.Sprint("binary.Read failed:", err))
	}

	return f
}

func read_byte(in *bufio.Reader) byte {
	var b byte

	if err := binary.Read(in, binary.LittleEndian, &b); err != nil {
		panic(fmt.Sprint("binary.Read failed:", err))
	}

	return b
}

func read_pointcloud(input *bufio.Reader, header ply_header) pointcloud {
	pc := make(map[float32]map[float32]map[float32]point)

	for c := int64(0); c < header.num_vertices; c++ {
		var cp coord_point

		for i := 0; i < len(header.field_order); i++ {
			switch header.field_order[i] {
			case REQ_X:
				cp.x = read_float32(input)
			case REQ_Y:
				cp.y = read_float32(input)
			case REQ_Z:
				cp.z = read_float32(input)
			case REQ_RED:
				cp.p[0] = read_byte(input)
			case REQ_GREEN:
				cp.p[1] = read_byte(input)
			case REQ_BLUE:
				cp.p[2] = read_byte(input)
			case REQ_ALPHA:
				if d, err := input.Discard(1); err != nil || d != 1 {
					panic("Unable to discard one byte.")
				}
			default:
				panic("Wrong use of field order.")
			}
		}
	}

	return pc
}

func main() {
	var header ply_header
	var input *bufio.Reader

	if infile, err := os.Open(os.Args[1]); err != nil {
		panic(err)
	} else {
		input = bufio.NewReader(infile)
	}
	header = parse_header(input)
	fmt.Println(header)
	cloud := read_pointcloud(input, header)
	fmt.Println(cloud)
}
