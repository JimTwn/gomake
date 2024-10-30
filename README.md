## gomake

**Note**: Highly experimental and flaky. Do not use in production.

Gomake is a library used to facilitate building of complex Go projects and any additional necessary operations.

The reasoning being that Go by itself, while useful, has a limited build system. It does not handle multi-project setups very well. Additionally, it is non-trivial to deal with external resources that need to be deployed in different ways depending on debug- or release modes. Build tags are not sufficient for this. The usual solution is to provide something like a Makefile which in turn invokes Go where needed. This works, but adds a dependency on Make or whatever other thirdparty build system one uses. We opt instead to write our build scripts in Go as a full Go program, which is compiled in the main project directory. From there it is invoked to build the desired project components.

Each project creates a new Go executable package (usually in a `build` subdirectory.) This package is our build program and imports the `gomake` linrary. Refer to the `example` directory for a sample project.


### Usage

```
$ go get https://github.com/jimtwn/gomake
```


### Dependencies

* Go 1.21.5+


### License

Unless otherwise stated, this project and its contents are provided under a 3-Clause BSD license. Refer to the LICENSE file for its contents.
