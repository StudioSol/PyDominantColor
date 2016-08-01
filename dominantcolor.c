#define Py_LIMITED_API
#include <Python.h>

PyObject * FromImage(PyObject *, PyObject *);

// Workaround missing variadic function support
// https://github.com/golang/go/issues/975
int PyArg_ParseTuple_S(PyObject * args, char ** s) {
    return PyArg_ParseTuple(args, "s", s);
}

static PyMethodDef DominantColorMethods[] = {
    {"FromImage", FromImage, METH_VARARGS, "Given an image file URI it returns the dominant color of the image."},
    {NULL, NULL, 0, NULL}
};

void initdominantcolor(void)
{
    Py_InitModule("dominantcolor", DominantColorMethods);
}
