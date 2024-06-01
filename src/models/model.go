package models

type Bucket struct {
	DocCount int    `json:"doc_count"`
	Key      string `json:"key"`
}

type Aggregations struct {
	DistinctValueCount struct {
		Buckets                 []Bucket `json:"buckets"`
		DocCountErrorUpperBound int      `json:"doc_count_error_upper_bound"`
		SumOtherDocCount        int      `json:"sum_other_doc_count"`
	} `json:"distinct_value_count"`
	MaxDistinctCounts struct {
		Keys  []string `json:"keys"`
		Value float32  `json:"value"`
	} `json:"max_distinct_counts"`
}

type SearchResult struct {
	Hits struct {
		Hits []Hit `json:"hits"`
	} `json:"hits"`
	Aggregations `json:"aggregations"`
}

type Hit struct {
	Source map[string]interface{} `json:"_source"`
}

type Document struct {
	Content string
}
