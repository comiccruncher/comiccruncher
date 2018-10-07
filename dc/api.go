package dc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const charactersUrl = "https://www.dccomics.com/proxy/search"
const ApiUrl = "https://www.dccomics.com"

type Api struct {
	httpClient        *http.Client
	CharacterEndpoint string // Define the characters endpoint. maybe remove and find more elegant solution for testing.
}

type ApiResult struct {
	TotalResults int `json:"result count"`
	Results      map[string]*CharacterResult
}

type CharacterResult struct {
	Id     string          `json:"id"`
	Fields CharacterFields `json:"fields"`
}

type CharacterFields struct {
	Body           []string `json:"body:value"`
	ProfilePicture []string `json:"field_profile_picture:file:url"`
	Name           string   `json:"dc_solr_sortable_title"`
	Url            string   `json:"url"`
}

func (a *Api) TotalCharacters() (int, error) {
	r, err := a.FetchCharacters(1)
	return r.TotalResults, err
}

func (a *Api) FetchCharacters(pageNumber int) (*ApiResult, error) {
	var apiResponse = new(ApiResult)
	request, err := http.NewRequest(http.MethodGet, a.CharacterEndpoint, nil)
	if err != nil {
		return apiResponse, nil
	}
	q := request.URL.Query()
	q.Add("type", "generic_character")
	q.Add("sortBy", "title-ASC")
	q.Add("characterType", "44070")
	q.Add("page", fmt.Sprintf("%d", pageNumber))
	request.URL.RawQuery = q.Encode()
	response, err := a.httpClient.Do(request)
	if response.Body != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return apiResponse, err
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusNotModified {
		return apiResponse, errors.New(fmt.Sprintf("got bad status code from %s: %d", a.CharacterEndpoint, response.StatusCode))
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return apiResponse, err
	}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return apiResponse, err
	}
	return apiResponse, nil
}

func NewDcApi(httpClient *http.Client) *Api {
	return &Api{
		httpClient:        httpClient,
		CharacterEndpoint: charactersUrl,
	}
}
