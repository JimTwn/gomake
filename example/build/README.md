This directory contains our actual build 'script'. It is a full Go program that uses the `gomake` library to build the other components in the example project.

Since this is a normal Go program, you can do anything you want in here. That includes not using the `gomake` library. It is just a set of helpers that make certain operations simpler. Nothing more.