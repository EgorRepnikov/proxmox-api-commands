package http_client

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

type RequestHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func InitHttpClient() *HttpClient {
	return &HttpClient{
		client: &fasthttp.Client{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

type HttpClient struct {
	client *fasthttp.Client
}

func (c *HttpClient) GET(
	url string,
	headers []*RequestHeader,
	timeout time.Duration,
) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	return c.sendReq(req, headers, timeout)
}

func (c *HttpClient) POST(
	url string,
	data []byte,
	headers []*RequestHeader,
	timeout time.Duration,
) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.SetBody(data)
	return c.sendReq(req, headers, timeout)
}

func (c *HttpClient) POSTGzip(
	url string,
	data []byte,
	headers []*RequestHeader,
	timeout time.Duration,
) (int, []byte, error) {
	var buf bytes.Buffer
	g := gzip.NewWriter(&buf)
	if _, err := g.Write(data); err != nil {
		return 0, nil, errors.Wrapf(err, "POSTGzip gzip write body, url='%s'", url)
	}
	if err := g.Close(); err != nil {
		return 0, nil, errors.Wrapf(err, "POSTGzip close gzip writer, url='%s'", url)
	}
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(url)
	req.Header.SetMethod("POST")
	req.Header.Set("Content-Encoding", "gzip")
	_, err := req.BodyWriter().Write(buf.Bytes())
	if err != nil {
		return 0, nil, errors.Wrapf(err, "POSTGzip write body in request, url='%s'", url)
	}
	status, data, err := c.sendReq(req, headers, timeout)
	if err != nil {
		return 0, nil, err
	}
	return status, data, nil
}

func (c *HttpClient) sendReq(
	req *fasthttp.Request,
	headers []*RequestHeader,
	timeout time.Duration,
) (int, []byte, error) {
	for _, v := range headers {
		req.Header.Set(v.Key, v.Value)
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	err := c.client.DoTimeout(req, resp, timeout)
	if err != nil {
		return 0, nil, errors.Wrapf(err, "sendReq fasthttp request, url='%s'", string(req.RequestURI()))
	}
	return resp.StatusCode(), readGzipBody(resp), nil
}

func readGzipBody(resp *fasthttp.Response) []byte {
	var body []byte
	contentEncoding := resp.Header.Peek("Content-Encoding")
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		var err error
		body, err = resp.BodyGunzip()
		if err == nil {
			log.Debug().Err(err).Msg("readGzipBody")
		}
	} else {
		body = resp.Body()
	}
	return body
}
