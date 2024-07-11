package reddit

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setup(t testing.TB) (*Client, *http.ServeMux) {
	mux := http.NewServeMux()

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	mux.HandleFunc("/api/v1/access_token", func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"access_token": "token1",
			"token_type": "bearer",
			"expires_in": 3600,
			"scope": "*"
		}`
		w.Header().Add(headerContentType, mediaTypeJSON)
		if _, err := fmt.Fprint(w, response); err != nil {
			t.Fatal(err)
		}
	})

	client, _ := NewClient(
		Credentials{"", "", "", ""},
		FromEnv,
		WithBaseURL(defaultBaseURL),
		WithTokenURL(defaultTokenURL),
	)

	return client, mux
}

const testDataPath = "../testdata"

func readFileContents(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Read the file into byte array
	bytes := make([]byte, 0, stat.Size())
	_, err = bufio.NewReader(file).Read(bytes)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return bytes, nil
}

// TESTING METHODS
