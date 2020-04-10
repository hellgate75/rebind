package client

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/hellgate75/rebind/rest/common"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
)

type ResponseHandler func(code int, message string, mdiaType common.MediaType,
	content []byte)

type ApiClient interface{}

type apiClient struct {
	baseUrl string
}

func (c *apiClient) Send(method common.WebMethod, path string,
	acceptType common.MediaType, contentType common.MediaType,
	body interface{}, respnseHandler ResponseHandler) error {
	jsonValue, err := encodeMediaType(contentType, body)
	if err != nil {
		return err
	}
	request, _ := http.NewRequest("POST", fmt.Sprintf("%s%s", c.baseUrl, path),
		bytes.NewBuffer(jsonValue))
	if contentType != "" {
		request.Header.Set("Content-Type", string(contentType))
	}
	if acceptType != "" {
		request.Header.Set("Accepts", string(acceptType))
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err
	} else {
		data, rErr := ioutil.ReadAll(response.Body)
		if rErr != nil {
			return rErr
		}
		if respnseHandler != nil {
			cType := ""
			if cth := response.Header.Get("Content-Type"); cth != "" {
				cType = cth
			} else if cth := response.Header.Get("content-type"); cth != "" {
				cType = cth
			} else {
				cType = string(acceptType)
			}
			respnseHandler(response.StatusCode,
				response.Status,
				common.MediaType(cType),
				data)
		} else {
			fmt.Println("Response:", string(data))
			return errors.New("Null response handler, cannot handle the response ...")
		}
	}
	return nil
}

func (c *apiClient) EncodeElement(mType common.MediaType, body interface{}) ([]byte, error) {
	return encodeMediaType(mType, body)
}

func (c *apiClient) DecodeElement(mType common.MediaType, content []byte, body interface{}) (interface{}, error) {
	return decodeMediaType(mType, body, content)
}

func encodeMediaType(mType common.MediaType, subject interface{}) ([]byte, error) {
	switch mType {
	case common.JSON_MEDIA_TYPE:
		bts, err := json.Marshal(subject)
		if err != nil {
			return []byte{}, err
		}
		return bts, nil
	case common.YAML_MEDIA_TYPE:
		bts, err := yaml.Marshal(subject)
		if err != nil {
			return []byte{}, err
		}
		return bts, nil
	case common.XML_MEDIA_TYPE:
		bts, err := xml.Marshal(subject)
		if err != nil {
			return []byte{}, err
		}
		return bts, nil
	case common.PLAIN_MEDIA_TYPE:
		return []byte(fmt.Sprintf("%v", subject)), nil
	default:
		return []byte{}, errors.New(fmt.Sprintf("Unable to parse type: %v", mType))
	}
}

func decodeMediaType(mType common.MediaType, target interface{}, content []byte) (interface{}, error) {
	switch mType {
	case common.JSON_MEDIA_TYPE:
		err := json.Unmarshal(content, target)
		return target, err
	case common.YAML_MEDIA_TYPE:
		err := yaml.Unmarshal(content, target)
		return target, err
	case common.XML_MEDIA_TYPE:
		err := xml.Unmarshal(content, target)
		return target, err
	case common.PLAIN_MEDIA_TYPE:
		return string(content), nil
	default:
		return []byte{}, errors.New(fmt.Sprintf("Unable to parse type: %v", mType))
	}
}
