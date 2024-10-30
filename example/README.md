## Example

This directory provides a sample project which relies on `gomake` to supply the build system. For the purpose of demonstration, this project has several sub-packages which rely on one another. The build code actually handling all this, is located in `build/`.

To use this example, do the following:

```
$ go build -C ./build -o ../make.exe
$ ./make.exe
```

To invoke individual units, specify their name:

```
$ ./make.exe clean
```
