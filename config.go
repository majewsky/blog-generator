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
	"path/filepath"
	"regexp"
	"strings"
)

//Configuration contains the configuration of the program.
type Configuration struct {
	SourceDir       string
	SourceURL       string
	TargetDir       string
	TargetURL       string
	PageName        string
	PageDescription string
}

//SourcePath returns a path below the Configuration.SourceDir.
func (c Configuration) SourcePath(path string) string {
	return filepath.Join(c.SourceDir, path)
}

//TargetPath returns a path below the Configuration.TargetDir.
func (c Configuration) TargetPath(path string) string {
	return filepath.Join(c.TargetDir, path)
}

//Config contains the configuration of the program.
var Config Configuration

func init() {
	//expect one argument (configuration file)
	if len(os.Args) != 2 {
		failBecause("usage: %s <config-file>\n", filepath.Base(os.Args[0]))
	}

	//read configuration file
	bytes, err := ioutil.ReadFile(os.Args[1])
	FailOnErr(err)

	//process configuration lines
	rx := regexp.MustCompile(`^(\S+)\s*(.+)$`)
	for _, line := range strings.Split(string(bytes), "\n") {
		//ignore empty lines, comments
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		match := rx.FindStringSubmatch(line)
		if match == nil {
			failBecause("invalid directive: %s\n", line)
		}

		switch match[1] {
		case "source-dir":
			Config.SourceDir = match[2]
		case "source-url":
			Config.SourceURL = match[2]
		case "target-dir":
			Config.TargetDir = match[2]
		case "target-url":
			Config.TargetURL = match[2]
		case "page-name":
			Config.PageName = match[2]
		case "page-desc":
			Config.PageDescription = match[2]
		}
	}

	if Config.SourceDir == "" {
		failBecause("missing source-dir")
	}
	if Config.SourceURL == "" {
		failBecause("missing source-url")
	}
	if Config.TargetDir == "" {
		failBecause("missing target-dir")
	}
	if Config.TargetURL == "" {
		failBecause("missing target-url")
	}
	if Config.PageName == "" {
		failBecause("missing page-name")
	}
	if Config.PageDescription == "" {
		failBecause("missing page-desc")
	}
}

func failBecause(msg string, args ...interface{}) {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	os.Stderr.Write([]byte(msg + "\n"))
	os.Exit(1)
}
