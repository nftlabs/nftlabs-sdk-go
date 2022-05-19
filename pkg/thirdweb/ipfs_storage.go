package thirdweb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"mime/multipart"
	"net/http"
	"strings"
)

type baseUriWithUris struct {
	baseUri string
	uris    []string
}
type storage interface {
	Get(uri string) ([]byte, error)
	Upload(data any, contractAddress string, signerAddress string) (string, error)
	UploadBatch(data []any, contractAddress string, signerAddress string) (*baseUriWithUris, error)
}

type uploadResponse struct {
	IpfsHash    string   `json:"ipfsHash"`
	PinSize     *big.Int `json:"PinSize"`
	Timestamp   string   `json:"Timestamp"`
	IsDuplicate bool     `json:"isDuplicate"`
	IpfsUri     string   `json:"IpfsUri"`
}

type ipfsStorage struct {
	Url string
}

func newIpfsStorage(uri string) storage {
	return &ipfsStorage{
		Url: uri,
	}
}

func (gw *ipfsStorage) Get(uri string) ([]byte, error) {
	gatewayUrl := replaceIpfsPrefixWithGateway(uri, gw.Url)
	resp, err := http.Get(gatewayUrl)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Bad status code, %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	return body, nil
}

// Upload method can be used to upload a generic payload to IPFS. NftLabs provides a default proxy
// in the SDK. You can override this with the ISdk.SetStorage
func (ipfs *ipfsStorage) Upload(data any, contractAddress string, signerAddress string) (string, error) {
	baseUriWithUris, err := ipfs.UploadBatch([]any{data}, contractAddress, signerAddress)
	if err != nil {
		return "", err
	}

	baseUri := baseUriWithUris.baseUri + "0"
	return baseUri, nil
}

// UploadBatch uploads a list of arbitrary objects and returns their URIs *in the order they were passed*
func (ipfs *ipfsStorage) UploadBatch(data []any, contractAddress string, signerAddress string) (*baseUriWithUris, error) {
	baseUriWithUris, err := ipfs.uploadBatchWithCid(data, contractAddress, signerAddress)
	if err != nil {
		return nil, err
	}

	return baseUriWithUris, nil
}

func (ipfs *ipfsStorage) getUploadToken(contractAddress string) (string, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("%v/grant", twIpfsServerUrl), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-App-Name", fmt.Sprintf("CONSOLE-GO-SDK-%v", contractAddress))
	result, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if result.StatusCode != http.StatusOK {
		return "", &failedToUploadError{
			statusCode: result.StatusCode,
		}
	}

	body, err := ioutil.ReadAll(result.Body)
	text := string(body)

	return text, nil
}

func (ipfs *ipfsStorage) uploadBatchWithCid(
	data []any,
	contractAddress string,
	signerAddress string,
) (*baseUriWithUris, error) {
	uploadToken, err := ipfs.getUploadToken(contractAddress)

	client := &http.Client{}
	fileNames := []string{}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err != nil {
		return nil, err
	}

	for i, asset := range data {
		jsonData, err := json.Marshal(asset)
		if err != nil {
			return nil, err
		}

		fileName := fmt.Sprintf("%v", i)
		fileNames = append(fileNames, fileName)

		part, err := writer.CreateFormFile("file", fmt.Sprintf("files/%v", fileName))
		part.Write(jsonData)
	}

	_ = writer.Close()

	req, err := http.NewRequest("POST", pinataIpfsUrl, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", uploadToken))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if result, err := client.Do(req); err != nil {
		return nil, err
	} else {
		if result.StatusCode != http.StatusOK {
			return nil, &failedToUploadError{
				statusCode: result.StatusCode,
				Payload:    data,
			}
		}

		var uploadMeta uploadResponse
		bodyBytes, err := ioutil.ReadAll(result.Body)
		if err != nil {
			return nil, &failedToUploadError{
				statusCode:      result.StatusCode,
				Payload:         data,
				UnderlyingError: err,
			}
		}

		if err := json.Unmarshal(bodyBytes, &uploadMeta); err != nil {
			return nil, &unmarshalError{
				body:            string(bodyBytes),
				typeName:        "UploadResponse",
				UnderlyingError: err,
			}
		}

		baseUri := "ipfs://" + uploadMeta.IpfsHash + "/"

		uris := []string{}
		for _, fileName := range fileNames {
			uri := baseUri + fileName
			uris = append(uris, uri)
		}

		return &baseUriWithUris{
			baseUri: baseUri,
			uris:    uris,
		}, nil
	}
}

func replaceIpfsPrefixWithGateway(ipfsUrl string, gatewayUrl string) string {
	if ipfsUrl == "" {
		return ""
	}

	gateway := gatewayUrl
	if !strings.HasSuffix(gatewayUrl, "/") {
		gateway = gatewayUrl + "/"
	}

	return strings.Replace(ipfsUrl, "ipfs://", gateway, 1)
}
