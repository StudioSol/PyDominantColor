package main

/*
#cgo pkg-config: python-2.7
#define Py_LIMITED_API
#include <Python.h>
int PyArg_ParseTuple_S(PyObject *, char **);
*/
import "C"

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/riff"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"github.com/brunomvsouza/dominantcolor"
)

var dominantColor *dominantcolor.DominantColor

func init() {
	dominantColor = dominantcolor.NewDefault()
}

// FromImage returns the dominant color (HEX format) of a given imageURI
//export FromImage
func FromImage(self, args *C.PyObject) *C.PyObject {
	var cImageURI *C.char
	if C.PyArg_ParseTuple_S(args, &cImageURI) == 0 {
		return C.PyString_FromString(C.CString(""))
	}

	imageURI := C.GoString(cImageURI)

	file, err := os.Open(imageURI)
	defer file.Close()
	if err != nil {
		return C.PyString_FromString(C.CString(""))
	}

	image, _, err := image.Decode(file)
	if err != nil {
		return C.PyString_FromString(C.CString(""))
	}

	dominantColor := dominantColor.FromImage(image)
	hexColor := rgbaToHex(&dominantColor)
	return C.PyString_FromString(C.CString(hexColor))
}

// rgbaToHex returns the correspondent HEX code for a given color.RGBA color
func rgbaToHex(color *color.RGBA) string {
	if color == nil {
		return ""
	}
	hexColor := fmt.Sprintf("%.2X%.2X%.2X", color.R, color.G, color.B)
	if hexColor == "" {
		return ""
	}
	return strings.ToLower(hexColor)
}

func main() {}
