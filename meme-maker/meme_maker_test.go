package meme_maker

import (
	"testing"

	meme_service "maas/meme-service"

	"github.com/stretchr/testify/assert"
)

func TestNewMeme_ReturnsDefaultValues(t *testing.T) {
	maker := &MemeMaker{}
	meme := maker.NewMeme()
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestWithTopText_OnlyEditsTopText(t *testing.T) {
	maker := &MemeMaker{}
	meme := maker.NewMeme().WithTopText("Different Top Text")
	assert.Equal(t, meme.TopText, "Different Top Text")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")

}

func TestWithBottomText_OnlyEditsBottomText(t *testing.T) {
	maker := &MemeMaker{}
	meme := maker.NewMeme().WithBottomText("Different Bottom Text")
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Different Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")

}

func TestWithImageLocation_OnlyEditsTopText(t *testing.T) {
	maker := &MemeMaker{}
	meme := maker.NewMeme().WithImageLocation("Different Location")
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Different Location")
}

func TestWithMethods_ChainWell(t *testing.T) {
	maker := &MemeMaker{}
	meme := maker.NewMeme().WithTopText("Test Top Text").WithBottomText("Test Bottom Text").WithImageLocation("Test Image Location")
	assert.Equal(t, meme.TopText, "Test Top Text")
	assert.Equal(t, meme.BottomText, "Test Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Test Image Location")
}

func TestBuildMeme_WhenQueryParamsAreNil_ReturnsDefaultMeme(t *testing.T) {
	maker := &MemeMaker{}
	meme, err := maker.BuildMeme(&meme_service.QueryParams{})
	assert.Nil(t, err)
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenQueryParamsHasQueryDefined_UpdatesMemeTopText(t *testing.T) {
	maker := &MemeMaker{}
	meme, err := maker.BuildMeme(&meme_service.QueryParams{Query: "My Query"})
	assert.Nil(t, err)
	assert.Equal(t, meme.TopText, "My Query")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenQueryParamsHasLatAndLonDefined_UpdatesMemeImageLocation(t *testing.T) {
	maker := &MemeMaker{}
	meme, err := maker.BuildMeme(&meme_service.QueryParams{Lat: 1.1111119, Lon: 2.0})
	assert.Nil(t, err)
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "1.111112 x 2.000000")
}

func TestBuildMeme_WhenQueryParamsHasLatOrLonAreDefinedButNotBoth_ReturnsDefaultMeme(t *testing.T) {
	maker := &MemeMaker{}
	meme1, err := maker.BuildMeme(&meme_service.QueryParams{Lat: 1.1111119})
	assert.Nil(t, err)
	assert.Equal(t, meme1.TopText, "Up Top")
	assert.Equal(t, meme1.BottomText, "Bottom Text")
	assert.Equal(t, meme1.ImageLocation, "Nowhere and everywhere")

	meme2, err := maker.BuildMeme(&meme_service.QueryParams{Lon: 1.1111119})
	assert.Nil(t, err)
	assert.Equal(t, meme2.TopText, "Up Top")
	assert.Equal(t, meme2.BottomText, "Bottom Text")
	assert.Equal(t, meme2.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenAllQueryParamsAreDefined_UpdatesMeme(t *testing.T) {
	maker := &MemeMaker{}
	meme, err := maker.BuildMeme(&meme_service.QueryParams{Query: "My Query", Lat: 1.1111119, Lon: 2.0})
	assert.Nil(t, err)
	assert.Equal(t, meme.TopText, "My Query")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "1.111112 x 2.000000")
}

func TestMakeMap_ReturnsMapOfMeme(t *testing.T) {
	maker := &MemeMaker{}
	expected := map[string]string{
		"top_text":       "Up Top",
		"bottom_text":    "Bottom Text",
		"image_location": "Nowhere and everywhere",
	}
	actual := maker.NewMeme().MakeMap()
	assert.Equal(t, expected, actual)
}
