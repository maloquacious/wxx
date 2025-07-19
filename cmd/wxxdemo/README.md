# wxxdemo
Wxxdemo shows the steps needed to extract, transform, and create a `.wxx` file.

The WXX package provides better tools for this,
but it is helpful to see how the sausage it made.

## Usage

```bash
go run ./cmd/wxxdemo --import testdata/input/blank-2017-1.73-1.0.wxx --output-path testdata/debug/ 

21:26:55 main.go:89:  importFile      == testdata/input/blank-2017-1.73-1.0.wxx
21:26:55 main.go:90:  outputPath      == testdata/debug
21:26:55 main.go:118: demo: completed setup checks      in 7.125µs
21:26:55 main.go:126: demo: read file from disk         in 25.25µs
21:26:55 main.go:134: demo: completed unzip             in 229.834µs
21:26:55 main.go:147: demo: created testdata/debug/input-utf-16.xml
21:26:55 main.go:148: demo: completed input-utf-16.xml  in 170.833µs
21:26:55 main.go:157: demo: completed utf-16 to utf-8   in 148µs
21:26:55 main.go:175: demo: updated utf-8 encoding      in 154.084µs
21:26:55 main.go:183: demo: created testdata/debug/input-utf-8.xml
21:26:55 main.go:184: demo: completed input-utf-8.xml   in 113.833µs
21:26:55 main.go:197: demo: read map from testdata/input/blank-2017-1.73-1.0.wxx 1.73
21:26:55 main.go:198: demo: completed wxml conversion   in 1.190292ms
21:26:55 main.go:208: demo: created testdata/debug/input.json
21:26:55 main.go:209: demo: completed input.json        in 513.333µs
21:26:55 main.go:222: demo: completed output.json       in 176.209µs
21:26:55 main.go:223: demo: created testdata/debug/output.json
21:26:55 main.go:231: demo: completed wmap to tmap      in 51.084µs
21:26:55 main.go:239: demo: completed tmap to xml       in 357.25µs
21:26:55 main.go:246: created testdata/debug/output-utf-8.xml
21:26:55 main.go:247: demo: completed output-utf-8.xml  in 3.143791ms
21:26:55 main.go:256: demo: completed utf-8 to utf-16   in 430.208µs
21:26:55 main.go:264: created testdata/debug/output-utf-16.xml
21:26:55 main.go:265: demo: completed output-utf-16.xml in 110.875µs
21:26:55 main.go:273: demo: completed compress xml      in 663.042µs
21:26:55 main.go:281: created testdata/debug/output.wxx
21:26:55 main.go:282: demo: completed output.wxx        in 4.467458ms
21:26:55 main.go:284: demo: completed                   in 4.473541ms
```