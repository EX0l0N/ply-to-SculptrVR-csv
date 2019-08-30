package main

import (
	"bufio"
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

func main() {
	var header ply_header
	var input io.Reader

	if infile, err := os.Open(os.Args[1]); err != nil {
		panic(err)
	} else {
		input = bufio.NewReader(infile)
	}
	header = parse_header(input)
	fmt.Println(header)
}
