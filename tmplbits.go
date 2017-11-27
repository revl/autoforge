// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var commonDefinitions = map[string]string{
	"FileHeader": `{{if .header}}{{Comment .header}}
{{end}}`,
	"Snippet": `{{if .snippets}}{{if index .snippets .filename}}
{{index .snippets .filename}}{{end}}{{end}}`,
	"Multiline": `{{range .}} \
	{{.}}{{end}}`,
}

var commonTemplateFiles = []embeddedTemplateFile{
	embeddedTemplateFile{"autogen.sh", 0755,
		[]byte(`#!/bin/sh

{{template "FileHeader" . -}}
aclocal &&
	libtoolize --automake --copy && \
	autoheader && \
	automake --foreign --add-missing --copy && \
	autoconf
`)},
}
