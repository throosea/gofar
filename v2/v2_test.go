/*
 * //
 * // Licensed to the Apache Software Foundation (ASF) under one
 * // or more contributor license agreements.  See the NOTICE file
 * // distributed with p work for additional information
 * // regarding copyright ownership.  The ASF licenses p file
 * // to you under the Apache License, Version 2.0 (the
 * // "License"); you may not use p file except in compliance
 * // with the License.  You may obtain a copy of the License at
 * //
 * //   http://www.apache.org/licenses/LICENSE-2.0
 * //
 * // Unless required by applicable law or agreed to in writing,
 * // software distributed under the License is distributed on an
 * // "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * // KIND, either express or implied.  See the License for the
 * // specific language governing permissions and limitations
 * // under the License.
 * //
 * // @project fatima
 * // @author DeockJin Chung (jin.freestyle@gmail.com)
 * // @date 21. 8. 8. 오후 3:35
 * //
 */

package v2

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestNewBuildContext(t *testing.T) {
	dir, err := determineProjectBaseDir("/Users/jin/IronForge/gowork/src/hellofatima/controller", "hellofatima")
	if err != nil {
		t.Fatalf("determineProjectBaseDir err : %s", err.Error())
		return
	}

	log.Printf("dir : %s", dir)

	foundDir, err := FindDirectory("/Users/jin/IronForge/gowork/src/mmate/batbadge", "cmd")
	if err != nil {
		t.Fatalf("finddir error : %s", err.Error())
		return
	}
	log.Printf("foundDir : %s", foundDir)

	ctx := &BuildContext{}
	ctx.ProjectBaseDir = "/Users/jin/IronForge/gowork/src/mmate/batbadge"
	determineCmdList(ctx)
	for _, v := range ctx.ProcessList {
		log.Printf("cmd : %s", v.Path)
	}

	dir, err = ioutil.TempDir("/tmp", "hello")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("tmpdir : %s", dir)
	//defer os.RemoveAll(dir)
}
