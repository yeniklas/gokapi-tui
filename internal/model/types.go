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
	FileRequestId       string `json:"FileRequestId"`
}

// FileRequest JSON field names are lowercase (Gokapi API convention for this endpoint)
type FileRequest struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Notes         string   `json:"notes"`
	Expiry        int64    `json:"expiry"`
	CreationDate  int64    `json:"creationdate"`
	MaxFiles      int      `json:"maxfiles"`
	MaxSize       int      `json:"maxsize"` // MB; 0 = unlimited
	ApiKey        string   `json:"apikey"`
	UploadedFiles int      `json:"uploadedfiles"`
	LastUpload    int64    `json:"lastupload"`
	TotalFileSize int64    `json:"totalfilesize"`
	FileIdList    []string `json:"fileidlist"`
}

type CreateFileRequestParams struct {
	Name     string
	Notes    string
	ExpiryAt int64 // unix timestamp; 0 = no expiry
	MaxFiles int   // 0 = unlimited
	MaxSize  int   // MB; 0 = unlimited
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
