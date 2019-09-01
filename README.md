# ply-to-SculptrVR-csv

### ply2csv takes [PLY files](https://en.wikipedia.org/wiki/PLY_(file_format)) and puts out `data.csv`.

## Motivation

[SculptrVR](http://www.sculptrvr.com/) is a nice VR app available for several platforms.  
It enables you to sculpt things out of nothing using additive and substractive operations in a virtual environment.

Unlike the most programs that let you create things in VR it seems to use voxels (octree based) under the hood and only uses triangles for visualization.

That said it seems like the perfect solution to me to work on the results of **3D scanning**, which create clouds of unconnected colored 3D points.  
Creating meshes out of nothing to use these point clouds can be a pain.

Using SculptrVR we might be able to work on these point clouds and use SculptrVR to export meshed models to other applications.

## Preperations

You are going to need some PLY file. There are some expectations to meed that I assumed to be acceptable during writing.

- binary, little endian encoding
- no normals on vertices (since they are muxed into the vertex stream it breaks my code)
- nothing but coordinates and color to be precise
- your model should be normalized to a [-1, 1] bounding box

All these requirements should be easy to meet, exporting you model as `.ply` with [MeshLab](http://www.meshlab.net/) and disabling everything but vertex color in the export.  
Make sure to check binary export. (No ASCII support)

## How to run

I have prepared binaries for Linux and Windows. You may grab a suitable version from the `bin/` folder of this repo.

- [Windows](https://github.com/EX0l0N/ply-to-SculptrVR-csv/raw/master/bin/windows_amd64/ply2csv.exe)
- [Linux](https://github.com/EX0l0N/ply-to-SculptrVR-csv/raw/master/bin/linux_amd64/ply2csv)

You are expected to run this from any kind `shell`. (sorry no GUI(, yet?))

Since I even use `bash` [(MSYS2)](https://www.msys2.org/) when on windows I have no idea how this would look on the default Windows command line.  
But you should be able to figure it out.

Be warned: **The output is alway `data.csv` in your current working directory**  
This might change if I find the time to work on minor stuff.

### Invocation

```
./ply2csv <scale-factor> <ply-file>
```

|option|meaning |
|:-|:-|
|`scale-factor`:|something that could be parsed as a float|
|`ply-file`:|a path to your `ply` file that works for your os|

In practice this could translate to something like:

```
./ply2csv 512 point_cloud.ply
```

## What it does

The current version of `ply2csv` will take your points, scale their coordinates by the scale factor and raster the resulting coordinates to `int`.  
If several points fall together due to the effect of rouding (which is actually a good thing to create more dense data), the colors of all those points will be averaged.

## What the future might bring

The author of SculptrVR has described several alternative formats for the `data.csv` file.

They sound interesting because they don't import voxels directly but place sizable spheres in SculptrVRs virtual space.  
It's likely that I add options to use one of these output formats as I expect the current output to be either:

- scattered tiny little voxels
- a clumped up tiny model that you can hardly use