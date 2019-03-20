package webmasq

import (
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	BinFileDebug = false //中间resp 二进制debug
	SelfHost     string
	selfScheme   = "https"
	catchReg     = regexp.MustCompile(`["'(]https?://[:alnum:][^"'(]+`) //激进版,all http/https url
	repl30X      = regexp.MustCompile(`(https?://)[^/]+(.*)`)

	transport http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

func Serve(self, certF, keyF, selfHost string) error {
	SelfHost = selfHost

	http.HandleFunc("/", masquerade)
	http.HandleFunc("/zzz", newSite)
	http.HandleFunc("/site", newSiteHandler)
	//http.HandleFunc("/login", login)

	http.HandleFunc("/favicon.ico", func(http.ResponseWriter, *http.Request) { return })
	return http.ListenAndServeTLS(self, certF, keyF, nil)
}

func newSite(wr http.ResponseWriter, req *http.Request) {
	wr.Write([]byte(newSiteHtml))
}

func newSiteHandler(wr http.ResponseWriter, req *http.Request) {
	logrus.Debugf("newSiteHandler(), newSite(), url=%s", req.URL.String())
	formHP := req.FormValue("siteAddr")
	if formHP == "" {
		logrus.Errorf("空查询字段")
		wr.Write([]byte(jump2NewSiteJS))
		return
	}

	logrus.Debugf("newSiteHandler(), request from form,%s", req.URL.String())
	zScheme := req.FormValue("webScheme")
	cs := zScheme + strings.Split(formHP, "/")[0]
	setCurSiteCookie(wr, cs)

	u, err := url.Parse(zScheme + formHP)
	if err != nil {
		logrus.Errorf("newSiteHandler(),url.Parse(%s) failed, %s", cs, err.Error())
		http.Error(wr, err.Error(), 500)
		return
	}

	proxy(u, wr, req)
}

func masquerade(wr http.ResponseWriter, req *http.Request) {
	logrus.Debugf("masquerade(), origin URL:%s", req.URL.String())

	var dstURL url.URL

	dz := req.FormValue("dstZite")
	if dz != "" {
		u, err := url.Parse(dz)
		if err != nil {
			logrus.Errorf("masquerade(), parse dz->url failed")
			http.Error(wr, err.Error(), 500)
			return
		}
		dstURL = *u
		logrus.Debugf("masquerade() by dstZite, scheme:%s, host:%s", dstURL.Scheme, dstURL.Host)
	} else {
		cz, err := req.Cookie("curZite")
		if err != nil {
			logrus.Errorf("masquerade(), No cookie,%s", req.URL.String())
			http.Error(wr, "no ck", 500)
			return
		}

		u, err := url.Parse(cz.Value)
		if err != nil {
			logrus.Errorf("masquerade(), cookie -> url failed,%s", cz.Value)
			http.Error(wr, "Parse-ck", 500)
			return
		}

		dstURL = *req.URL
		dstURL.Scheme = u.Scheme
		dstURL.Host = u.Host

	}

	logrus.Debugf("masquerade(), dest url=%s", dstURL.String())

	proxy(&dstURL, wr, req)
}

func proxy(dstURL *url.URL, wr http.ResponseWriter, req *http.Request) {
	logrus.Debugf("proxy(): ver=%s dstURL=%s", req.Proto, dstURL.String())

	var rd io.Reader
	if req.Method == "POST" {
		rd = req.Body
	}
	outreq, err := http.NewRequest(req.Method, dstURL.String(), rd)
	if err != nil {
		logrus.Errorf("proxy(),http.NewRequest(),%v", err)
		http.Error(wr, err.Error(), 500)
		return
	}

	outreq.Header = cloneHeader(req.Header)
	outreq.Close = false
	removeConnectionHeaders(outreq.Header)
	for _, h := range hopHeaders {
		hv := outreq.Header.Get(h)
		if hv == "" {
			continue
		}
		if h == "Te" && hv == "trailers" {
			continue
		}
		outreq.Header.Del(h)
	}

	*outreq.URL = *dstURL

	//去掉br
	//ae := outreq.Header["Accept-Encoding"]
	//for i, _ := range ae {
	//	if ae[i] == "br" {
	//		ae = append(ae[:i], ae[i+1:]...)
	//		outreq.Header["Accept-Encoding"] = ae
	//		logrus.Warningf("proxy(), br found !! after del:%v", outreq.Header.Get("Accept-Encoding"))
	//		break
	//	}
	//}
	outreq.Header.Set("Accept-Encoding", "gzip")
	outreq.Header.Set("Referer", "https://www.baidu.com/"+dstURL.Host+"/"+dstURL.RequestURI())

	logrus.Debugf("proxy(), outreq=%v", outreq)
	res, err := transport.RoundTrip(outreq)
	if err != nil {
		logrus.Errorf("proxy(),transport.RoundTrip(), %s", err.Error())
		http.Error(wr, err.Error(), 500)
		return
	}

	removeConnectionHeaders(res.Header)
	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	//logrus.Debugf("proxy(), res.Header: %v", res.Header)
	res.Header.Del("Content-Length")
	copyHeader(wr.Header(), res.Header)

	announcedTrailers := len(res.Trailer)
	if announcedTrailers > 0 {
		trailerKeys := make([]string, 0, len(res.Trailer))
		for k := range res.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		wr.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	modify30XHeader(wr, outreq, res, dstURL)
	wr.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		if fl, ok := wr.(http.Flusher); ok {
			fl.Flush()
		}
	}

	replaceURL(wr, outreq, res)

	if len(res.Trailer) == announcedTrailers {
		copyHeader(wr.Header(), res.Trailer)
		return
	}

	for k, vv := range res.Trailer {
		k = http.TrailerPrefix + k
		for _, v := range vv {
			wr.Header().Add(k, v)
		}
	}

}

func modify30XHeader(wr http.ResponseWriter, req *http.Request, res *http.Response, originURL *url.URL) {
	if res.StatusCode != 301 && res.StatusCode != 302 && res.StatusCode != 303 && res.StatusCode != 307 {
		return
	}

	nl := res.Header.Get("Location")
	u, err := url.Parse(nl)
	if err != nil {
		logrus.Errorf("modify30XHeader(), parse location header to url failed, url.Parse(%s),%v", nl, err)
		return
	}

	var v = make(url.Values, 1)
	v.Add("destZite", u.Scheme+"://"+nl)
	u.RawQuery = v.Encode()
	u.Scheme = selfScheme
	u.Host = SelfHost
	res.Header.Set("Location", u.String())

	if u.Scheme != originURL.Scheme || u.Host != originURL.Host {
		logrus.Debugf("modify30XHeader(), scheme/host changed, refresh cookie")
		setCurSiteCookie(wr, u.Scheme+"://"+u.Host)
	}

	logrus.Debugf("30X modify header, old:%s, new：%s", originURL, nl)
}

func replaceURL(wr http.ResponseWriter, req *http.Request, res *http.Response) {
	defer res.Body.Close()
	ce := res.Header.Get("Content-Encoding")
	logrus.Debugf("replaceURL(), CE=%s", ce)

	var buf []byte
	var err error

	switch ce {
	case "":
		buf, err = ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Errorf("replaceURL(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
			http.Error(wr, err.Error(), 500)
			return
		}
	case "gzip":
		gr, err := gzip.NewReader(res.Body)
		if err != nil {
			logrus.Errorf("replaceURL(), gzip.NewReader(), %v", err)
			http.Error(wr, err.Error(), 500)
			return
		}
		buf, err = ioutil.ReadAll(gr)
		if err != nil {
			logrus.Errorf("replaceURL(), ioutil.ReadAll(), %v", err)
			http.Error(wr, err.Error(), 500)
			return
		}
	default:
		logrus.Warningf("replaceURL(), CE: %s not support, do not replace url", ce)
		buf, err = ioutil.ReadAll(res.Body)
		if err != nil {
			logrus.Errorf("replaceURL(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
			http.Error(wr, err.Error(), 500)
			return
		}
	}

	if len(buf) == 0 {
		logrus.Errorf("replaceURL(), zero-len buf")
		return
	}

	writeToBinFile(buf, req.Host+"_"+url.PathEscape(req.RequestURI))

	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
	if ce == "gzip" {
		gw := gzip.NewWriter(wr)
		gw.Write([]byte(str))
		gw.Flush()
	} else {
		wr.Write([]byte(str))
	}
}

func genReplFunc() func(string) string {

	return func(s string) string {
		v := make(url.Values, 1)
		v.Set("dstZite", s[1:])

		var u url.URL
		u.Scheme = selfScheme
		u.Host = SelfHost
		u.RawPath = "/"
		u.RawQuery = v.Encode()

		str := string(s[0]) + u.String() //fixme
		logrus.Debugf("***before repl:%s after repl:%s", s, str)
		return str
	}
}

func login(wr http.ResponseWriter, req *http.Request) {
	uname := req.FormValue("uname")
	upwd := req.FormValue("upwd")
	if uname == "" || upwd == "" {
		sendAuth(wr)
		return
	}

	h := Md5sum(uname + "/" + upwd)
	if !auth(h) {
		sendAuth(wr)
		return
	}

	ck := http.Cookie{
		Name:     "zite",
		Value:    h,
		Path:     "/",
		HttpOnly: true,
	}

	http.SetCookie(wr, &ck)
	wr.Write([]byte(newSiteHtml))
	return
}

func setCurSiteCookie(wr http.ResponseWriter, s string) {
	ck := http.Cookie{
		Name:     "curZite",
		Value:    s,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(wr, &ck)
}

func sendAuth(wr http.ResponseWriter) {
	wr.Write([]byte(loginHtml))
}

func auth(hashStr string) bool {
	for i, _ := range userAuth {
		if hashStr == userAuth[i] {

			return true
		}
	}
	return false
}

func Md5sum(str string) string {
	w := md5.New()
	io.WriteString(w, str+salt)
	return hex.EncodeToString(w.Sum(nil))
}

func writeToBinFile(buf []byte, fname string) {
	if !BinFileDebug {
		return
	}
	logrus.Debugf("writeToBinFile(), fname=%s", fname)
	err := ioutil.WriteFile(fname, buf, 0444)
	if err != nil {
		logrus.Errorf("writeToBinFile(), %v", err)
	}
}

////安全备份
//func roundTrip1(dstURL *url.URL, wr http.ResponseWriter, req *http.Request) {
//	client := http.Client{}
//	outreq, err := http.NewRequest("GET", dstURL.String(), nil)
//	if err != nil {
//		logrus.Errorf("proxy(),http.NewRequest(),%v", err)
//		return
//	}
//
//	outreq.Header = cloneHeader(req.Header)
//	outreq.Close = false
//	removeConnectionHeaders(outreq.Header)
//
//	// Remove hop-by-hop headers to the backend. Especially
//	// important is "Connection" because we want a persistent
//	// connection, regardless of what the client sent to us.
//	for _, h := range hopHeaders {
//		hv := outreq.Header.Get(h)
//		if hv == "" {
//			continue
//		}
//		if h == "Te" && hv == "trailers" {
//			// Issue 21096: tell backend applications that
//			// care about trailer support that we support
//			// trailers. (We do, but we don't go out of
//			// our way to advertise that unless the
//			// incoming client request thought it was
//			// worth mentioning)
//			continue
//		}
//		outreq.Header.Del(h)
//	}
//
//	resp, err := client.Do(outreq)
//	if err != nil {
//		logrus.Errorf("proxy(), client.Do(), %s", err.Error())
//		return
//	}
//
//	removeConnectionHeaders(resp.Header)
//
//	for _, h := range hopHeaders {
//		resp.Header.Del(h)
//	}
//
//	copyHeader(wr.Header(), resp.Header)
//
//	// The "Trailer" header isn't included in the Transport's response,
//	// at least for *http.Transport. Build it up from Trailer.
//	announcedTrailers := len(resp.Trailer)
//	if announcedTrailers > 0 {
//		trailerKeys := make([]string, 0, len(resp.Trailer))
//		for k := range resp.Trailer {
//			trailerKeys = append(trailerKeys, k)
//		}
//		wr.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
//	}
//
//	wr.WriteHeader(resp.StatusCode)
//	if len(resp.Trailer) > 0 {
//		// Force chunking if we saw a response trailer.
//		// This prevents net/http from calculating the length for short
//		// bodies and adding a Content-Length.
//		if fl, ok := wr.(http.Flusher); ok {
//			fl.Flush()
//		}
//	}
//
//	buf, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		logrus.Errorf("roundTrip(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
//		http.Error(wr, err.Error(), 500)
//		resp.Body.Close()
//		return
//	}
//
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
//	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
//	wr.Write([]byte(str))
//	resp.Body.Close()
//}
//
//
////安全备份
//func roundTrip_0(dstURL *url.URL, wr http.ResponseWriter, req *http.Request) {
//	resp, err := http.Get(dstURL.String())
//	if err != nil {
//		logrus.Errorf("roundTrip(),http.Get(%s) failed, %s", dstURL.String(), err.Error())
//		http.Error(wr, err.Error(), resp.StatusCode)
//		return
//	}
//	defer resp.Body.Close()
//
//	buf, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		logrus.Errorf("roundTrip(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
//		http.Error(wr, err.Error(), 500)
//		return
//	}
//
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
//	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
//	wr.Write([]byte(str))
//}

//安全备份
//func roundTrip(dstURL *url.URL, wr http.ResponseWriter) {
//	resp, err := http.Get(dstURL.String())
//	if err != nil {
//		logrus.Errorf("roundTrip(),http.Get(%s) failed, %s", dstURL.String(), err.Error())
//		http.Error(wr, err.Error(), resp.StatusCode)
//		return
//	}
//	defer resp.Body.Close()
//
//	buf, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		logrus.Errorf("roundTrip(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
//		http.Error(wr, err.Error(), 500)
//		return
//	}
//
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
//	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
//	wr.Write([]byte(str))
//}
