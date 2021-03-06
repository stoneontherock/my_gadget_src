package webmasq

import (
	//"bytes"
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
	selfHost   string
	selfScheme                   = "https"
	catchReg                     = regexp.MustCompile(`(src|href|content) *= *"http[^"]+`)
	transport  http.RoundTripper = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		//DisableCompression:    true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
)

func Serve(self, certF, keyF string) error {
	selfHost = self

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
	logrus.Debugf("newSite(), url=%s", req.URL.String())
	formHP := req.FormValue("formHP")
	if formHP == "" {
		logrus.Errorf("query string is empty")
		wr.Write([]byte(jumpToForm))
		return
	}

	logrus.Debugf("从form表单发起的请求,%s", req.URL.String())
	zScheme := req.FormValue("formS")

	cs := zScheme + strings.Split(formHP, "/")[0]
	setCurSiteCookie(wr, cs)

	u, err := url.Parse(zScheme + formHP)
	if err != nil {
		logrus.Errorf("newSite(),url.Parse(%s) failed, %s", cs, err.Error())
		return
	}
	proxy(u, wr, req)

}

func masquerade(wr http.ResponseWriter, req *http.Request) {
	logrus.Debugf("masquerade(), %s", req.URL.String())

	req.Cookie("")

	var dstURL url.URL
	dstURL = *req.URL

	dz := req.FormValue("dstZite")
	if dz != "" {
		u, err := url.Parse(dz)
		if err != nil {
			logrus.Errorf("masquerade(),解析dz为url失败")
			return
		}
		dstURL = *u
		logrus.Debugf("使用dstZite,scheme:%s, host:%s", dstURL.Scheme, dstURL.Host)
	} else {
		cz, err := req.Cookie("curZite")
		if err != nil {
			logrus.Errorf("未带cookie,%s", req.URL.String())
			http.Error(wr, "No-CoOKIE", 502)
			return
		}

		u, err := url.Parse(cz.Value)
		if err != nil {
			logrus.Errorf("masquerade(), 甜饼解析成url失败,%s", cz.Value)
			http.Error(wr, "Parse-COokie", 502)
			return
		}
		dstURL.Scheme = u.Scheme
		dstURL.Host = u.Host
	}

	logrus.Debugf("masquerade(), dst url=%s", dstURL.String())

	proxy(&dstURL, wr, req)
}

func proxy(dstURL *url.URL, wr http.ResponseWriter, r *http.Request) {
	logrus.Debugf("proxy: %s", dstURL.String())

	var rd io.Reader
	if r.Method == "POST" {
		rd = r.Body
	}
	outreq, err := http.NewRequest(r.Method, dstURL.String(), rd)
	if err != nil {
		logrus.Errorf("roundTrip(),http.NewRequest(),%v", err)
		return
	}

	outreq.Header = cloneHeader(r.Header)
	outreq.Close = false

	removeConnectionHeaders(outreq.Header)

	// Remove hop-by-hop headers to the backend. Especially
	// important is "Connection" because we want a persistent
	// connection, regardless of what the client sent to us.
	for _, h := range hopHeaders {
		hv := outreq.Header.Get(h)
		if hv == "" {
			continue
		}
		if h == "Te" && hv == "trailers" {
			// Issue 21096: tell backend applications that
			// care about trailer support that we support
			// trailers. (We do, but we don't go out of
			// our way to advertise that unless the
			// incoming client request thought it was
			// worth mentioning)
			continue
		}
		outreq.Header.Del(h)
	}

	*outreq.URL = *dstURL
	res, err := transport.RoundTrip(outreq)
	if err != nil {
		logrus.Errorf("roundTrip(),transport.RoundTrip(), %s", err.Error())
		return
	}

	removeConnectionHeaders(res.Header)

	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	res.Header.Del("Content-Length")
	copyHeader(wr.Header(), res.Header)

	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(res.Trailer)
	if announcedTrailers > 0 {
		trailerKeys := make([]string, 0, len(res.Trailer))
		for k := range res.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		wr.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	wr.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := wr.(http.Flusher); ok {
			fl.Flush()
		}
	}

	replaceURL(wr, res.Body)

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

func genReplFunc() func(string) string {
	return func(s string) string {
		i := strings.Index(s, "=")
		k := strings.TrimRight(s[:i], " ")
		v := strings.TrimLeft(s[i+1:], " ")
		v = strings.TrimLeft(v, `"`)
		logrus.Debugf("html=%s, k=%s, v=%s", s, k, v)

		if strings.HasPrefix(k, "src") || strings.HasPrefix(k, "href") || strings.HasSuffix(k, "content") {
			v = modifySrcHrefValue(v)
		}

		str := k + `=` + `"` + v
		//logrus.Debugf("modifySrcHref()之后, v:%s", v)
		return str
	}
}

func modifySrcHrefValue(v string) string {
	vu, err := url.Parse(v)
	if err != nil {
		logrus.Errorf("解析src|href|content的值到URL结构失败,%s", err.Error())
		return v
	}

	var me url.URL
	me = *vu
	me.Scheme = selfScheme
	me.Host = selfHost

	var dst = make(url.Values)
	dst.Add("dstZite", v)

	me.RawQuery += "&" + dst.Encode()
	logrus.Debugf("**** %v ***", me)

	return me.String()
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

func replaceURL(wr http.ResponseWriter, rc io.ReadCloser) {
	defer rc.Close()

	buf, err := ioutil.ReadAll(rc)
	if err != nil {
		logrus.Errorf("roundTrip(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
		http.Error(wr, err.Error(), 503)
		return
	}
	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
	wr.Write([]byte(str))
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
//		http.Error(wr, err.Error(), 503)
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
//		http.Error(wr, err.Error(), 503)
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
//		http.Error(wr, err.Error(), 503)
//		return
//	}
//
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
//	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
//	wr.Write([]byte(str))
//}
