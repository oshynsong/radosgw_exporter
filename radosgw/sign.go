//sign.go - defines the AWS S3 signature functionility

package radosgw

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	signAlgorithm = "AWS4-HMAC-SHA256"
	signRegion    = "us-east-1"
	signService   = "s3"
	signRequest   = "aws4_request"
)

// Sign - generates the authorization string with a http request and sign option.
//
// PARAMS:
//     - request: *http.Request for this sign
//     - accessKeyId: the access key id for this sign
//     - secretAccessKey: the secret access key for this sign
// RETURN:
//     - request: the signed http request with authorization headers set
func Sign(request *http.Request, accessKeyId, secretAccessKey string) *http.Request {
	// Step 1. create canonical request
	//   The canonical request structure is
	//       HTTPRequestMethod + '\n' +
	//       CanonicalURI + '\n' +
	//       CanonicalQueryString + '\n' +
	//       CanonicalHeaders + '\n' +
	//       SignedHeaders + '\n' +
	//       HexEncode(Hash(RequestPayload))
	// reference https://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
	canonicalRequestMethod := request.Method

	// Generate canonical request uri
	uri := request.URL.Path
	if len(uri) == 0 {
		uri = "/"
	}
	if strings.HasPrefix(uri, "/") {
		uri = uri[1:]
	}
	canonicalRequestUri := "/" + uriEncode(uri, false)

	// Generate canonical request query string
	queryString := request.URL.Query().Encode()
	canonicalQueryString := strings.Replace(queryString, "+", "%20", -1)

	// Set the headers and generate the canonical headers, signed headers and payload
	var payload []byte
	if request.Body != nil {
		payload, _ = ioutil.ReadAll(request.Body)
		request.Body = ioutil.NopCloser(bytes.NewReader(payload))
	}
	hashedPayload := hashSha256Hex(payload)
	request.Header.Set("X-Amz-Content-Sha256", hashedPayload)
	date := signDateString()
	if request.Header.Get("X-Amz-Date") == "" {
		request.Header.Set("X-Amz-Date", date)
	}
	if request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	}
	request.Header.Set("Host", request.Host)

	var sortedHeaderKeys []string
	for key, _ := range request.Header {
		switch key {
		case "Content-Type", "Content-Md5", "Host":
		default:
			if !strings.HasPrefix(key, "X-Amz-") {
				continue
			}
		}
		sortedHeaderKeys = append(sortedHeaderKeys, strings.ToLower(key))
	}
	sort.Strings(sortedHeaderKeys)

	var headersToSign string
	for _, key := range sortedHeaderKeys {
		value := strings.TrimSpace(request.Header.Get(key))
		if key == "host" {
			//AWS does not include port in signing request.
			if strings.Contains(value, ":") {
				split := strings.Split(value, ":")
				port := split[1]
				if port == "80" || port == "443" {
					value = split[0]
				}
			}
		}
		headersToSign += key + ":" + value + "\n"
	}
	signedHeaders := strings.Join(sortedHeaderKeys, ";")
	canonicalRequest := strings.Join([]string{
		canonicalRequestMethod,
		canonicalRequestUri,
		canonicalQueryString,
		headersToSign,
		signedHeaders,
		hashedPayload,
	}, "\n")
	hashedCanonicalRequest := hashSha256Hex([]byte(canonicalRequest))

	// Step 2. create the string to sign
	//   StringToSign =
	//       Algorithm + \n +
	//       RequestDateTime + \n +
	//       CredentialScope + \n +
	//       HashedCanonicalRequest
	credentialScope := fmt.Sprintf("%s/%s/%s/%s", date[:8], signRegion, signService, signRequest)
	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		signAlgorithm, date, credentialScope, hashedCanonicalRequest)

	// Step 3. calculate the signature
	kSecret := secretAccessKey
	kDate := hmacSha256([]byte("AWS4"+kSecret), date[:8])
	kRegion := hmacSha256(kDate, signRegion)
	kService := hmacSha256(kRegion, signService)
	kSigning := hmacSha256(kService, signRequest)
	signature := hex.EncodeToString(hmacSha256(kSigning, stringToSign))

	// Step 4. generate the authorization string
	authorizationString := fmt.Sprintf("%s Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		signAlgorithm, accessKeyId, credentialScope, signedHeaders, signature)

	request.Header.Set("Authorization", authorizationString)
	return request
}

func hmacSha256(key []byte, strToSign string) []byte {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(strToSign))
	return hasher.Sum(nil)
}

func hashSha256Hex(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}

func signDateString() string {
	utc := time.Now().UTC()
	return utc.Format("20060102T150405Z")
}

func uriEncode(uri string, encodeSlash bool) string {
	var byteBuf bytes.Buffer
	for _, b := range []byte(uri) {
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') ||
			b == '-' || b == '_' || b == '.' || b == '~' || (b == '/' && !encodeSlash) {
			byteBuf.WriteByte(b)
		} else {
			byteBuf.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}
	return byteBuf.String()
}
