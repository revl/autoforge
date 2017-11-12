// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var commonTemplates = map[string]string{"FileHeader": `{{if .header -}}
# {{Comment .header}}
#

{{end}}`,
}
