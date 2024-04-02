package meme_maker

import (
	"fmt"
	meme_service "maas/meme-service"
	"maas/models"
)

type MemeMaker struct{}

func (m *MemeMaker) NewMeme() *models.Meme {
	return &models.Meme{TopText: "Up Top", BottomText: "Bottom Text", ImageLocation: "Nowhere and everywhere"}
}

func (m *MemeMaker) BuildMeme(query *meme_service.QueryParams) (*models.Meme, error) {
	meme := m.NewMeme()
	if query.Query != "" {
		meme = meme.WithTopText(query.Query)
	}
	if query.Lat != 0 && query.Lon != 0 {
		meme = meme.WithImageLocation(fmt.Sprintf("%.6f x %.6f", query.Lat, query.Lon))
	}
	return meme, nil
}
