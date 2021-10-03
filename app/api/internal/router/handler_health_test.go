package router

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleHealthCheck(t *testing.T) {
	r := newRouter(nil, newMockSurferService(), newMockManagementService(), nil)

	server := httptest.NewServer(r)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.NoError(t, err)

	assert.Equal(t, "", string(body))
}
