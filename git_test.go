/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with p work for additional information
 * regarding copyright ownership.  The ASF licenses p file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use p file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 * @project fatima
 * @author DeockJin Chung (jin.freestyle@gmail.com)
 * @date 22. 8. 22. 오후 9:02
 */

package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestReadGitInfo(t *testing.T) {
	path, _ := os.Getwd()
	gitInfo := readGitInfo(path)
	assert.True(t, gitInfo.Valid, "git info is not valid")
	fmt.Printf("BranchName : %s\n", gitInfo.BranchName)
	fmt.Printf("CommitHash : %s\n", gitInfo.CommitHash)
	fmt.Printf("LastCommitMessage : %s\n", gitInfo.LastCommitMessage)
}
