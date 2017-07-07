# {{.Name}}: {{.Description}}
#
# {{.Copyright}}
#
# {{.License}}
#
{{if .SRC_MAKEFILE_AM}}
{{.SRC_MAKEFILE_AM}}{{end}}
if DEBUG
bin_PROGRAMS = {{.Name}}d
else
bin_PROGRAMS = {{.Name}}
endif

sources ={{range $v := .SOURCES}} \
	{{$v}}{{end}}

{{VarName .Name}}d_SOURCES = $(sources)
{{VarName .Name}}_SOURCES = $(sources)

{{if .SRC_EXTRA_DIST}}EXTRA_DIST = \
{{.SRC_EXTRA_DIST}}
{{end}}MAINTAINERCLEANFILES = Makefile.in
