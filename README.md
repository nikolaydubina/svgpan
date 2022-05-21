# svgpan

_Pan and Zoom of SVG in your Go front-end app in browser._

This is port of JS [svgpan](https://github.com/aleofreddi/svgpan). As of 2022-05-21, Google has just took a bunch of JS code and [pasted](https://github.com/google/pprof/blob/master/third_party/svgpan/svgpan.go) it as raw const string. 🤦🏻‍♂️

### Troubleshooting

If you are getting following error

```
$ go get github.com/nikolaydubina/svgpan
package github.com/nikolaydubina/svgpan
    imports syscall/js: build constraints exclude all Go files in /usr/local/go/src/syscall/js
```

Download like

```
$ GOARCH=wasm GOOS=js go get github.com/nikolaydubina/svgpan
```
