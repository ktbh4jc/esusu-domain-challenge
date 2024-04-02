package models

type Meme struct {
	TopText       string `json:"top_text"`
	BottomText    string `json:"bottom_text"`
	ImageLocation string `json:"image_location"`
}

func (m *Meme) MakeMap() map[string]string {
	return map[string]string{
		"top_text":       m.TopText,
		"bottom_text":    m.BottomText,
		"image_location": m.ImageLocation,
	}
}

func (m *Meme) WithTopText(topText string) *Meme {
	m.TopText = topText
	return m
}

func (m *Meme) WithBottomText(bottomText string) *Meme {
	m.BottomText = bottomText
	return m
}

func (m *Meme) WithImageLocation(imageLocation string) *Meme {
	m.ImageLocation = imageLocation
	return m
}
