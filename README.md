----

_This project is under active development, but it's far from being ready.
Please do not attempt to make use of it until this notice disappears._

----

# autoforge

This utility is an attempt to add a higher level of abstraction to GNU
Autotools as a build system. Autoforge uses a project definition file in
YAML format to generate all autoconf and automake files from a project
template.  It also tracks dependencies between projects and creates a
meta-Makefile to build those projects in the correct order.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the project definition file
is quite simple, which makes starting a new project a breeze.

## Packages
