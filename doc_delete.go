package elastic

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/joaosoft/errors"
)

type DeleteResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

type DeleteHit struct {
	Found  bool   `json:"found"`
	Result string `json:"result"`
}

type Delete struct {
	client *Elastic
	index  string
	typ    string
	id     string
	method string
}

func NewDelete(client *Elastic) *Delete {
	return &Delete{
		client: client,
		method: http.MethodDelete,
	}
}

func (e *Delete) Index(index string) *Delete {
	e.index = index
	return e
}

func (e *Delete) Type(typ string) *Delete {
	e.typ = typ
	return e
}

func (e *Delete) Id(id string) *Delete {
	e.id = id
	return e
}

func (e *Delete) Execute() error {

	// delete data from elastic
	var query string
	if e.typ != "" {
		query += fmt.Sprintf("/%s", e.typ)
	}

	if e.id != "" {
		query += fmt.Sprintf("/%s", e.id)
	}

	request, err := http.NewRequest(e.method, fmt.Sprintf("%s/%s%s", e.client.config.Endpoint, e.index, query), nil)
	if err != nil {
		return errors.NewError(err)
	}

	response, err := http.DefaultClient.Do(request)
	defer response.Body.Close()

	// unmarshal data
	body, err := ioutil.ReadAll(response.Body)

	if e.id != "" {
		elasticResponse := DeleteHit{}
		json.Unmarshal(body, &elasticResponse)

		if !elasticResponse.Found || elasticResponse.Result != "deleted" {
			return errors.FromString("couldn't delete the resource")
		}
	} else {
		elasticResponse := DeleteResponse{}
		json.Unmarshal(body, &elasticResponse)

		if !elasticResponse.Acknowledged {
			return errors.FromString("couldn't delete the resource")
		}
	}

	return nil
}
