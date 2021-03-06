----

_This project is experimental, unfinished, and all but abandoned._

----

# autoforge

This utility is an attempt to implement a higher level build system on
top of GNU Autotools. Autoforge generates autoconf scripts and automake
source files from whole project templates. It also tracks inter-project
dependencies and creates a meta-Makefile to build the projects in the
correct order.

## Packages

An Autoforge package is a directory containing C/C++ source files and a
package definition file, whose name must consist of the package
directory name and the extension `.yaml`.

Among other parameters, the package definition file specifies a template
that the package uses, which, in turn, determines the type of binary
that the package produces.

Because project templates encapsulate a great deal of complexity that
comes with using Autotools, the structure of the package definition file
is quite simple, which makes starting a new project a breeze.

## Project templates

Project templates contain autoconf and automake source files required
for building the project. Autoforge provides several generic templates.
Additional templates can be created ad hoc.

## Project definition files

By imposing certain restrictions on the project structure, Autoforge
keeps the differences between the projects generated from the same
template to a minimum. For a new project, its package definition file
must specify just a few crucial parameters, such as the name of the
project, the type of the license it uses, etc. A separate section
below describes the full list of parameters that can appear in a
package definition file.

## Package search path

The `AUTOFORGE_PKG_PATH` environment variable defines a colon-separated
list of directories that contain packages.  Autoforge searches for
package definition files in subdirectories of the `AUTOFORGE_PKG_PATH`
directories. Subdirectories without such files are ignored.

## Build directories

Autoforge requires that the packages are built in a dedicated directory
separate from the source tree. A new build directory must be initialized
first by running `autoforge -init`. Certain build configuration
parameters can be specified only during the initialization.

The purpose of each build directory is defined by the combination of
libraries and applications being built.

## Autoforge commands

### Get information on the available packages

The `-query` switch shows the list of all packages found on the search
paths along with package descriptions and other essential information.

To suppress this detailed output and limit the list to just package
names, use the `-brief` option.

### Initialize the build directory

To initialize the build directory, use the `-init` switch. The following
options define the initialization parameters:

- `-installdir`

  Set the target directory for `make install`.

- `-docdir`

  Set the installation directory for documentation.

- `-pkgpath`

  The list of directories where to search for packages. This parameter
  overrides the value of the `$AUTOFORGE_PKG_PATH` environment variable.

- `-workspacedir`

  Set the build directory, which is the current working directory
  by default.

### Prepare packages for building and generate the meta-Makefile

Using Autoforge is an iterative process. Aside from a very limited
number of configuration parameters specified during the initialization,
the build directory can be repurposed at any time by choosing a
different range of packages to build.

After the build directory has been initialized, Autotools source files
must be generated for one or more packages, which must be specified on
the command line as a list of individual packages or a range of packages,
see below. This is the default mode of operation for Autoforge; it is
activated when no other mode is triggered by a command line switch (e.g.
`-init`).

The package range is a selection of packages in the following format:
`[base_pkg]:[dep_pkg]`, where `base_pkg` is a base package and `dep_pkg`
is a package that requires it. When specified like that, the selection
includes the dependency chain of packages from `base_pkg` to `dep_pkg`.
Both base and dependent packages can be omitted, in which case all base
packages or all dependent packages, respectively, will be included in
the selection.

## Appendix. The list of package definition file parameters

Here is the full list of variables that can appear in a package
definition file:

- `name`

  The name of the package. This name does not have to match the name
  of the directory that contains the package.

- `template`

  The name of the package template. At the moment, either `library` or
  `application`.

- `version`

  Package version for use by Automake.

- `license`

  Either a short name of the license ("MIT", "LGPL", "GPL", "Apache",
  "Apache v2.0", etc.) or the full text of the license.

- `version_info`

  API/ABI revision for use by Libtool.

- `requires`

  The list of libraries that the package requires.

- `headers`

  For a library, the list of C/C++ headers exported by the library.

- `sources`

  The list of C/C++ sources containing the implementation.

- `configure`

  A snippet to be embedded in the `configure.in` file. Can be a mix of
  Bourne shell code and Autoconf macros.
