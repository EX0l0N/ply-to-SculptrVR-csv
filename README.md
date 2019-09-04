# ply-to-SculptrVR-csv

### ply2csv takes [PLY files](https://en.wikipedia.org/wiki/PLY_(file_format)) and puts out `Data.csv`.

![](https://raw.githubusercontent.com/EX0l0N/ply-to-SculptrVR-csv/master/img/eike.jpg)

## Motivation

[SculptrVR](http://www.sculptrvr.com/) is a nice VR app available for several platforms.  
It enables you to sculpt things out of nothing using additive and substractive operations in a virtual environment.

Unlike the most programs that let you create things in VR it uses voxels (octree based) under the hood and only uses triangles for visualization.

That said it seems like the perfect solution to me to work on the results of **3D scanning**, which create clouds of unconnected colored 3D points.  
Creating meshes out of nothing to use these point clouds can be a pain.

Using SculptrVR we might be able to work on these point clouds and use SculptrVR to export meshed models to other applications.

## Preperations

You are going to need some PLY file. There are some expectations to meet that I assumed to be acceptable during writing.

- binary, little endian encoding
- no normals on vertices (since they are muxed into the vertex stream it breaks my code)
- nothing but coordinates and color to be precise
- your model should be normalized to a [-1, 1] bounding box (for scaling)

All these requirements should be easy to meet, exporting you model as `.ply` with [MeshLab](http://www.meshlab.net/) and disabling everything but vertex color in the export.  
Make sure to check binary export. (No ASCII support)

## How to run

I have prepared binaries for Linux and Windows. You may grab a suitable version from the `bin/` folder of this repo.

- [Windows](https://github.com/EX0l0N/ply-to-SculptrVR-csv/raw/master/bin/windows_amd64/ply2csv.exe)
- [Linux](https://github.com/EX0l0N/ply-to-SculptrVR-csv/raw/master/bin/linux_amd64/ply2csv)

You are expected to run this from any kind `shell`. (sorry no GUI(, yet?))

Since I even use `bash` [(MSYS2)](https://www.msys2.org/) when on windows I have no idea how this would look on the default Windows command line.  
But you should be able to figure it out.

Be warned: **The output is always `Data.csv` in your current working directory**  
This might change if I find the time to work on minor stuff.

### Invocation

```
./ply2csv <scale-factor> [<sphere-size>] <ply-file>
```

|option|meaning |
|:-|:-|
|`scale-factor`:|something that could be parsed as a float|
|`sphere-size`:|**optional** some float|
|`ply-file`:|a path to your `ply` file that works for your os|

In practice this could translate to something like:

```
./ply2csv 512 point_cloud.ply
```

**OR**

```
./ply2csv 200 1.5 point_cloud.ply
```

## What it does

#### If you don't specify an extra `sphere-size`…
`ply2csv` will take your points, scale their coordinates by the scale factor and raster the resulting coordinates to `int`.  
If several points fall together due to the effect of rouding (which is actually a good thing to create more dense data), the colors of all those points will be averaged.

This effect should be used to scale your point cloud to an optimal size, where most voxels are touching each other, but you don't loose to many of them because they got merged.

*Go play with scale!*

#### If you use `sphere-size`…
then `ply2csv` will only scale coordinates, but not raster anything and neither colors nor vertices will be merged. As a result the `Data.csv` will be populated with (scaled) floating point coordinates and every point will be interpreted as sphere later on. The size of that sphere will be the value that you specified, so it is fixed for all points.

To get a good model with mostly solid geometry you need to play with scale and sphere-size. Again optimal values depend on your data.

In both cases the Y & Z axis will be swapped While saving , because ply default is Y-up and Z-depth.

## How to import to SculptrVR

It's a little secret, actually - _and it does only work for the PC version!_

You have to move the `Data.csv` file into a folder named `CSVs` at the top-level of the SculptrVR installation folder.  
_Which is **not** your documents folder_.

Here's a piece of my SteamLibrary to help you figure out where:

```
SteamLibrary
└── steamapps
    ├── common
    │   └── sculptrvr
    │       ├── Engine
    │       ├── SculptrVR
    │       │   ├── Binaries
    │       │   ├── Content
    │       │   ├── CSVs
    │       │   │   └── Data.csv
    │       │   └── Plugins
    │       └── SculptrVR.exe
```

If you created that folder and put `Data.csv` there, you may press `ctrl-shift-L` anytime in SculptrVR to load the data.  
Be sure that **the window has focus** (if you see a steam dialog in front, click into SculptrVRs window).

**In voxel mode you will not see very much, as long as your current layer is not set to block mode.**

The other two display modes try to smooth surfaces, which will totally eat up a single layer of voxels - especially when there are holes in it.

**If you used the sphere mode…**

then you should see spheres popping up, but the whole import might take up a minute until all spheres have been placed.

## What the future might bring

The author of SculptrVR has described several alternative formats for the `Data.csv` file.

There is still a possibility to specify coordinates as floats and put spheres instead of plain voxels. This might be interesting in some cases, so I'm going to implement that, too.  
Maybe this will help with very scattered point clouds.
