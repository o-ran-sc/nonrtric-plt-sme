// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2022: Nordix Foundation
//   %%
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//   ========================LICENSE_END===================================
//

package helmmanagement

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"helm.sh/helm/v3/pkg/time"
	"oransc.org/nonrtric/capifcore/internal/helmmanagement/mocks"
)

func TestNoChartURL_repoNotSetUp(t *testing.T) {
	managerUnderTest := NewHelmManager(nil)

	res := managerUnderTest.SetUpRepo("repoName", "")

	assert.Nil(t, res)
	assert.False(t, managerUnderTest.setUp)
}

func TestSetUpRepoExistingRepoFile_repoShouldBeAddedToReposFile(t *testing.T) {
	settings := createReposFile(t)

	managerUnderTest := NewHelmManager(settings)

	repoName := filepath.Dir(settings.RepositoryConfig)
	repoURL := "http://url"
	managerUnderTest.repo = getChartRepo(settings)

	res := managerUnderTest.SetUpRepo(repoName, repoURL)

	assert.Nil(t, res)
	assert.True(t, containsRepo(settings.RepositoryConfig, repoName))
	assert.True(t, managerUnderTest.setUp)
}

func TestSetUpRepoFail_shouldNotBeSetUp(t *testing.T) {
	settings := createReposFile(t)

	managerUnderTest := NewHelmManager(settings)

	res := managerUnderTest.SetUpRepo("repoName", "repoURL")

	assert.NotNil(t, res)
	assert.False(t, managerUnderTest.setUp)
}

func createReposFile(t *testing.T) *cli.EnvSettings {
	reposDir, err := os.MkdirTemp("", "helm")
	if err != nil {
		t.Errorf("Unable to create temporary directory for repos due to: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(reposDir)
	})

	reposFile := reposDir + "/index.yaml"
	settings := &cli.EnvSettings{
		RepositoryConfig: reposFile,
	}

	repoData := repo.File{
		Generated:    time.Now().Time,
		Repositories: []*repo.Entry{},
	}
	data, err := yaml.Marshal(&repoData)
	if err != nil {
		assert.Fail(t, "Unable to marshal repo config yaml")
	}
	err2 := os.WriteFile(settings.RepositoryConfig, data, 0666)
	if err2 != nil {
		assert.Fail(t, "Unable to write repo config file")
	}
	return settings
}

func getChartRepo(settings *cli.EnvSettings) *repo.ChartRepository {
	repoURL := "http://repoURL"
	c := repo.Entry{
		Name: filepath.Dir(settings.RepositoryConfig),
		URL:  repoURL,
	}
	r, _ := repo.NewChartRepository(&c, getter.All(settings))
	r.Client = getCLientMock(repoURL)
	return r
}

func getCLientMock(repoURL string) *mocks.Getter {
	clientMock := &mocks.Getter{}
	b := []byte("apiVersion: v1\nentries: {}\ngenerated: \"2022-10-28T09:14:08+02:00\"\nserverInfo: {}\n")
	z := bytes.NewBuffer(b)
	clientMock.On("Get", repoURL+"/index.yaml", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(z, nil)
	return clientMock
}

func containsRepo(repoFile, repoName string) bool {
	file, err := os.Open(repoFile)
	if err != nil {
		log.Fatalf("Error opening repository file: %v", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		row := scanner.Text()
		if strings.Contains(row, repoName) {
			return true
		}
	}
	return false
}
