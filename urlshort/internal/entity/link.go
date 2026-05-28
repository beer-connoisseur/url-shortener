package entity

import "errors"

type Link struct {
	ID           string
	OriginalLink string
	ShortLink    string
}

var ErrShortLinkNotFound = errors.New("short link not found")
var ErrOriginalLinkNotFound = errors.New("original link not found")
var ErrShortLinkAlreadyExists = errors.New("short link already exists")
var ErrOriginalLinkAlreadyExists = errors.New("original link already exists")
