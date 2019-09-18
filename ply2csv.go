package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"flag"
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

var config struct {
	scale, sphere   float64
	help, massive   bool
	infile, outfile string
}

type ply_header struct {
	num_vertices int64
	has_alpha    bool
	field_order  []byte
}

type point [3]byte

type fcoords [3]float32
type icoords [3]int32

func (fc fcoords) scale_and_raster(f float64) icoords {
	return icoords{
		int32(math.Round(f * float64(fc[0]))),
		int32(math.Round(f * float64(fc[1]))),
		int32(math.Round(f * float64(fc[2]))),
	}
}

type coord_point struct {
	x, y, z float32
	r, g, b byte
}

func (cp coord_point) point() point {
	return point{cp.r, cp.g, cp.b}
}

func (cp *coord_point) set_rgb(p point) {
	cp.r = p[0]
	cp.g = p[1]
	cp.b = p[2]
}

func (cp *coord_point) set_xyz(fc fcoords, f float32) {
	cp.x = f * fc[0]
	cp.y = f * fc[1]
	cp.z = f * fc[2]
}

func (cp coord_point) fcoords() fcoords {
	return fcoords{cp.x, cp.y, cp.z}
}

type floatpointcloud map[fcoords][]point
type intpointcloud map[icoords]point

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
	pc := make(map[fcoords][]point)

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
				cp.r = read_byte(input)
			case REQ_GREEN:
				cp.g = read_byte(input)
			case REQ_BLUE:
				cp.b = read_byte(input)
			case REQ_ALPHA:
				if d, err := input.Discard(1); err != nil || d != 1 {
					panic("Unable to discard one byte.")
				}
			default:
				panic("Wrong use of field order.")
			}
		}

		fc := cp.fcoords()

		if _, ok := pc[fc]; !ok {
			pc[fc] = make([]point, 1, 3)
			pc[fc][0] = cp.point()
		} else {
			fmt.Printf("Double vertex encountered at %q, %q, %q.\n", cp.x, cp.y, cp.z)
			pc[fc] = append(pc[fc], cp.point())
		}
	}

	return pc
}

func raster_and_merge_pointcloud(fsc float64, fpc floatpointcloud) intpointcloud {
	rasted := make(map[icoords][]point)
	ipc := make(map[icoords]point)

	fmt.Println("Scaling vertex indices to int32…")

	for fco, _ := range fpc {
		scaled := fco.scale_and_raster(fsc)
		if _, ok := rasted[scaled]; !ok {
			rasted[scaled] = make([]point, 0, 5)
		}
		rasted[scaled] = append(rasted[scaled], fpc[fco]...)
	}

	fmt.Println("Merging colors of double vertices…")

	for ico, _ := range rasted {
		var reds int
		var greens int
		var blues int

		var point_slice_length = len(rasted[ico])

		for _, point := range rasted[ico] {
			reds += int(point[0])
			greens += int(point[1])
			blues += int(point[2])
		}

		ipc[ico] = point{byte(reds / point_slice_length), byte(greens / point_slice_length), byte(blues / point_slice_length)}
	}

	return ipc
}

func open_data_csv() (*os.File, *bufio.Writer) {
	f, err := os.OpenFile(config.outfile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
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

func auto_level(s int) (int32, int) {
	var max_dim int = 1

	for level := 20; level >= 0; level-- {
		if s <= max_dim {
			return int32(max_dim), level
		}

		max_dim <<= 1
	}

	panic("Could not find level for this scale.")
}

func check_dimensions(md int32, ico icoords) bool {
	if ico[0] > md || ico[0] < -md {
		fmt.Println("X value to big for scale.")
		return false
	}
	if ico[1] > md || ico[1] < -md {
		fmt.Println("Y value to big for scale.")
		return false
	}
	if ico[2] > md || ico[2] < -md {
		fmt.Println("Z value to big for scale.")
		return false
	}
	return true
}

func write_data_csv_from_raster(scl float64, r intpointcloud) {
	maxd, level := auto_level(int(math.Ceil(math.Abs(scl))))
	f, bw := open_data_csv()

	bw.WriteString("X, Y, Z, level, R, G, B, mat\r\n")

	for ico, p := range r {
		if !check_dimensions(maxd, ico) {
			continue
		}
		bw.WriteString(fmt.Sprintf("%d, %d, %d, %d, %d, %d, %d, 255\r\n", ico[0], ico[2], ico[1], level, p[0], p[1], p[2]))
	}

	cleanup(f, bw)
}

func dump_data_csv_with_scaled_sphere_positions(scl, spz float32, fpc floatpointcloud) {
	var fp coord_point
	f, bw := open_data_csv()

	bw.WriteString("X, Y, Z, Radius, R, G, B\r\n")

	for fco, _ := range fpc {
		fp.set_xyz(fco, scl)
		for _, p := range fpc[fco] {
			fp.set_rgb(p)

			bw.WriteString(fmt.Sprintf("%.6f, %.6f, %.6f, %.3f, %d, %d, %d\r\n", fp.x, fp.z, fp.y, spz, fp.r, fp.g, fp.b))
		}
	}

	cleanup(f, bw)
}

func check_flags() bool {
	flag.Parse()

	config.infile = flag.Arg(0)

	if config.help {
		flag.PrintDefaults()
		return false
	}

	if config.sphere != -1 && config.massive {
		fmt.Println("Sphere mode and massive mode can't be enabled at the same time.")
		return false
	}

	if config.infile == "" {
		fmt.Println("No input ply filename given.")
		return false
	}

	return true
}

func init() {
	const help = "this text"

	flag.BoolVar(&config.help, "h", false, help)
	flag.BoolVar(&config.help, "help", false, help)
	flag.Float64Var(&config.scale, "scale", 100, "multiply you models coordinates by this factor")
	flag.Float64Var(&config.sphere, "sphere", -1, "enable sphere mode and put spheres of this size")
	flag.BoolVar(&config.massive, "massive", false, "try to create more dense data by using multiple voxels per point")
	flag.StringVar(&config.outfile, "o", "Data.csv", "use this filepath as output")
}

func main() {
	var (
		header ply_header
		input  *bufio.Reader
	)

	if !check_flags() {
		return
	}

	if infile, err := os.Open(config.infile); err != nil {
		panic(err)
	} else {
		defer infile.Close()
		input = bufio.NewReader(infile)
	}

	fmt.Println("Reading ply header…")
	header = parse_header(input)
	fmt.Println("Reading ply vertex data…")
	cloud := read_pointcloud(input, header)

	if config.sphere < 0 {
		raster := raster_and_merge_pointcloud(config.scale, cloud)
		fmt.Printf("Writing %q.\n", config.outfile)
		write_data_csv_from_raster(config.scale, raster)
	} else {
		fmt.Printf("Wrting %q for sphere import.\n", config.outfile)
		dump_data_csv_with_scaled_sphere_positions(float32(config.scale), float32(config.sphere), cloud)
	}
}
