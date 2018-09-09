package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (proxy *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// parts[0] = ignored, should be empty
	// parts[1] = jump host
	// parts[2] = destination address
	parts := strings.SplitN(r.RequestURI, "/", 3)
	if len(parts) != 3 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// build a new request
	req, err := http.NewRequest(r.Method, "http://"+parts[2], nil)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "unable to build request:", err)
		return
	}

	// set context and body
	req = req.WithContext(&proxyContext{context.Background(), proxy, parts[1]})
	req.Body = r.Body

	// do the request
	res, err := proxy.httpClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error())
		return
	}

	// copy response header and body
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}