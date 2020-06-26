package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/trezorg/lingualeo/pkg/logger"
)

func getJSONFromString(body *string, target interface{}) error {
	return json.Unmarshal([]byte(*body), &target)
}

func readBody(resp *http.Response) (*string, error) {
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := string(data)
	return &res, nil
}
