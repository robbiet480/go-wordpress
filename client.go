package wordpress

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"

	"github.com/parnurzeal/gorequest"
)

const (
	CollectionUsers      = "users"
	CollectionPosts      = "posts"
	CollectionPages      = "pages"
	CollectionMedia      = "media"
	CollectionMeta       = "meta"
	CollectionRevisions  = "revisions"
	CollectionComments   = "comments"
	CollectionTaxonomies = "taxonomies"
	CollectionTerms      = "terms"
	CollectionStatuses   = "statuses"
	CollectionTypes      = "types"
	CollectionSettings   = "settings"
	CollectionCategories = "categories"
	CollectionTags       = "tags"
)

type GeneralError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    int    `json:"data"` // Unsure if this is consistent
}

type Options struct {
	BaseAPIURL string

	// Basic Auth
	Username string
	Password string

	JWTToken string
	// TODO: support OAuth authentication
}

type Client struct {
	req     *gorequest.SuperAgent
	options *Options
	baseURL string
}

// Used to create a new SuperAgent object.
func newHTTPClient() *gorequest.SuperAgent {
	client := gorequest.New()
	client.Client = &http.Client{Jar: nil}
	client.Transport = &http.Transport{
		DisableKeepAlives: true,
	}
	return client
}

func NewClient(options *Options) *Client {
	req := newHTTPClient()

	if options.Username != "" && options.Password != "" {
		req = req.SetBasicAuth(options.Username, options.Password)
	} else if options.JWTToken != "" {
		req = req.AppendHeader("Authorization", options.JWTToken)
	}

	req = req.RedirectPolicy(func(r gorequest.Request, via []gorequest.Request) error {
		// perform BasicAuth on each redirect request.
		// (requests are cookie-less; so we need to keep re-auth-ing again)
		httpReq := http.Request(*r)
		httpReq.SetBasicAuth(options.Username, options.Password)
		log.Println("REDIRECT", r, options.Username, options.Password)
		return nil
	})
	return &Client{
		req:     req,
		options: options,
		baseURL: options.BaseAPIURL,
	}
}

func (client *Client) Users() *UsersCollection {
	return &UsersCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionUsers),
	}
}
func (client *Client) Posts() *PostsCollection {
	return &PostsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionPosts),
	}
}
func (client *Client) Pages() *PagesCollection {
	return &PagesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionPages),
	}
}
func (client *Client) Media() *MediaCollection {
	return &MediaCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionMedia),
	}
}
func (client *Client) Comments() *CommentsCollection {
	return &CommentsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionComments),
	}
}
func (client *Client) Taxonomies() *TaxonomiesCollection {
	return &TaxonomiesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTaxonomies),
	}
}
func (client *Client) Terms() *TermsCollection {
	return &TermsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTerms),
	}
}
func (client *Client) Statuses() *StatusesCollection {
	return &StatusesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionStatuses),
	}
}
func (client *Client) Types() *TypesCollection {
	return &TypesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTypes),
	}
}
func (client *Client) Settings() *SettingsCollection {
	return &SettingsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionSettings),
	}
}
func (client *Client) Categories() *CategoriesCollection {
	return &CategoriesCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionCategories),
	}
}
func (client *Client) Tags() *TagsCollection {
	return &TagsCollection{
		client: client,
		url:    fmt.Sprintf("%v/%v", client.baseURL, CollectionTags),
	}
}

func (client *Client) List(url string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	client.req.TargetType = "json"

	req := client.req.Get(url)
	if client.options.JWTToken != "" {
		req.Set("Authorization", client.options.JWTToken)
	}
	resp, body, errSlice := req.Query(params).EndBytes()

	if errSlice != nil && len(errSlice) > 0 {
		return nil, body, errSlice[len(errSlice)-1]
	}
	err := unmarshallResponse(resp, body, result)
	_resp := http.Response(*resp)
	return &_resp, body, err
}

func (client *Client) Create(url string, content interface{}, result interface{}) (*http.Response, []byte, error) {
	contentVal := unpackInterfacePointer(content)
	client.req.TargetType = "json"
	req := client.req.Post(url).Send(contentVal)
	if client.options.JWTToken != "" {
		req.Set("Authorization", client.options.JWTToken)
	}
	resp, body, errSlice := req.EndBytes()
	if errSlice != nil && len(errSlice) > 0 {
		return nil, body, errSlice[len(errSlice)-1]
	}
	err := unmarshallResponse(resp, body, result)
	_resp := http.Response(*resp)
	return &_resp, body, err
}
func (client *Client) Get(url string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	client.req = client.req.AppendHeader("Authorization", client.options.JWTToken)

	client.req.TargetType = "json"
	req := client.req.Get(url)
	if client.options.JWTToken != "" {
		req.Set("Authorization", client.options.JWTToken)
	}
	resp, body, errSlice := req.Query(params).EndBytes()

	if errSlice != nil && len(errSlice) > 0 {
		return nil, body, errSlice[len(errSlice)-1]
	}
	err := unmarshallResponse(resp, body, result)
	_resp := http.Response(*resp)

	return &_resp, body, err
}
func (client *Client) Update(url string, content interface{}, result interface{}) (*http.Response, []byte, error) {

	contentVal := unpackInterfacePointer(content)

	client.req.TargetType = "json"
	req := client.req.Post(url).Send(contentVal)
	req.Set("HTTP_X_HTTP_METHOD_OVERRIDE", "PUT")

	if client.options.JWTToken != "" {
		req.Set("Authorization", client.options.JWTToken)
	}

	resp, body, errSlice := req.EndBytes()
	if errSlice != nil && len(errSlice) > 0 {
		return nil, body, errSlice[len(errSlice)-1]
	}
	err := unmarshallResponse(resp, body, result)
	_resp := http.Response(*resp)
	return &_resp, body, err
}
func (client *Client) Delete(url string, params interface{}, result interface{}) (*http.Response, []byte, error) {
	client.req.TargetType = "json"
	req := client.req.Get(url).Query(params).Query("_method=DELETE")
	req.Set("HTTP_X_HTTP_METHOD_OVERRIDE", "DELETE")

	if client.options.JWTToken != "" {
		req.Set("Authorization", client.options.JWTToken)
	}

	resp, body, errSlice := req.End()
	by := []byte(body)
	if errSlice != nil && len(errSlice) > 0 {
		return resp, by, errSlice[len(errSlice)-1]
	}
	err := unmarshallResponse(resp, by, result)
	_resp := http.Response(*resp)
	return &_resp, by, err
}
func (client *Client) PostData(url string, content []byte, contentType string, filename string, result interface{}) (*http.Response, []byte, error) {

	// gorequest does not support POST-ing raw data
	// so, we have to manually create a HTTP client
	s := client.req.Post(url)

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fileField, _ := w.CreateFormFile("file", filename)
	fileField.Write(content)
	w.Close()

	req, err := http.NewRequest(s.Method, s.Url, &buf)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Content-Disposition", fmt.Sprintf("filename=%v", filename))

	if client.options.JWTToken != "" {
		req.Header.Set("Authorization", client.options.JWTToken)
	}
	// Add basic auth
	req.SetBasicAuth(s.BasicAuth.Username, s.BasicAuth.Password)

	// Set Transport
	s.Client.Transport = s.Transport

	// Send request
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	err = unmarshallResponse(resp, body, result)
	_resp := http.Response(*resp)
	return &_resp, body, err
}

func unpackInterfacePointer(content interface{}) interface{} {
	val := reflect.ValueOf(content)
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.IsValid() {
		return val.Interface()
	}
	return nil
}
