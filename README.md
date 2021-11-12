_Pan and Zoom of SVG in your Go front-end app in browser._

This is port of JS [svgpan](https://github.com/aleofreddi/svgpan]). As of 2021-11-12, Google has just took a bunch of JS code and [pasted](https://github.com/google/pprof/blob/master/third_party/svgpan/svgpan.go) it as raw const string. 🤦🏻‍♂️

### Who is using this?

- https://github.com/nikolaydubina/jsonl-graph

### Troubleshooting

If you are getting following error

```
$ go get github.com/nikolaydubina/go-svgpan
package github.com/nikolaydubina/go-svgpan
    imports syscall/js: build constraints exclude all Go files in /usr/local/go/src/syscall/js
```

Try to download like

```
$ GOARCH=wasm GOOS=js go get github.com/nikolaydubina/go-svgpan
```
