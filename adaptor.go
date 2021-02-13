package http2

import (
	"bytes"
	"github.com/valyala/fasthttp"
	"strconv"
)

func fasthttpRequestHeaders(hp *HPACK, req *fasthttp.Request) {
	hp.Range(func(hf *HeaderField) {
		k, v := hf.KeyBytes(), hf.ValueBytes()
		if !hf.IsPseudo() { // TODO: && !bytes.Equal(k, strUserAgent) {
			req.Header.AddBytesKV(k, v)
			return
		}

		uri := req.URI()
		switch k[1] {
		case 'm': // method
			req.Header.SetMethodBytes(v)
		case 'p': // path
			uri.SetPathBytes(v)
		case 's': // scheme
			uri.SetSchemeBytes(v)
		case 'a': // authority
			uri.SetHostBytes(v)
			// TODO: See below?? case 'u': // user-agent
			// 	req.Header.SetUserAgentBytes(v)
		}
	})
}

func fasthttpResponseHeaders(hp *HPACK, res *fasthttp.Response) {
	hp.AddBytesK(strStatus,
		strconv.FormatInt(
			int64(res.Header.StatusCode()), 10,
		),
	)

	hp.AddBytesK(strContentLength,
		strconv.FormatInt(int64(len(res.Body())), 10),
	)

	res.Header.VisitAll(func(k, v []byte) {
		hp.AddBytes(bytes.ToLower(k), v)
	})
}
