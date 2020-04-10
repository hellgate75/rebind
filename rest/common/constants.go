package common

type WebMethod string

type MediaType string

const (
	GET_WEB_METHOD    WebMethod = "GET"
	POST_WEB_METHOD   WebMethod = "POST"
	PUT_WEB_METHOD    WebMethod = "PUT"
	DELETE_WEB_METHOD WebMethod = "DELETE"
	JSON_MEDIA_TYPE   MediaType = "application/json"
	YAML_MEDIA_TYPE   MediaType = "text/yaml"
	XML_MEDIA_TYPE    MediaType = "application/xml"
	PLAIN_MEDIA_TYPE  MediaType = "plain/text"
)
