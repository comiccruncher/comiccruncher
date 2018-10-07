package cerebro

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/aimeelaplant/comiccruncher/comic"
	"github.com/aimeelaplant/comiccruncher/internal/log"
	"github.com/andybalholm/cascadia"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// AlterEgoIdentifier is the struct for identifying an alter ego.
type AlterEgoIdentifier struct {
	httpClient   *http.Client
	characterSvc comic.CharacterServicer
}


func (i *AlterEgoIdentifier) get(url string) (io.ReadCloser, error) {
	resp, err := i.httpClient.Get(url)
	if err != nil {
		return resp.Body, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotModified {
		return resp.Body, errors.New("got bad status code")
	}
	return resp.Body, nil
}

// Parses the alter ego from a URL for a DC character.
func (i *AlterEgoIdentifier) parseDC(url string) (string, error) {
	body, err := i.get(url)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return "", err
	}
	var realName string
	doc.Find(".entity-field-collection-item").Each(func(i int, selection *goquery.Selection) {
		if _, ex := selection.Attr("class"); ex {
			textSelection := strings.TrimSpace(selection.Text())
			if idx := strings.Index(textSelection, "Real Name"); idx != -1 {
				realName = textSelection[idx+9:]
			}
		}
	})
	return realName, nil
}

// Parses the alter ego for a Marvel character.
func (i *AlterEgoIdentifier) parseMarvel(c comic.Character) (string, error) {
	isMain := true
	sources, err := i.characterSvc.Sources(c.ID, comic.VendorTypeCb, &isMain)
	if err != nil {
		return "", err
	}
	var realName string
	for _, source := range sources {
		body, err := i.get(source.VendorUrl)
		if err != nil {
			body.Close()
			return "", err
		}
		doc, err := goquery.NewDocumentFromReader(charmap.ISO8859_1.NewDecoder().Reader(body))
		doc.FindMatcher(cascadia.MustCompile("table[width=\"884\"]")).Each(func(i int, selection *goquery.Selection) {
			selectionText := selection.Text()
			if idx := strings.Index(selectionText, "Real Name: "); idx == -1 {
				return
			}
			for _, s := range strings.SplitAfter(selectionText, "\n") {
				if idx2 := strings.Index(s, "Real Name: "); idx2 != -1 && s != "Real Name: " {
					trimmedS := strings.TrimSpace(s)
					if len(trimmedS) >= idx2+11 {
						realName = trimmedS[idx2+11:]
					}
				}
			}
		})
		body.Close()
		if realName != "" {
			break
		}
	}
	return realName, nil
}

// Name gets the alter-ego name for a character.
func (i *AlterEgoIdentifier) Name(c comic.Character) (string, error) {
	if c.VendorUrl == "" {
		return "", errors.New("empty vendor url")
	}
	var realName string
	var err error
	if c.Publisher.Slug == "dc" {
		realName, err = i.parseDC(c.VendorUrl)
	}
	if c.Publisher.Slug == "marvel" {
		realName, err = i.parseMarvel(c)
	}
	return realName, err
}

// AlterEgoImporter imports the alter ego as an other name for a character
type AlterEgoImporter struct {
	identifier   AlterEgoIdentifier
	characterSvc comic.CharacterServicer
}

// Import imports a character's other_name by identifying a real name from an external source.
func (i *AlterEgoImporter) Import(slugs []comic.CharacterSlug) error {
	 characters, err := i.characterSvc.Characters(slugs, 0, 0)
	 if err != nil {
	 	return err
	 }
	for _, c := range characters {
		if c.OtherName != "" {
			// If a character already has an other name, then don't change it.
			continue
		}
		realName, err := i.identifier.Name(*c)
		if err != nil {
			return err
		}
		if realName == "" {
			continue
		}
		var firstAndLastName string
		// Find the middle name pattern
		if matches := string(regexp.MustCompile(`( '(\w+)' )|( (\w+) )`).Find([]byte(realName))); matches != "" {
			// Strip out the middle name
			firstAndLastName = strings.Replace(realName, string(matches), " ", -1)
		} else {
			firstAndLastName = realName
		}
		// If the character's name isn't the same as the parsed first and last name...
		if c.Name != firstAndLastName {
			// Set the other name.
			c.OtherName = firstAndLastName
			log.CEREBRO().Info("other name for character", zap.String("character", c.Name), zap.String("other name", c.OtherName))
		}
	}

	return i.characterSvc.UpdateAll(characters)

}

// NewAlterEgoImporter creates a new alter ego importer.
func NewAlterEgoImporter(container *comic.PGRepositoryContainer) *AlterEgoImporter {
	svc := comic.NewCharacterService(container)
	return &AlterEgoImporter{
		identifier: AlterEgoIdentifier{
			httpClient:   http.DefaultClient,
			characterSvc: svc,
		},
		characterSvc: svc,
	}
}
