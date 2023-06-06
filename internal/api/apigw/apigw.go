package apigw

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	"github.com/stax-labs/terraform-provider-stax/internal/api/auth"
	"github.com/stax-labs/terraform-provider-stax/internal/api/openapi/client"
)

func RequestSigner(region string, credsRetriever auth.CredentialsRetrieverFn) client.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		return signRequest(region, ctx, req, credsRetriever)
	}
}

func signRequest(region string, ctx context.Context, req *http.Request, credsReRetriever auth.CredentialsRetrieverFn) error {
	signer := v4.NewSigner()
	credentials, err := credsReRetriever(ctx)
	if err != nil {
		return err
	}

	checksum, err := calculateSum(ctx, req)
	if err != nil {
		return err
	}

	err = signer.SignHTTP(ctx, credentials, req, checksum, "execute-api", region, time.Now())
	if err != nil {
		fmt.Printf("failed to sign request: (%v)\n", err)
		return err
	}

	return nil
}

func calculateSum(_ context.Context, req *http.Request) (string, error) {
	payload, err := readAndReplaceBody(req)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(payload)

	return hex.EncodeToString(hash[:]), nil
}

func readAndReplaceBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return []byte(""), nil
	}

	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	req.Body = io.NopCloser(bytes.NewReader(payload))

	return payload, nil
}
