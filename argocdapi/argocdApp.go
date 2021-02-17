package argocdapis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type argocdapp struct {
	Metadata Metadata `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type Source struct {
	RepoURL        string `json:"repoURL"`
	TargetRevision string `json:"targetRevision"`
	Chart          string `json:"chart"`
}
type Destination struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type Automated struct {
	Prune      bool `json:"prune"`
	SelfHeal   bool `json:"selfHeal"`
	AllowEmpty bool `json:"allowEmpty"`
}
type SyncPolicy struct {
	Automated *Automated `json:",omitempty"`
}
type Spec struct {
	Project     string      `json:"project"`
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`
	SyncPolicy  *SyncPolicy `json:",omitempty"`
	SyncOptions []string    `json:"syncOptions"`
}

type appResponse struct {
	Token string `json:"token"`
}

type argocdappstatus struct {
	Status struct {
		Health struct {
			Status string `json:"status"`
		} `json:"health"`
	} `json:"status"`
}

func NewArgocdApp(appName string, artifactoryURL string, targetVersion string, chartName string, namespace string) argocdapp {

	newapp := &argocdapp{
		Metadata: Metadata{
			Name:      appName,
			Namespace: "argocd",
		},
		Spec: Spec{
			Project: "default",
			Source: Source{
				RepoURL:        artifactoryURL,
				TargetRevision: targetVersion,
				Chart:          chartName,
			},
			Destination: Destination{
				Name:      "in-cluster",
				Namespace: namespace,
			},
			SyncOptions: []string{"CreateNamespace=true"},
		},
	}
	return *newapp
}

func GetToken(client http.Client, Username string, Password string, url string) string {

	jsonData := map[string]string{"username": Username, "password": Password}
	jsonValue, _ := json.Marshal(jsonData)

	tokenURL := url + "/api/v1/session"

	response, err := client.Post(tokenURL, "application/json", bytes.NewBuffer(jsonValue))

	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	}

	data, _ := ioutil.ReadAll(response.Body)

	var responseObject appResponse
	json.Unmarshal(data, &responseObject)

	return responseObject.Token
}

func (a argocdapp) CreateArgocdApp(client http.Client, token string, url string) {

	createURL := url + "/api/v1/applications"
	method := "POST"
	jsonReq, _ := json.Marshal(a)
	payload := bytes.NewBuffer(jsonReq)

	req, err := http.NewRequest(method, createURL, payload)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
}

func (a argocdapp) UpdateArgocdApp(client http.Client, token string, url string) {

	updateURL := url + "/api/v1/applications/" + a.Metadata.Name
	method := "PUT"
	jsonReq, err := json.Marshal(a)
	payload := bytes.NewBuffer(jsonReq)
	bearerToken := "Bearer " + token

	req, err := http.NewRequest(method, updateURL, payload)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Authorization", bearerToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))

}

func (a argocdapp) GetArgocdAppStatus(client http.Client, token string, url string) string {

	getURL := url + "/api/v1/applications/" + a.Metadata.Name
	method := "GET"

	req, err := http.NewRequest(method, getURL, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	var heathofapp argocdappstatus
	json.Unmarshal(body, &heathofapp)

	return heathofapp.Status.Health.Status
}

func (a argocdapp) DeleteArgocdApp(client http.Client, token string, url string) {

	deleteURL := url + "/api/v1/applications/" + a.Metadata.Name
	method := "DELETE"

	req, err := http.NewRequest(method, deleteURL, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+token)

	q := req.URL.Query()
	q.Add("cascade", "true")
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	fmt.Println(res.StatusCode)
}

