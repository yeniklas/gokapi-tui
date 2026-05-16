package model

type GokapiFile struct {
	Id                  string `json:"Id"`
	Name                string `json:"Name"`
	Size                string `json:"Size"`
	SizeBytes           int64  `json:"SizeBytes"`
	ContentType         string `json:"ContentType"`
	UrlDownload         string `json:"UrlDownload"`
	UrlHotlink          string `json:"UrlHotlink"`
	ExpireAt            int64  `json:"ExpireAt"`
	ExpireAtString      string `json:"ExpireAtString"`
	UploadDate          int64  `json:"UploadDate"`
	DownloadsRemaining  int64  `json:"DownloadsRemaining"`
	DownloadCount       int64  `json:"DownloadCount"`
	UnlimitedDownloads  bool   `json:"UnlimitedDownloads"`
	UnlimitedTime       bool   `json:"UnlimitedTime"`
	IsPasswordProtected bool   `json:"IsPasswordProtected"`
}

type UploadParams struct {
	AllowedDownloads int
	ExpiryDays       int
	Password         string
}

type UploadResponse struct {
	Result          string     `json:"Result"`
	FileInfo        GokapiFile `json:"FileInfo"`
	IncludeFilename bool       `json:"IncludeFilename"`
}
