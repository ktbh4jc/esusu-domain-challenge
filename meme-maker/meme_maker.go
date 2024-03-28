package meme_maker

import (
	"fmt"
	query_params "maas/query-params"
)

type Meme struct {
	TopText       string `json:"top_text"`
	BottomText    string `json:"bottom_text"`
	ImageLocation string `json:"image_location"`
}

func NewMeme() *Meme {
	return &Meme{TopText: "Up Top", BottomText: "Bottom Text", ImageLocation: "Nowhere and everywhere"}
}

func (m *Meme) withTopText(topText string) *Meme {
	m.TopText = topText
	return m
}

func (m *Meme) withBottomText(bottomText string) *Meme {
	m.BottomText = bottomText
	return m
}

func (m *Meme) withImageLocation(imageLocation string) *Meme {
	m.ImageLocation = imageLocation
	return m
}

func BuildMeme(query *query_params.QueryParams) *Meme {
	meme := NewMeme()
	if query.Query != "" {
		meme = meme.withTopText(query.Query)
	}
	if query.Lat != 0 && query.Lon != 0 {
		meme = meme.withImageLocation(fmt.Sprintf("%.6f x %.6f", query.Lat, query.Lon))
	}
	return meme
}

func (m *Meme) MakeMap() map[string]string {
	return map[string]string{
		"top_text":       m.TopText,
		"bottom_text":    m.BottomText,
		"image_location": m.ImageLocation,
	}
}
