package marvel

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

const charactersURL = "https://gateway.marvel.com/v1/public/characters"

type API struct {
	httpClient        *http.Client
	CharacterEndpoint string // Define the character endpoint. maybe remove and find more elegant solution for testing.
}

type Character struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Thumbnail   struct {
		Path      string `json:"path"`
		Extension string `json:"extension"`
	} `json:"thumbnail"`
	Urls []struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"urls"`
	Comics struct {
		Available int `json:"available"`
	} `json:"comics"`
}

type Criteria struct {
	Limit   int
	Offset  int
	OrderBy string
}

// Marvel *sometimes* returns a completely different JSON structure if is an error.
// But for invalid input or 409's they return a regular result.
type ErrorResult struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Result struct {
	Code            int    `json:"code"`
	Status          string `json:"status"`
	Copyright       string `json:"copyright"`
	AttributionText string `json:"attributionText"`
	ETag            string `json:"ETag"`
}

type Container struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Total  int `json:"total"`
	Count  int `json:"count"`
}

type CharactersResultContainer struct {
	Container
	Results []*Character `json:"results"`
}

type CharactersResultWrapper struct {
	Result
	CharactersResultContainer `json:"data"`
}

func (api *API) TotalCharacters() (int, error) {
	result, resultError, err := api.Characters(&Criteria{
		Limit: 1,
	})

	if resultError != nil {
		return 0, errors.New(resultError.Message)
	}
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// Returns the API response for getting characters, an error from the API result, or a system-related error.
func (api *API) Characters(criteria *Criteria) (*CharactersResultWrapper, *ErrorResult, error) {
	var apiResponse = new(CharactersResultWrapper)
	var url string
	if api.CharacterEndpoint == "" {
		url = charactersURL
	} else {
		url = api.CharacterEndpoint
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return apiResponse, nil, err
	}
	q := request.URL.Query()
	q.Add("apikey", os.Getenv("CC_MARVEL_PUBLIC_KEY"))
	q.Add("ts", "1")
	hash := md5.Sum([]byte(fmt.Sprintf("%d%s%s", 1, os.Getenv("CC_MARVEL_PRIVATE_KEY"), os.Getenv("CC_MARVEL_PUBLIC_KEY"))))
	q.Add("hash", fmt.Sprintf("%x", hash))
	q.Add("limit", fmt.Sprintf("%d", criteria.Limit))
	q.Add("offset", fmt.Sprintf("%d", criteria.Offset))
	q.Add("orderBy", criteria.OrderBy)
	request.URL.RawQuery = q.Encode()
	response, err := api.httpClient.Do(request)
	if response.Body != nil {
		defer response.Body.Close()
	}
	if err != nil {
		return apiResponse, nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return apiResponse, nil, err
	}
	// a map container to decode the JSON structure into
	jsonBody := make(map[string]interface{})
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, nil, err
	}
	// This is a stupid fix for the API's bad structure of *SOMETIMES* returning an error with a "message"
	// key if an error happens.
	// API only returns 2-3 fields when there's an error, so check the length before we iterate
	// over all the keys of the JSON.
	if len(jsonBody) <= 3 {
		for k := range jsonBody {
			if k == "message" {
				errorResult := new(ErrorResult)
				err = json.Unmarshal(body, &errorResult)
				return nil, errorResult, err
			}
		}
	}
	err = json.Unmarshal(body, &apiResponse)
	return apiResponse, nil, err
}

func NewMarvelAPI(client *http.Client) *API {
	return &API{
		httpClient: client,
	}
}
