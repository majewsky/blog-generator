/*******************************************************************************
*
* Copyright 2016 Stefan Majewsky <majewsky@gmx.net>
*
* This program is free software: you can redistribute it and/or modify it under the
* terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* This program is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* this program. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	dir, err := os.Open("assets")
	failOnErr(err)
	fis, err := dir.Readdir(-1)
	failOnErr(err)

	str := "package main\n//WARNING: autogenerated, do not edit\n"
	for _, fi := range fis {
		if fi.Name() == "main.go" {
			continue
		}

		data, err := ioutil.ReadFile("assets/" + fi.Name())
		failOnErr(err)

		str += fmt.Sprintf("var %s = %#v\n",
			fileNameToIdentifier(fi.Name()),
			string(data),
		)
	}

	failOnErr(ioutil.WriteFile("assets.go", []byte(str), 0644))
}

func failOnErr(err error) {
	if err != nil {
		os.Stderr.Write([]byte(err.Error() + "\n"))
		os.Exit(1)
	}
}

var rx1 = regexp.MustCompile(`[^A-Za-z]+`)
var rx2 = regexp.MustCompile(`(?:-|^)[A-Za-z]?`)

func fileNameToIdentifier(name string) string {
	normalized := rx1.ReplaceAllString(name, "-")
	return "Asset" + rx2.ReplaceAllStringFunc(normalized, func(s string) string {
		return strings.ToUpper(strings.TrimPrefix(s, "-"))
	})
}
