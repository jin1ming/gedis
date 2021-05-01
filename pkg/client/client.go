package client

import (
	"github.com/pkg/errors"
	"net/http"
)

var ErrRedirect = errors.New("unexpected redirect in response")

type Client struct {
	scheme		string
	host		string
	protocol	string
	addr		string
	client		*http.Client
	version		string
}