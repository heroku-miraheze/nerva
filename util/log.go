// Copyright © 2016 Alexander Gugel <alexander.gugel@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package util

import (
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

// LogWarn logs the passed in error.
func LogWarn(ctx *log.Entry, err error, reason string) {
	ctx.WithFields(log.Fields{"error": err}).Warn(reason)
}

// LogErr logs the passed in error.
func LogErr(ctx *log.Entry, err error, reason string) {
	ctx.WithFields(log.Fields{"error": err}).Error(reason)
}

// LogFatal logs the passed in error and exits.
func LogFatal(ctx *log.Entry, err error, reason string) {
	ctx.WithFields(log.Fields{"error": err}).Fatal(reason)
}

// RequestContext creates a new context used for logging from a plain HTTP
// request.
func RequestContext(req *http.Request) *log.Entry {
	return log.WithFields(log.Fields{
		"RemoteAddr": req.RemoteAddr,
		"RequestURI": req.RequestURI,
		"URL":        req.URL,
	})
}

// LogRequest logs an incoming HTTP request.
func LogRequest(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		RequestContext(req).Info("request")
		handle(w, req, ps)
	}
}
