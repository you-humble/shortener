package model

type URLStore struct {
	UUID        int    `json:"uuid"`
	UserID      string `json:"user_id"`
	Short       string `json:"short_url"`
	Original    string `json:"original_url"`
	DeletedFlag bool   `json:"-"`
}

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	Original      string `json:"original_url"`
}

type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	Short         string `json:"short_url"`
}

type DeleteURLsRequest struct {
	UserID string
	URLs   []string
}
