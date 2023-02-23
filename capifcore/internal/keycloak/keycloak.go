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

	"oransc.org/nonrtric/capifcore/internal/config"
)

//go:generate mockery --name AccessManagement
type AccessManagement interface {
	// Get JWT token for a client.
	// Returns JWT token if client exits and credentials are correct otherwise returns error.
	GetToken(clientId, clientPassword, scope string, realm string) (Jwttoken, error)
}

type KeycloakManager struct {
	keycloakServerUrl string
	realms            map[string]string
}

func NewKeycloakManager(cfg *config.Config) *KeycloakManager {

	keycloakUrl := "http://" + cfg.AuthorizationServer.Host + ":" + cfg.AuthorizationServer.Port

	return &KeycloakManager{
		keycloakServerUrl: keycloakUrl,
		realms:            cfg.AuthorizationServer.Realms,
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

func (km *KeycloakManager) GetToken(clientId, clientPassword, scope string, realm string) (Jwttoken, error) {
	var jwt Jwttoken
	getTokenUrl := km.keycloakServerUrl + "/realms/" + realm + "/protocol/openid-connect/token"

	resp, err := http.PostForm(getTokenUrl,
		url.Values{"grant_type": {"client_credentials"}, "client_id": {clientId}, "client_secret": {clientPassword}})

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
