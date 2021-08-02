//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
// @project gofar
// @author 1100282
// @date 2017. 3. 26. PM 5:25
//

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"throosea.com/gofar/lib"
	"time"
)

var (
	resourceList []string
	processType  = "GENERAL"
)

var usage = `usage: %s [option] file os_arc

golang fatima package builder

positional arguments:
  file                  program name
  os_arc                optional. e.g) linux_amd64

optional arguments:
  -b  string
		extra binaries. e.g) extrabin1,extrabin2,extrabin3
`

var cmdFlag CmdFlags
var version = "0.0.1"

type CmdFlags struct {
	ProgramName string
	OsArc       string
	ExtraBin    []string
}

func (c CmdFlags) String() string {
	extra := ""
	for i, v := range c.ExtraBin {
		if i > 0 {
			extra += ","
		}
		extra += v
	}
	return fmt.Sprintf("programName=[%s], osArc=[%s], extra=[%s]", c.ProgramName, c.OsArc, extra)
}

func parseCmdLines() bool {
	flag.Usage = func() {
		fmt.Printf(string(usage), os.Args[0])
	}

	var extraBinList string
	flag.StringVar(&extraBinList, "b", "", "extra binaries. e.g) extrabin1,extrabin2,extrabin3")
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		return false
	}

	cmdFlag.ProgramName = flag.Args()[0]
	if len(flag.Args()) >= 2 {
		cmdFlag.OsArc = flag.Args()[1]
	}

	if len(extraBinList) > 0 {
		cmdFlag.ExtraBin = make([]string, 0)
		for _, v := range strings.Split(extraBinList, ",") {
			cmdFlag.ExtraBin = append(cmdFlag.ExtraBin, strings.TrimSpace(v))
		}
	}

	//fmt.Printf("%s\n", cmdFlag)

	gopath := os.Getenv("GOPATH")

	if len(cmdFlag.ExtraBin) > 0 {
		extraPathList := make([]string, 0)
		for _, v := range cmdFlag.ExtraBin {
			p, e := lib.EnsureBinary(cmdFlag.OsArc, v, gopath)
			if e != nil {
				p, e = lib.EnsureBinaryWithPath(cmdFlag.OsArc, cmdFlag.ProgramName, "", ".")
				if e != nil {
					fmt.Fprintf(os.Stderr, "%s\n", e.Error())
					return false
				}
			}
			extraPathList = append(extraPathList, p)
		}
		cmdFlag.ExtraBin = extraPathList
	}

	return true
}

var outsideGopath = false

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "version" {
			fmt.Printf("gofar version %s\n", version)
			return
		}
	}

	if !parseCmdLines() {
		return
	}

	gopath := os.Getenv("GOPATH")
	workingDir, _ := os.Getwd()

	binpath, e := lib.EnsureBinary(cmdFlag.OsArc, cmdFlag.ProgramName, gopath)
	if e != nil {
		binpath, e = lib.EnsureBinaryWithPath(cmdFlag.OsArc, cmdFlag.ProgramName, "", workingDir)
		if e != nil {
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			os.Exit(-1)
			return
		}
		outsideGopath = true
	}

	fmt.Printf("binpath : %s (%s)\n", binpath, getFileModtime(binpath))
	resourceList = make([]string, 0)
	foundDir := findPropertyFromSrc(cmdFlag.ProgramName, filepath.Join(gopath, "src"))
	if len(resourceList) == 0 {
		fmt.Fprintf(os.Stderr, "not found %s resource. try again from current directory\n", cmdFlag.ProgramName)
		foundDir = findPropertyFromSrc(cmdFlag.ProgramName, workingDir)
		if len(resourceList) == 0 {
			fmt.Fprintf(os.Stderr, "not found %s resource\n", cmdFlag.ProgramName)
			os.Exit(-1)
			return
		}
	}

	farDir := filepath.Join(gopath, "far", cmdFlag.ProgramName)

	fmt.Printf("farDir : %s\n", farDir)

	determineProcessType(cmdFlag.ProgramName)

	fmt.Printf("EnsureDirectory farDir : %s\n", farDir)
	lib.EnsureDirectory(farDir)
	farName := fmt.Sprintf("%s.far", cmdFlag.ProgramName)
	farPath := filepath.Join(farDir, farName)

	fmt.Printf("target : %s\n", farPath)

	tmpDir, err := ioutil.TempDir("", "gofar")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}
	defer os.RemoveAll(tmpDir) // clean up

	// binary 복사
	targetBin := filepath.Join(tmpDir, cmdFlag.ProgramName)
	err = lib.CopyFile(binpath, targetBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}
	os.Chmod(targetBin, 0755)

	for _, v := range cmdFlag.ExtraBin {
		p := filepath.Base(v)
		targetBin := filepath.Join(tmpDir, p)
		err = lib.CopyFile(v, targetBin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			return
		}
		os.Chmod(targetBin, 0755)
		fmt.Printf("extra binpath : %s (%s)\n", v, getFileModtime(v))
	}

	for _, r := range resourceList {
		target := filepath.Join(tmpDir, filepath.Base(r))
		err = lib.CopyFile(r, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			os.Exit(-1)
			return
		}
	}

	// create deployment.json
	m := make(map[string]interface{})
	m["process"] = cmdFlag.ProgramName
	m["process_type"] = processType
	if len(cmdFlag.ExtraBin) > 0 {
		binNameList := make([]string, 0)
		for _, v := range cmdFlag.ExtraBin {
			binNameList = append(binNameList, filepath.Base(v))
		}
		m["extra_bin"] = binNameList
	}
	build := make(map[string]interface{})
	zoneName, _ := time.Now().Zone()
	build["time"] = time.Now().Format(yyyyMMddHHmmss) + " " + zoneName
	// find author
	user, err := lib.ExecuteShell("whoami")
	if err != nil {
		fmt.Fprintf(os.Stderr, "whoami error : %s\n", err.Error())
		user = "unknown"
	}
	build["user"] = strings.TrimSpace(user)
	fmt.Printf("srcDir : %s\n", foundDir)
	gitBranch, gitHaveDir := lib.ReadGitBranch(foundDir, cmdFlag.ProgramName, "")
	if len(gitBranch) > 0 {
		git := make(map[string]string)
		git["branch"] = gitBranch
		gitCommit := lib.ReadGitCommit(gitHaveDir, gitBranch)
		git["commit"] = gitCommit
		fmt.Printf("mark build info : branch=%s, commit=%s\n", gitBranch, gitCommit)
		build["git"] = git
	}
	m["build"] = build

	b, err := json.Marshal(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}
	depfile := filepath.Join(tmpDir, "deployment.json")
	err = ioutil.WriteFile(depfile, b, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(-1)
		return
	}

	lib.Zipit(tmpDir, farPath)
	fmt.Printf("Finished far zip : %s\n", farPath)
}

const (
	yyyyMMddHHmmss = "2006-01-02 15:04:05"
)

var includeSuffixList = [...]string{"properties", "xml", "json", "yaml", "sh", "yml"}

func findPropertyFromSrc(proc string, path string) string {
	foundDir := path
	candidate := false
	//if strings.HasPrefix(proc, filepath.Base(path)) {
	if filepath.Base(path) == proc {
		candidate = true
	} else if filepath.Base(path) == "cmd" {
		if outsideGopath || pathHasProcName(path, proc) {
			candidate = true
		}
	}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail to read dir : %s\n", err.Error())
		return foundDir
	}

	dirList := make([]os.FileInfo, 0)

	for _, file := range files {
		if file.Name()[0] == '.' {
			continue
		}

		if file.IsDir() {
			dirList = append(dirList, file)
		} else if candidate {
			for _, s := range includeSuffixList {
				if strings.HasSuffix(file.Name(), s) {
					resourceList = append(resourceList, filepath.Join(path, file.Name()))
					break
				}
			}
		}
	}

	if len(resourceList) > 0 {
		return foundDir
	}

	for _, dir := range dirList {
		foundDir = findPropertyFromSrc(proc, filepath.Join(path, dir.Name()))
		if len(resourceList) > 0 {
			return foundDir
		}
	}

	return foundDir
}

func pathHasProcName(path, proc string) bool {
	for true {
		dir := filepath.Dir(path)
		if len(dir) == 0 {
			break
		}
		base := filepath.Base(dir)
		if base == "src" {
			return false
		} else if base == proc {
			return true
		}
		path = dir
	}
	return false
}

func determineProcessType(proc string) {
	target := proc + ".ui.xml"

	for _, file := range resourceList {
		if filepath.Base(file) == target {
			processType = "USER_INTERACTIVE"
			return
		}
	}
}

func getFileModtime(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return err.Error()
	}
	return info.ModTime().String()
}
