package mlog

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPattern(t *testing.T) {
	ms := pattern.FindStringSubmatch("/logs/a/level")
	assert.Len(t, ms, 2)
	assert.Equal(t, "a", ms[1])
}

func TestParseLogName(t *testing.T) {
	td := []struct {
		path    string
		logName string
		err     error
	}{
		{
			path:    "/logs/a/level",
			logName: "a",
			err:     nil,
		},
		{
			path:    "/logs/a/name",
			logName: "",
			err:     errors.New("invalid path"),
		},
	}

	for _, d := range td {
		t.Run(d.path, func(t *testing.T) {
			logName, err := parseLogName(d.path)
			assert.Equal(t, d.logName, logName)
			assert.Equal(t, d.err, err)
		})
	}
}

func TestServeHTTP(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(ServeHTTP))
	defer ts.Close()

	rsp, err := http.Get(ts.URL + "/logs/a/level")
	assert.NoError(t, err)

	assert.Equal(t, rsp.StatusCode, http.StatusOK)
	body, err := ioutil.ReadAll(rsp.Body)
	defer Close(rsp.Body)
	assert.NoError(t, err)
	t.Log(string(body))

	req, err := http.NewRequest(http.MethodPut, ts.URL+"/logs/a/level", strings.NewReader(`{"level":"info"}`))
	assert.NoError(t, err)
	rsp, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, rsp.StatusCode, http.StatusOK)
	body, err = ioutil.ReadAll(rsp.Body)
	defer Close(rsp.Body)
	assert.NoError(t, err)
	t.Log(string(body))
}
