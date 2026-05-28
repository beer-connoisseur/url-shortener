package controller

import "pkg/generated/urlshort/api/urlshort/v1"

type API struct {
	urlshort.UrlshortServer
}

func New(urlshortServer urlshort.UrlshortServer) *API {
	return &API{
		UrlshortServer: urlshortServer,
	}
}
