package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"io"
	"math"
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

type floatpointcloud map[float32]map[float32]map[float32][]point
type intpointcloud map[int32]map[int32]map[int32]point

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
			checked_fields[REQ_FORMAT] = true
			if line[1] != "binary_little_endian" || line[2] != "1.0" {
				fmt.Println(`Format needs to be exactly "binary_little_endian 1.0"`)
				fmt.Printf("Got %q instead.", line[1:])
				panic("Can't parse format")
			}
		case "element":
			if line[1] == "vertex" {
				checked_fields[REQ_VERTEX] = true
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

	if !(checked_fields[REQ_FORMAT] &&
		checked_fields[REQ_VERTEX] &&
		checked_fields[REQ_X] &&
		checked_fields[REQ_Y] &&
		checked_fields[REQ_Z] &&
		checked_fields[REQ_RED] &&
		checked_fields[REQ_GREEN] &&
		checked_fields[REQ_BLUE]) {

		fmt.Println(checked_fields)
		panic("Did not see all the required fields in header. Giving up.")
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

func read_pointcloud(input *bufio.Reader, header ply_header) floatpointcloud {
	pc := make(map[float32]map[float32]map[float32][]point)

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

		if _, ok := pc[cp.x]; !ok {
			pc[cp.x] = make(map[float32]map[float32][]point)
		}
		if _, ok := pc[cp.x][cp.y]; !ok {
			pc[cp.x][cp.y] = make(map[float32][]point)
		}
		if _, ok := pc[cp.x][cp.y][cp.z]; !ok {
			pc[cp.x][cp.y][cp.z] = make([]point, 1, 3)
			pc[cp.x][cp.y][cp.z][0] = cp.p
		} else {
			fmt.Printf("Double vertex encountered at %q, %q, %q.\n", cp.x, cp.y, cp.z)
			pc[cp.x][cp.y][cp.z] = append(pc[cp.x][cp.y][cp.z], cp.p)
		}
	}

	return pc
}

func raster_and_merge_pointcloud(fsc float64, fpc floatpointcloud) intpointcloud {
	rasted := make(map[int32]map[int32]map[int32][]point)
	ipc := make(map[int32]map[int32]map[int32]point)

	fmt.Println("Scaling vertex indices to int32…")

	for x, _ := range fpc {
		var scaledx = int32(math.Round(fsc * float64(x)))

		if _, ok := rasted[scaledx]; !ok {
			rasted[scaledx] = make(map[int32]map[int32][]point)
		}

		for y, _ := range fpc[x] {
			var scaledy = int32(math.Round(fsc * float64(y)))

			if _, ok := rasted[scaledx][scaledy]; !ok {
				rasted[scaledx][scaledy] = make(map[int32][]point)
			}

			for z, _ := range fpc[x][y] {
				var scaledz = int32(math.Round(fsc * float64(z)))

				if _, ok := rasted[scaledx][scaledy][scaledz]; !ok {
					rasted[scaledx][scaledy][scaledz] = make([]point, 0, 5)
				}
				rasted[scaledx][scaledy][scaledz] = append(rasted[scaledx][scaledy][scaledz], fpc[x][y][z]...)
			}
		}
	}

	fmt.Println("Merging colors of double vertices…")

	for x, _ := range rasted {

		if _, ok := ipc[x]; !ok {
			ipc[x] = make(map[int32]map[int32]point)
		}

		for y, _ := range rasted[x] {

			if _, ok := ipc[x][y]; !ok {
				ipc[x][y] = make(map[int32]point)
			}

			for z, _ := range rasted[x][y] {
				var reds int
				var greens int
				var blues int

				var point_slice_length = len(rasted[x][y][z])

				for _, point := range rasted[x][y][z] {
					reds += int(point[0])
					greens += int(point[1])
					blues += int(point[2])
				}

				ipc[x][y][z] = point{byte(reds / point_slice_length), byte(greens / point_slice_length), byte(blues / point_slice_length)}
			}
		}
	}

	return ipc
}

func open_data_csv() (*os.File, *bufio.Writer) {
	f, err := os.OpenFile("Data.csv", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(f)

	return f, w
}

func cleanup(f *os.File, bw *bufio.Writer) {
	if err := bw.Flush(); err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}

func write_data_csv_from_raster(r intpointcloud) {
	f, bw := open_data_csv()

	bw.WriteString("X, Y, Z, level, R, G, B, mat\r\n")

	for x, _ := range r {
		for y, _ := range r[x] {
			for z, _ := range r[x][y] {
				p := r[x][y][z]
				bw.WriteString(fmt.Sprintf("%d, %d, %d, 10, %d, %d, %d, 255\r\n", x, z, y, p[0], p[1], p[2]))
			}
		}
	}

	cleanup(f, bw)
}

func dump_data_csv_with_sacled_sphere_positions(scl, spz float32, fpc floatpointcloud) {
	var fp coord_point
	f, bw := open_data_csv()

	bw.WriteString("X, Y, Z, Radius, R, G, B\r\n")

	for x, _ := range fpc {
		for y, _ := range fpc[x] {
			for z, _ := range fpc[x][y] {
				for p, _ := range fpc[x][y][z] {
					fp.p = fpc[x][y][z][p]
					fp.x = scl * x
					fp.y = scl * y
					fp.z = scl * z

					bw.WriteString(fmt.Sprintf("%.6f, %.6f, %.6f, %.3f, %d, %d, %d\r\n", fp.x, fp.z, fp.y, spz, fp.p[0], fp.p[1], fp.p[2]))
				}
			}
		}
	}

	cleanup(f, bw)
}

func main() {
	var (
		header       ply_header
		file_arg_pos = 2
		input        *bufio.Reader
		scale        float64
		spheresize   float32 = -1
	)

	if len(os.Args) < 3 {
		fmt.Println("Usage: ./ply2csv <scale-factor> [<sphere-size>] <ply-file>")
		return
	}

	if s, err := strconv.ParseFloat(os.Args[1], 64); err != nil {
		panic(fmt.Sprintf("Can not parse %q as scale.\n", os.Args[1]))
	} else {
		scale = s
	}

	if s, err := strconv.ParseFloat(os.Args[2], 64); err == nil {
		spheresize = float32(s)
		if spheresize < 0 {
			panic("Oh no. Why did you specify a negative sphere size?")
		}
		file_arg_pos++
	}

	if infile, err := os.Open(os.Args[file_arg_pos]); err != nil {
		panic(err)
	} else {
		defer infile.Close()
		input = bufio.NewReader(infile)
	}

	fmt.Println("Reading ply header…")
	header = parse_header(input)
	fmt.Println("Reading ply vertex data…")
	cloud := read_pointcloud(input, header)

	if spheresize < 0 {
		raster := raster_and_merge_pointcloud(scale, cloud)
		fmt.Println("Wrting Data.csv.")
		write_data_csv_from_raster(raster)
	} else {
		fmt.Println("Wrting Data.csv for sphere import.")
		dump_data_csv_with_sacled_sphere_positions(float32(scale), spheresize, cloud)
	}
}
