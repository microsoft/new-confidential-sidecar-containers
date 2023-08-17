// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package filemanager

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Microsoft/confidential-sidecar-containers/pkg/common"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// tokenRefresher is a function callback passed during the creation of token credentials
// its implementation shall update an expired token with a new token and return the new
// expiring duration.
func tokenRefresher(credential azblob.TokenCredential) (t time.Duration) {

	// we extract the audience from the existing token so that we can set the resource
	// id for retrieving a new (refresh) token  for the same audience.
	currentToken := credential.Token()
	// JWT tokens comprise three fields. the second field is the payload (or claims).
	// we care about the `aud` attribute of the payload
	curentTokenFields := strings.Split(currentToken, ".")
	logrus.Infof("Current token fields: %v", curentTokenFields)

	payload, err := base64.RawURLEncoding.DecodeString(curentTokenFields[1])
	if err != nil {
		logrus.Errorf("Error decoding base64 token payload: %s", err)
		return 0
	}
	logrus.Infof("Current token payload: %s", string(payload))

	var payloadMap map[string]interface{}
	err = json.Unmarshal([]byte(payload), &payloadMap)
	if err != nil {
		logrus.Errorf("Error unmarshalling token payload: %s", err)
		return 0
	}
	audience := payloadMap["aud"].(string)

	identity := common.Identity{
		ClientId: payloadMap["appid"].(string),
	}

	// retrieve token using the existing token audience
	logrus.Infof("Retrieving new token for audience %s and identity %s", audience, identity)
	refreshToken, err := common.GetToken(audience, identity)

	if err != nil {
		logrus.Errorf("Error retrieving token: %s", err)
		return 0
	}
	logrus.Infof("Retrieved new token: %s", refreshToken.AccessToken)

	// Duration expects nanosecond count
	ExpiresInSeconds, err := strconv.ParseInt(refreshToken.ExpiresIn, 10, 64)
	if err != nil {
		logrus.Errorf("Error parsing token expiration to seconds: %s", err)
		return 0
	}
	credential.SetToken(refreshToken.AccessToken)
	return time.Duration(1000 * 1000 * 1000 * ExpiresInSeconds)
}

// For more information about the library used to access Azure:
//
//     https://pkg.go.dev/github.com/Azure/azure-storage-blob-go/azblob

func AzureSetup(urlString string, urlPrivate bool, identity common.Identity) error {
	// Create a ContainerURL object that wraps a blob's URL and a default
	// request pipeline.
	//
	// The pipeline indicates how the outgoing HTTP request and incoming HTTP
	// response is processed. It specifies things like retry policies, logging,
	// deserialization of HTTP response payloads, and more:
	//
	// https://pkg.go.dev/github.com/Azure/azure-storage-blob-go/azblob#hdr-URL_Types
	logrus.Infof("Connecting to Azure...")
	u, err := url.Parse(urlString)
	if err != nil {
		return errors.Wrapf(err, "Can't parse URL string %s", urlString)
	}

	if urlPrivate {
		// we use token credentials to access private azure blob storage the blob's
		// url Host denotes the scope/audience for which we need to get a token
		logrus.Info("Using token credentials to access private azure blob storage...")

		var token common.TokenResponse
		count := 0
		logrus.Infof("Getting token for https://%s", u.Host)
		for {
			token, err = common.GetToken("https://"+u.Host, identity)

			if err != nil {
				logrus.Info("Can't obtain a token required for accessing private blobs. Will retry in case the ACI identity sidecar is not running yet...")
				time.Sleep(3 * time.Second)
				count++
				if count == 20 {
					return errors.Wrapf(err, "Timeout of 60 seconds expired. Could not obtain token")
				}
			} else {
				logrus.Infof("Token obtained: %s \nContinuing...", token.AccessToken)
				break
			}
		}

		tokenCredential := azblob.NewTokenCredential(token.AccessToken, tokenRefresher)
		logrus.Infof("Token credential created: %s", tokenCredential.Token())
		fm.blobURL = azblob.NewPageBlobURL(*u, azblob.NewPipeline(tokenCredential, azblob.PipelineOptions{}))
		logrus.Infof("Blob URL created: %s", fm.blobURL)
	} else {
		// we can use anonymous credentials to access public azure blob storage
		logrus.Info("Using anonymous credentials to access public azure blob storage...")

		anonCredential := azblob.NewAnonymousCredential()
		logrus.Infof("Anonymous credential created: %s", anonCredential)
		fm.blobURL = azblob.NewPageBlobURL(*u, azblob.NewPipeline(anonCredential, azblob.PipelineOptions{}))
		logrus.Infof("Blob URL created: %s", fm.blobURL)
	}

	// Use a never-expiring context
	fm.ctx = context.Background()

	logrus.Info("Getting size of file...")
	// Get file size
	getMetadata, err := fm.blobURL.GetProperties(fm.ctx, azblob.BlobAccessConditions{},
		azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return errors.Wrapf(err, "Can't get blob file size")
	}
	fm.contentLength = getMetadata.ContentLength()
	logrus.Infof("Blob Size: %d bytes", fm.contentLength)

	// Setup data downloader and uploader
	fm.downloadBlock = AzureDownloadBlock
	fm.uploadBlock = AzureUploadBlock

	return nil
}

func AzureUploadBlock(blockIndex int64, b []byte) (err error) {
	logrus.Info("Uploading block...")
	bytesInBlock := GetBlockSize()
	var offset int64 = blockIndex * bytesInBlock
	logrus.Infof("Block offset %d = block index %d * bytes in blck %d", offset, blockIndex, bytesInBlock)

	r := bytes.NewReader(b)
	_, err = fm.blobURL.UploadPages(fm.ctx, offset, r, azblob.PageBlobAccessConditions{},
		nil, azblob.NewClientProvidedKeyOptions(nil, nil, nil))
	if err != nil {
		return errors.Wrapf(err, "Can't upload block")
	}

	return nil
}

func AzureDownloadBlock(blockIndex int64) (err error, b []byte) {
	logrus.Info("Downloading block...")
	bytesInBlock := GetBlockSize()
	var offset int64 = blockIndex * bytesInBlock
	logrus.Infof("Block offset %d = block index %d * bytes in blck %d", offset, blockIndex, bytesInBlock)
	var count int64 = bytesInBlock

	get, err := fm.blobURL.Download(fm.ctx, offset, count, azblob.BlobAccessConditions{},
		false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		var empty []byte
		return errors.Wrapf(err, "Can't download block"), empty
	}

	blobData := &bytes.Buffer{}
	reader := get.Body(azblob.RetryReaderOptions{})
	_, err = blobData.ReadFrom(reader)
	// The client must close the response body when finished with it
	reader.Close()

	if err != nil {
		var empty []byte
		return errors.Wrapf(err, "ReadFrom() failed for block"), empty
	}

	return nil, blobData.Bytes()
}
