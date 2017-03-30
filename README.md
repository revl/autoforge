----

_This project is under active development, but it's far from being ready.
Please do not attempt to make use of it until this notice disappears._

----

# autoforge

This utility is an attempt to implement a higher level build system on
top of GNU Autotools. Autoforge uses project definition files in YAML
format to generate autoconf scripts and automake source files from
project templates. It also tracks dependencies between those projects
and creates a meta-Makefile to build the projects in the correct order.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the project definition file
is quite simple, which makes starting a new project a breeze.

## Packages
