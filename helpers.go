package hangups

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func GetAuthHeaders(sapisid string) map[string]string {
	originUrl := "https://talkgadget.google.com"
	timestampMsec := time.Now().Unix() * 1000

	authString := fmt.Sprintf("%d %s %s", timestampMsec, sapisid, originUrl)
	hash := sha1.New()
	hash.Write([]byte(authString))
	hashBytes := hash.Sum(nil)
	hexSha1 := hex.EncodeToString(hashBytes)
	sapisidHash := fmt.Sprintf("SAPISIDHASH %d_%s", timestampMsec, hexSha1)
	return map[string]string{"authorization": sapisidHash, "x-origin": originUrl, "x-goog-authuser": "0"}
}

func ApiRequest(endpointUrl, contentType, responseType, cookies, sapisid string, headers map[string]string, payload []byte) ([]byte, error) {
	authHeaders := GetAuthHeaders(sapisid)
	for headerKey, headerVal := range authHeaders {
		headers[headerKey] = headerVal
	}
	headers["cookie"] = cookies
	headers["content-type"] = contentType
	// This header is required for Protocol Buffer responses, which causes
	// them to be base64 encoded:
	headers["X-Goog-Encode-Response-If-Executable"] = "base64"

	urlObject, _ := url.Parse(endpointUrl)
	urlParams := url.Values{}
	urlParams.Set("alt", responseType)
	urlObject.RawQuery = urlParams.Encode()

	req, err := http.NewRequest("post", urlObject.String(), bytes.NewBufferString(string(payload)))
	for headerKey, headerVal := range headers {
		req.Header.Set(headerKey, headerVal)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return bodyBytes, nil
}
