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
	"fmt"
	"os"
	"path/filepath"
	"throosea.com/gofar/lib"
	"io/ioutil"
	"strings"
	"encoding/json"
	"time"
)

var resourceList []string
var binprefix string
var processType = "GENERAL"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "you have to specify program name. e.g) gofar mmbatch linux_amd64\n")
		return
	}
	proc := os.Args[1]
	if len(os.Args) >= 3 {
		binprefix = os.Args[2]
	}
	gopath := os.Getenv("GOPATH")

	binpath, e := lib.EnsureBinary(binprefix, proc, gopath)
	if e != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}

	fmt.Printf("binpath : %s\n", binpath)
	resourceList = make([]string, 0)
	foundDir := findPropertyFromSrc(proc + ".", filepath.Join(gopath, "src"))
	if len(resourceList) == 0 {
		fmt.Fprintf(os.Stderr, "not found %s resource\n", proc)
		return
	}

	farDir := filepath.Join(gopath, "far", proc)

	determineProcessType(proc)

	lib.EnsureDirectory(farDir)
	//farName := time.Now().Format("2006-01-02_15-04-05.far")
	//farName = fmt.Sprintf("%s_%s", proc, farName)
	farName := fmt.Sprintf("%s.far", proc)
	farPath := filepath.Join(farDir, farName)

	fmt.Printf("target : %s\n", farPath)

	tmpDir, err := ioutil.TempDir("", "gofar")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}
	defer os.RemoveAll(tmpDir) // clean up

	// binary 복사
	targetBin := filepath.Join(tmpDir, proc)
	err = lib.CopyFile(binpath, targetBin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}
	os.Chmod(targetBin, 0755)

	for _, r := range resourceList {
		target := filepath.Join(tmpDir, filepath.Base(r))
		err = lib.CopyFile(r, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", e.Error())
			return
		}
	}

	// create deployment.json
	m := make(map[string]interface{})
	m["process"] = proc
	m["process_type"] = processType
	build := make(map[string]interface{})
	zoneName, _ := time.Now().Zone()
	build["time"] = time.Now().Format(yyyyMMddHHmmss) + " " + zoneName
	fmt.Printf("srcDir : %s\n", foundDir)
	gitBranch := lib.ReadGitBranch(foundDir)
	if len(gitBranch) > 0 {
		git := make(map[string]string)
		git["branch"] = gitBranch
		gitCommit := lib.ReadGitCommit(foundDir, gitBranch)
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
		fmt.Fprintf(os.Stderr, "%s\n", e.Error())
		return
	}

	lib.Zipit(tmpDir, farPath)
	fmt.Printf("Finished far zip : %s\n", farPath)
}

const (
	yyyyMMddHHmmss = "2006-01-02 15:04:05"
)

var includeSuffixList = [...]string{"properties", "xml", "json", "yaml", "sh"}

func findPropertyFromSrc(proc string, path string) string {
	foundDir := path
	candidate := false
	if strings.HasPrefix(proc, filepath.Base(path)) {
		candidate = true
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

func determineProcessType(proc string) {
	target := proc + ".ui.xml"

	for _, file := range resourceList {
		if filepath.Base(file) == target {
			processType = "USER_INTERACTIVE"
			return
		}
	}
}