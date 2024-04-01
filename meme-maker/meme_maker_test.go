package meme_maker

import (
	"testing"

	query_params "maas/query-params"

	"github.com/stretchr/testify/assert"
)

func TestNewMeme_ReturnsDefaultValues(t *testing.T) {
	meme := NewMeme()
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestWithTopText_OnlyEditsTopText(t *testing.T) {
	meme := NewMeme().withTopText("Different Top Text")
	assert.Equal(t, meme.TopText, "Different Top Text")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")

}

func TestWithBottomText_OnlyEditsBottomText(t *testing.T) {
	meme := NewMeme().withBottomText("Different Bottom Text")
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Different Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")

}

func TestWithImageLocation_OnlyEditsTopText(t *testing.T) {
	meme := NewMeme().withImageLocation("Different Location")
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Different Location")
}

func TestWithMethods_ChainWell(t *testing.T) {
	meme := NewMeme().withTopText("Test Top Text").withBottomText("Test Bottom Text").withImageLocation("Test Image Location")
	assert.Equal(t, meme.TopText, "Test Top Text")
	assert.Equal(t, meme.BottomText, "Test Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Test Image Location")
}

func TestBuildMeme_WhenQueryParamsAreNil_ReturnsDefaultMeme(t *testing.T) {
	meme := BuildMeme(&query_params.QueryParams{})
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenQueryParamsHasQueryDefined_UpdatesMemeTopText(t *testing.T) {
	meme := BuildMeme(&query_params.QueryParams{Query: "My Query"})
	assert.Equal(t, meme.TopText, "My Query")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenQueryParamsHasLatAndLonDefined_UpdatesMemeImageLocation(t *testing.T) {
	meme := BuildMeme(&query_params.QueryParams{Lat: 1.1111119, Lon: 2.0})
	assert.Equal(t, meme.TopText, "Up Top")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "1.111112 x 2.000000")
}

func TestBuildMeme_WhenQueryParamsHasLatOrLonAreDefinedButNotBoth_ReturnsDefaultMeme(t *testing.T) {
	meme1 := BuildMeme(&query_params.QueryParams{Lat: 1.1111119})
	assert.Equal(t, meme1.TopText, "Up Top")
	assert.Equal(t, meme1.BottomText, "Bottom Text")
	assert.Equal(t, meme1.ImageLocation, "Nowhere and everywhere")

	meme2 := BuildMeme(&query_params.QueryParams{Lon: 1.1111119})
	assert.Equal(t, meme2.TopText, "Up Top")
	assert.Equal(t, meme2.BottomText, "Bottom Text")
	assert.Equal(t, meme2.ImageLocation, "Nowhere and everywhere")
}

func TestBuildMeme_WhenAllQueryParamsAreDefined_UpdatesMeme(t *testing.T) {
	meme := BuildMeme(&query_params.QueryParams{Query: "My Query", Lat: 1.1111119, Lon: 2.0})
	assert.Equal(t, meme.TopText, "My Query")
	assert.Equal(t, meme.BottomText, "Bottom Text")
	assert.Equal(t, meme.ImageLocation, "1.111112 x 2.000000")
}

func TestMakeMap_ReturnsMapOfMeme(t *testing.T) {
	expected := map[string]string{
		"top_text":       "Up Top",
		"bottom_text":    "Bottom Text",
		"image_location": "Nowhere and everywhere",
	}
	actual := NewMeme().MakeMap()
	assert.Equal(t, expected, actual)
}
