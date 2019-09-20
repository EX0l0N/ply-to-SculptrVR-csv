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

**Be warned:** _The complete syntax was just changed to use flags!_

If you have used another version before, you still should read this carefully again.

### Usage

```
$> ply2csv

  -h    this text
  -help
        this text
  -massive
        try to create more dense data by using multiple voxels per point
  -o string
        use this filepath as output (default "Data.csv")
  -scale float
        multiply you models coordinates by this factor (default 100)
  -sphere float
        enable sphere mode and put spheres of this size (default -1)

```

#### This gives you three basic modes of operation which I'll explain here:

- All three modes accept an extra `-o <output-file>` flag.
- The Y & Z axis will always be swapped While saving, because ply default is Y-up and Z-depth.

### _default_ mode

`./ply2csv <your-input.ply>`

`ply2csv` will take your points, scale their coordinates by the default scale factor of `100` and raster the resulting coordinates to `int`.  
If several points fall together due to the effect of rouding (which is actually a good thing to create more dense data), the colors of all those points will be averaged.

This effect should be used to scale your point cloud to an optimal size, where most voxels are touching each other, but you don't loose to many of them because they got merged.

To use a scale factor other than `100`, you need to use the `-scale` flag

`./ply2csv -scale 768 <your-input.ply>`

*Go play with scale!*

### _sphere_ mode

This mode was implemented to satisfy another CSV format requirement one could use to import data into SculptrVR.  
_It's a good idea to think of this mode as being unusable in SculptrVR at the moment._

To use sphere mode you need to use the `-sphere` flag - most likely in some combination with `-scale`, eg.:

`./ply2csv -sphere 1.5 -scale 256 <your-input.ply>`

In this mode `ply2csv` will only scale coordinates, but not raster anything and neither colors nor vertices will be merged. As a result the ouput will be populated with (scaled) floating point coordinates and every point will be interpreted as sphere later on. The size of that sphere will be the value that you specified, so it is fixed for all points.  
Think of cm as base unit for a size estimate.

To get a good model with mostly solid geometry you need to play with scale and sphere-size. Again optimal values depend on your data.  
**Be warned:** This mode will use _a lot_ of RAM during the import.

### _massive_ mode

Massive mode is a slightly modified version of the _default_ mode, which does not only place a single voxel per rastered coordinate, it places *seven* of them in the form of a 3D cross.

![](https://raw.githubusercontent.com/EX0l0N/ply-to-SculptrVR-csv/master/img/cross.jpg)

Use massive mode like the _default_ mode with an extra `-massive` flag.  

`./ply2csv -massive -scale 768 <your-input.ply>`

**The `-massive` and `-sphere` can not be combined** (that would make no sense).

This mode actually implements what sphere mode is meant do at SculptrVRs side, but it does so while exporting data.  
It helps to create _massive_ geometry that could not only be used if the containing SculptrVR layer is set to voxel, it creates geometry that should be two to three voxels thick (depending on your data) while still maintaining color accuracy of one voxel in most cases.

Geometry like this can be smoothed without being destroyed or removed in the process. Also you might export it as a real mesh without that voxel look later on.

![](https://raw.githubusercontent.com/EX0l0N/ply-to-SculptrVR-csv/master/img/smooth.jpg)

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

**If you exported plain single voxels you need to set your current SculptrVR layer to block mode!**  

Otherwise it's very likely that most of you geometry will be invisible.  
The other two display modes try to smooth surfaces, which will totally eat up a single layer of voxels - especially when there are holes in it.

**If you used the sphere mode…**

then you should see spheres popping up, but the whole import might take up a minute until all spheres have been placed.  
Actually I couldn't test this successfully for at least one time, the placing of spheres seems to depend on your viewing scale so scaling in and out messed up my first import.  
The next try I had been waiting patiently until all my RAM got eaten up. I don't have test data with less than 800k points, which seems so be too much for sphere placement.  
The RAM usage is directly connected to the level of detail you configured your sculpting tools to.

Nevertheless I got the format right. So if this is useful to you, have fun!

**Most of the time, when you're tempted to use _sphere_ mode, just use the new _massive_ mode instead.**
