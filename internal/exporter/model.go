package exporter

type Chunk struct {
	Text      string `json:"text"`
	IsThought bool   `json:"isThought"`
}

type ChunkedPrompt struct {
	Chunks []Chunk `json:"chunks"`
}

type Root struct {
	ChunkedPrompt ChunkedPrompt `json:"chunkedPrompt"`
}
