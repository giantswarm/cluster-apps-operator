package proxy

type Proxy struct {
	NoProxy    string `json:"noProxy"`
	HttpProxy  string `json:"http"`
	HttpsProxy string `json:"https"`
}
