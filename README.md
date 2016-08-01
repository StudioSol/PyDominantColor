PyDominantColor is a C Python extension responsible for returning the dominant color of an image.

It is compatible only with Python 2.7, but it shouldn't be hard to port it for Python 3+.

```bash
# Build the docker image needed to build the extension
docker build -t builderimage .

# Build de dominantcolor.so extension
docker run --rm -v "$PWD":/go/src/github.com/StudioSol/PyDominantColor \
    -w /go/src/github.com/StudioSol/PyDominantColor builderimage \
    go build -v -buildmode=c-shared -o dominantcolor.so
```
