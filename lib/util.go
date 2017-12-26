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
// @date 2017. 3. 26. PM 5:45
//

package lib

import (
	"path/filepath"
	"fmt"
	"os"
	"io"
	"strings"
	"archive/zip"
	"io/ioutil"
)

func EnsureBinary(binprefix string, proc string, gopathDir string) (string, error) {
	var binpath string
	if len(binprefix) == 0 {
		binpath = filepath.Join(gopathDir, "bin", proc)
	} else {
		binpath = filepath.Join(gopathDir, "bin", binprefix, proc)
	}

	return binpath, CheckFileExist(binpath)
}

func EnsureDirectory(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dir, 0755)
		}
	}

	return nil
}

func CheckFileExist(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("not exist file : %s", path)
		}
	}

	return nil
}

func CopyFile(src string, dst string) error {
	fmt.Printf("src : %s\n", src)
	sFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sFile.Close()

	eFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer eFile.Close()

	_, err = io.Copy(eFile, sFile) // first var shows number of bytes
	if err != nil {
		return err
	}

	err = eFile.Sync()
	if err != nil {
		return err
	}

	return nil
}


func Zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	_, err = os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	//if info.IsDir() {
	//	baseDir = filepath.Base(source)
	//}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		//fmt.Printf("zip path : %s\n", path)
		if source == path {
			return nil
		}

		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func ReadGitBranch(srcDir string) string	{
	headFile := filepath.Join(srcDir, ".git", "HEAD")

	dat, err := ioutil.ReadFile(headFile)
	if err != nil {
		return ""
	}

	head := strings.Trim(string(dat), " \r\n\t")
	idx := strings.LastIndex(head, "/")
	return head[idx+1:]
}

func ReadGitCommit(srcDir string, branch string) string {
	commitFile := filepath.Join(srcDir, ".git", "refs", "heads", branch)
	dat, err := ioutil.ReadFile(commitFile)
	if err != nil {
		return ""
	}

	return string(dat)[:12]
}