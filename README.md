## gomake

Gomake is a library and commandline tool used to facilitate building of Go projects and any additional necessary operations.

The reasoning being that Go by itself, while useful, has a limited build system. It does not handle multi-project setups very well. Additionally, it is non-trivial to deal with external resources that need to be deployed in different ways depending on debug- or release modes. Build tags are not sufficient for this.

Gomake uses a mix of Go's Benchmark- and Test-file setup, combined with the way Zig handles building. This means each project creates a `build.go` file in its root. This file defines build rules with all necessary operations. It is itself a regular Go file which is compiled and run by the `gomake` command line tool upon invocation. The build file imports the `gomake/build` package which offers all the necessary tools to facilitate easy build management.

Gomake tries amend Go's own build system where needed and otherwise seeks to stay out of its way.


### Usage

```
$ git clone https://github.com/jimtwn/gomake
$ cd gomake
$ go install
```


### Dependencies

* Go 1.21.5+


### License

Unless otherwise stated, this project and its contents are provided under a 3-Clause BSD license. Refer to the LICENSE file for its contents.
