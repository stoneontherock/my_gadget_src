package cc

import (
	"cc/core/config"
	"net/http"
)

type ICC interface {
	HandleHttpRequest(resp http.ResponseWriter, req *http.Request) error
}

func NewCC(config *config.LocationLevel) ICC {
	return nil
}
