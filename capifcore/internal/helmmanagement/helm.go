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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/repo"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

//go:generate mockery --name HelmManager
type HelmManager interface {
	AddToRepo(repoName, url string) error
	InstallHelmChart(namespace, repoName, chartName, releaseName string) error
	UninstallHelmChart(namespace, chartName string)
}

type helmManagerImpl struct {
	settings *cli.EnvSettings
}

func NewHelmManager(s *cli.EnvSettings) *helmManagerImpl {
	return &helmManagerImpl{
		settings: s,
	}
}

func (hm *helmManagerImpl) AddToRepo(repoName, url string) error {
	repoFile := hm.settings.RepositoryConfig

	//Ensure the file directory exists as it is required for file locking
	err := os.MkdirAll(filepath.Dir(repoFile), os.ModePerm)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	b, err := os.ReadFile(repoFile)
	if err != nil {
		return err
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}

	if f.Has(repoName) {
		log.Debugf("repository name (%s) already exists\n", repoName)
		return nil
	}

	c := repo.Entry{
		Name: repoName,
		URL:  url,
	}

	r, err := repo.NewChartRepository(&c, getter.All(hm.settings))
	if err != nil {
		return err
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		err := errors.Wrapf(err, "looks like %q is not a valid chart repository or cannot be reached", url)
		return err
	}

	f.Update(&c)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return err
	}
	log.Debugf("%q has been added to your repositories\n", repoName)
	return nil
}

func (hm *helmManagerImpl) InstallHelmChart(namespace, repoName, chartName, releaseName string) error {
	actionConfig, err := getActionConfig(namespace)
	if err != nil {
		return err
	}

	install := action.NewInstall(actionConfig)

	cp, err := install.ChartPathOptions.LocateChart(fmt.Sprintf("%s/%s", repoName, chartName), hm.settings)
	if err != nil {
		log.Error("Unable to locate chart!")
		return err
	}

	chartRequested, err := loader.Load(cp)
	if err != nil {
		log.Error("Unable to load chart path!")
		return err
	}

	install.Namespace = namespace
	install.ReleaseName = releaseName
	_, err = install.Run(chartRequested, nil)
	if err != nil {
		log.Error("Unable to run chart!")
		return err
	}
	log.Debug("Successfully onboarded ", namespace, repoName, chartName, releaseName)
	return nil
}

func (hm *helmManagerImpl) UninstallHelmChart(namespace, chartName string) {
	actionConfig, err := getActionConfig(namespace)
	if err != nil {
		log.Error("unable to get action config: ", err)
		return
	}

	iCli := action.NewUninstall(actionConfig)

	resp, err := iCli.Run(chartName)
	if err != nil {
		log.Error("Unable to uninstall chart ", chartName, err)
		return
	}
	log.Debug("Successfully uninstalled chart: ", resp.Release.Name)
}

func getActionConfig(namespace string) (*action.Configuration, error) {
	actionConfig := new(action.Configuration)
	// Create the rest config instance with ServiceAccount values loaded in them
	config, err := rest.InClusterConfig()
	if err != nil {
		// fallback to kubeconfig
		home, exists := os.LookupEnv("HOME")
		if !exists {
			home = "/root"
		}
		kubeconfigPath := filepath.Join(home, ".kube", "config")
		if envvar := os.Getenv("KUBECONFIG"); len(envvar) > 0 {
			kubeconfigPath = envvar
		}
		if err := actionConfig.Init(kube.GetConfig(kubeconfigPath, "", namespace), namespace, os.Getenv("HELM_DRIVER"), log.Debugf); err != nil {
			log.Error(err)
		}
	} else {
		// Create the ConfigFlags struct instance with initialized values from ServiceAccount
		kubeConfig := genericclioptions.NewConfigFlags(false)
		kubeConfig.APIServer = &config.Host
		kubeConfig.BearerToken = &config.BearerToken
		kubeConfig.CAFile = &config.CAFile
		kubeConfig.Namespace = &namespace
		if err := actionConfig.Init(kubeConfig, namespace, os.Getenv("HELM_DRIVER"), log.Debugf); err != nil {
			log.Error(err)
		}
	}
	return actionConfig, err
}
