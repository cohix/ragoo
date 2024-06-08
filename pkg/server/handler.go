package server

import (
	"net/http"

	"github.com/cohix/ragoo/pkg/config"
)

func handlerForRoute(route config.Route) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
