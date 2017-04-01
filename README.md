----

_This project is under active development, but it's far from being ready.
Please do not attempt to make use of it until this notice disappears._

----

# autoforge

This utility is an attempt to implement a higher level build system on
top of GNU Autotools. Autoforge generates autoconf scripts and automake
source files from whole project templates. It also tracks inter-project
dependencies and creates a meta-Makefile to build the projects in the
correct order.

## Packages

An Autoforge package is a collection of C/C++ source files combined with
a project definition file in YAML format. Among other parameters, the
project definition file specifies a template that the package uses,
which, in turn, determines the type of binary that the package produces.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the project definition file
is quite simple, which makes starting a new project a breeze.

## Package templates
