package downloader

import (
	"github.com/admpub/spider/app/downloader/request"
	"github.com/admpub/spider/app/spider"
	"github.com/admpub/spider/logs"
)

// The Downloader interface.
// You can implement the interface by implement function Download.
// Function Download need to return Page instance pointer that has request result downloaded from Request.
type Downloader interface {
	Download(*spider.Spider, *request.Request, logs.Logs) *spider.Context
}
