# resize

The `resize` command resizes Worldographer map files by adding or removing rows and columns.

## Usage

```
resize -input file.wxx -output newfile.wxx [options]
```

## Options

- `-input file` - Load .wxx file (required)
- `-output file` - Create .wxx file (required)  
- `-debug-utf8 file` - Create debug UTF-8 XML file (optional)
- `-top int` - Number of rows to add to top (negative to crop)
- `-bottom int` - Number of rows to add to bottom (negative to crop)
- `-left int` - Number of columns to add to left (negative to crop)
- `-right int` - Number of columns to add to right (negative to crop)
- `-zoom int` - Zoom level in output file (default 1)
- `-debug-sizing` - Show sizing and orientation information
- `-version` - Show version
- `-build-info` - Show version with build info

## Examples

### Expanding a map

Add 2 rows to the top and 4 columns to the left:
```bash
resize -input map.wxx -output bigger.wxx -top 2 -left 4
```

### Cropping a map

Remove 1 row from the top and 2 columns from the right:
```bash
resize -input map.wxx -output smaller.wxx -top -1 -right -2
```

### Mixed operations

Add 3 rows to bottom, crop 1 column from left:
```bash
resize -input map.wxx -output modified.wxx -bottom 3 -left -1
```

## Behavior

- **Expanding**: New tiles are filled with the "Blank" terrain type
- **Cropping**: Tiles outside the crop area are discarded
- **Features and Labels**: Coordinates are automatically translated to maintain correct positioning
- **Features and Labels off-map**: When cropping, features or labels that end up with coordinates ≤ 0 are removed
- **Minimum size**: The resulting map must be at least 2x2 tiles

## Constraints

- The `-left` parameter must be an even number (Worldographer requirement)
- Input and output files cannot be the same
- The input file must contain a "Blank" terrain slot
