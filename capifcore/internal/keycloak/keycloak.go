// -
//   ========================LICENSE_START=================================
//   O-RAN-SC
//   %%
//   Copyright (C) 2023: Nordix Foundation
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

package keycloak

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"oransc.org/nonrtric/capifcore/internal/config"
	"oransc.org/nonrtric/capifcore/internal/restclient"
)

//go:generate mockery --name AccessManagement
type AccessManagement interface {
	// Get JWT token for a client.
	// Returns JWT token if client exits and credentials are correct otherwise returns error.
	GetToken(realm string, data map[string][]string) (Jwttoken, error)
	// Add new client in keycloak
	AddClient(clientId string, realm string) error
}

type AdminUser struct {
	User     string
	Password string
}

type KeycloakManager struct {
	keycloakServerUrl string
	admin             AdminUser
	realms            map[string]string
	client            restclient.HTTPClient
}

func NewKeycloakManager(cfg *config.Config, c restclient.HTTPClient) *KeycloakManager {

	keycloakUrl := "http://" + cfg.AuthorizationServer.Host + ":" + cfg.AuthorizationServer.Port

	return &KeycloakManager{
		keycloakServerUrl: keycloakUrl,
		client:            c,
		admin: AdminUser{
			User:     cfg.AuthorizationServer.AdminUser.User,
			Password: cfg.AuthorizationServer.AdminUser.Password,
		},
		realms: cfg.AuthorizationServer.Realms,
	}
}

type Jwttoken struct {
	AccessToken      string `json:"access_token"`
	IDToken          string `json:"id_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}

func (km *KeycloakManager) GetToken(realm string, data map[string][]string) (Jwttoken, error) {
	var jwt Jwttoken
	getTokenUrl := km.keycloakServerUrl + "/realms/" + realm + "/protocol/openid-connect/token"

	resp, err := http.PostForm(getTokenUrl, data)

	if err != nil {
		return jwt, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return jwt, err
	}
	if resp.StatusCode != http.StatusOK {
		return jwt, errors.New(string(body))
	}

	json.Unmarshal([]byte(body), &jwt)
	return jwt, nil
}

type Client struct {
	AdminURL               string `json:"adminUrl,omitempty"`
	BearerOnly             bool   `json:"bearerOnly,omitempty"`
	ClientID               string `json:"clientId,omitempty"`
	Enabled                bool   `json:"enabled,omitempty"`
	PublicClient           bool   `json:"publicClient,omitempty"`
	RootURL                string `json:"rootUrl,omitempty"`
	ServiceAccountsEnabled bool   `json:"serviceAccountsEnabled,omitempty"`
}

func (km *KeycloakManager) AddClient(clientId string, realm string) error {
	data := url.Values{"grant_type": {"password"}, "username": {km.admin.User}, "password": {km.admin.Password}, "client_id": {"admin-cli"}}
	token, err := km.GetToken("master", data)
	if err != nil {
		log.Errorf("error wrong credentials or url %v\n", err)
		return err
	}

	createClientUrl := km.keycloakServerUrl + "/admin/realms/" + realm + "/clients"
	newClient := Client{
		ClientID:               clientId,
		Enabled:                true,
		ServiceAccountsEnabled: true,
		BearerOnly:             false,
		PublicClient:           false,
	}

	body, _ := json.Marshal(newClient)
	var headers = map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + token.AccessToken}
	if error := restclient.Post(createClientUrl, body, headers, km.client); error != nil {
		log.Errorf("error with http request: %+v\n", err)
		return err
	}

	log.Info("Created new client")
	return nil

}
