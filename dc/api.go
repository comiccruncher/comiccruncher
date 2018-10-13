package dc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const charactersURL = "https://www.dccomics.com/proxy/search"

// APIURL is the base url for the API.
const APIURL = "https://www.dccomics.com"

// API is the client to communicate with the DC API.
type API struct {
	httpClient        *http.Client
	CharacterEndpoint string // Define the characters endpoint. maybe remove and find more elegant solution for testing.
}

// APIResult contains the result of the API.
type APIResult struct {
	TotalResults int `json:"result count"`
	Results      map[string]*CharacterResult
}

// CharacterResult represents the result of each character.
type CharacterResult struct {
	ID     string          `json:"id"`
	Fields CharacterFields `json:"fields"`
}

// CharacterFields represent the fields for a character.
type CharacterFields struct {
	Body           []string `json:"body:value"`
	ProfilePicture []string `json:"field_profile_picture:file:url"`
	Name           string   `json:"dc_solr_sortable_title"`
	URL            string   `json:"url"`
}

// TotalCharacters gets the number of characters from the result.
func (a *API) TotalCharacters() (int, error) {
	r, err := a.Characters(1)
	return r.TotalResults, err
}

// Characters gets all the characters.
func (a *API) Characters(pageNumber int) (*APIResult, error) {
	var apiResponse = new(APIResult)
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
		return apiResponse, fmt.Errorf("got bad status code from %s: %d", a.CharacterEndpoint, response.StatusCode)
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

// NewDcAPI creates a new dc api.
func NewDcAPI(httpClient *http.Client) *API {
	return &API{
		httpClient:        httpClient,
		CharacterEndpoint: charactersURL,
	}
}
