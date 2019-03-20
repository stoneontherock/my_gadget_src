package webmasq

//import (
//	//"bytes"
//	"crypto/md5"
//	"encoding/hex"
//	"github.com/sirupsen/logrus"
//	"io"
//	"io/ioutil"
//	"net"
//	"net/http"
//	"net/url"
//	"regexp"
//	"strings"
//	"time"
//)
//
//var (
//	selfHost   string
//	selfScheme = "https"
//	catchReg   = regexp.MustCompile(`(src|href|content) *= *"http[^"]+`) //todo location??
//	repl30X    = regexp.MustCompile(`(https?://)[^/]+(.*)`)
//
//	transport http.RoundTripper = &http.Transport{
//		Proxy: http.ProxyFromEnvironment,
//		DialContext: (&net.Dialer{
//			Timeout:   30 * time.Second,
//			KeepAlive: 30 * time.Second,
//			DualStack: true,
//		}).DialContext,
//		DisableCompression:    true,
//		MaxIdleConns:          100,
//		IdleConnTimeout:       90 * time.Second,
//		TLSHandshakeTimeout:   10 * time.Second,
//		ExpectContinueTimeout: 1 * time.Second,
//	}
//)
//
//func Serve(self, certF, keyF string) error {
//	selfHost = self
//
//	http.HandleFunc("/", masquerade)
//	http.HandleFunc("/zzz", newSite)
//	http.HandleFunc("/site", newSiteHandler)
//	//http.HandleFunc("/login", login)
//
//	http.HandleFunc("/favicon.ico", func(http.ResponseWriter, *http.Request) { return })
//	return http.ListenAndServeTLS(self, certF, keyF, nil)
//}
//
//func newSite(wr http.ResponseWriter, req *http.Request) {
//	wr.Write([]byte(newSiteHtml))
//}
//
//func newSiteHandler(wr http.ResponseWriter, req *http.Request) {
//	logrus.Debugf("newSite(), url=%s", req.URL.String())
//	formH := req.FormValue("siteAddr")
//	if formH == "" {
//		wr.Write([]byte(jump2NewSiteJS))
//		return
//	}
//
//	logrus.Debugf("从form表单发起的请求,%s", req.URL.String())
//	zScheme := req.FormValue("webScheme")
//
//	cs := strings.TrimRight(zScheme+formH, "/")
//	setCurSiteCookie(wr, cs)
//	u, err := url.Parse(cs)
//	if err != nil {
//		logrus.Errorf("newSite(),url.Parse(%s) failed, %s", cs, err.Error())
//		return
//	}
//	proxy(u, wr, req)
//
//}
//
//func masquerade(wr http.ResponseWriter, req *http.Request) {
//	//if req.Method != "GET" {
//	//	logrus.Errorf("masquerade(), 不支持的方法,%s", req.Method)
//	//	//http.Error(wr,"Unsupport MEthod",404)
//	//	http.NotFound(wr, req)
//	//	return
//	//}
//	logrus.Debugf("masquerade(), 原始URL:%s", req.URL.String())
//
//	//req.Cookie("")
//
//	var dstURL *url.URL
//	var err error
//	//dstURL = req.URL
//
//	dz := req.FormValue("dstZite")
//	if dz != "" {
//		dstURL, err = url.Parse(dz)
//		if err != nil {
//			logrus.Errorf("masquerade(),解析dz为url失败")
//			return
//		}
//		logrus.Debugf("使用dstZite,scheme:%s, host:%s", dstURL.Scheme, dstURL.Host)
//	} else {
//		siteAddr := req.FormValue("siteAddr")
//		if siteAddr == "" {
//			wr.Write([]byte(jump2NewSiteJS))
//			return
//		}
//
//		webScheme := req.FormValue("webScheme")
//		dstURL, err = url.Parse(webScheme + siteAddr)
//		if err != nil {
//			logrus.Errorf("form表单输入的url不合法")
//			wr.Write([]byte(jump2NewSiteJS))
//		}
//
//		cs := dstURL.Scheme + "://" + dstURL.Host
//		setCurSiteCookie(wr, cs)
//
//		logrus.Debugf("从form表单发起的请求,url结构体： %p | %[1]v", dstURL)
//
//	}
//
//	logrus.Debugf("masquerade(), 目标url=%s", dstURL.String())
//
//	proxy(dstURL, wr, req)
//}
//
//func proxy(dstURL *url.URL, wr http.ResponseWriter, req *http.Request) {
//	logrus.Debugf("proxy(): %s", dstURL.String())
//
//	outreq, err := http.NewRequest(req.Method, dstURL.String(), nil)
//	if err != nil {
//		logrus.Errorf("proxy(),http.NewRequest(),%v", err)
//		return
//	}
//
//	outreq.Header = cloneHeader(req.Header)
//	outreq.Close = false
//	removeConnectionHeaders(outreq.Header)
//	for _, h := range hopHeaders {
//		hv := outreq.Header.Get(h)
//		if hv == "" {
//			continue
//		}
//		if h == "Te" && hv == "trailers" {
//			continue
//		}
//		outreq.Header.Del(h)
//	}
//
//	*outreq.URL = *dstURL
//	//outreq.Body = req.Body //fixme 是否需要？
//	res, err := transport.RoundTrip(outreq)
//	if err != nil {
//		logrus.Errorf("proxy(),transport.RoundTrip(), %s", err.Error())
//		return
//	}
//
//	removeConnectionHeaders(res.Header)
//	for _, h := range hopHeaders {
//		res.Header.Del(h)
//	}
//
//	//logrus.Debugf("proxy(), res.Header: %v", res.Header)
//	res.Header.Del("Content-Length")
//	copyHeader(wr.Header(), res.Header)
//
//	announcedTrailers := len(res.Trailer)
//	if announcedTrailers > 0 {
//		trailerKeys := make([]string, 0, len(res.Trailer))
//		for k := range res.Trailer {
//			trailerKeys = append(trailerKeys, k)
//		}
//		wr.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
//	}
//
//	modify30XHeader(wr, req, res, dstURL)
//	wr.WriteHeader(res.StatusCode)
//	if len(res.Trailer) > 0 {
//		if fl, ok := wr.(http.Flusher); ok {
//			fl.Flush()
//		}
//	}
//
//	n, err := wr.Write([]byte(replaceURL(wr, res.Body)))
//	if err != nil {
//		logrus.Errorf("写wr失败，%dB, %v", n, err)
//		return
//	}
//
//	if len(res.Trailer) == announcedTrailers {
//		copyHeader(wr.Header(), res.Trailer)
//		return
//	}
//
//	for k, vv := range res.Trailer {
//		k = http.TrailerPrefix + k
//		for _, v := range vv {
//			wr.Header().Add(k, v)
//		}
//	}
//
//}
//
//func modify30XHeader(wr http.ResponseWriter, req *http.Request, res *http.Response, originURL *url.URL) {
//	if res.StatusCode != 301 && res.StatusCode != 302 && res.StatusCode != 303 && res.StatusCode != 307 {
//		return
//	}
//
//	nl := res.Header.Get("Location")
//	u, err := url.Parse(nl)
//	if err != nil {
//		logrus.Errorf("proxy(),url.Parse(%s),%v", nl, err)
//		return
//	}
//
//	var v = make(url.Values, 1)
//	v.Add("destZite", u.Scheme+"://"+nl)
//	u.RawQuery = v.Encode()
//	u.Scheme = selfScheme
//	u.Host = selfHost
//	res.Header.Set("Location", u.String())
//
//	if u.Scheme != originURL.Scheme || u.Host != originURL.Host {
//		logrus.Debugf("scheme/host变更，刷新cookie")
//		setCurSiteCookie(wr, u.Scheme+"://"+u.Host)
//	}
//
//	logrus.Debugf("30X改头, 原:%s,新：%s", originURL, nl)
//}
//
//func replaceURL(wr http.ResponseWriter, rc io.ReadCloser) string {
//	defer rc.Close()
//
//	buf, err := ioutil.ReadAll(rc)
//	if err != nil {
//		logrus.Errorf("replaceURL(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
//		http.Error(wr, err.Error(), 503)
//		return ""
//	}
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc()) //fixme
//
//	return str
//}
//
//func genReplFunc() func(string) string {
//	return func(s string) string {
//		i := strings.Index(s, "=")
//		k := strings.TrimRight(s[:i], " ")
//		v := strings.TrimLeft(s[i+1:], " ")
//		v = strings.TrimLeft(v, `"`)
//		logrus.Debugf("html=%s, k=%s, v=%s", s, k, v)
//
//		if strings.HasPrefix(k, "src") || strings.HasPrefix(k, "href") || strings.HasSuffix(k, "content") {
//			v = modifySrcHrefValue(v)
//		}
//
//		str := k + `=` + `"` + v
//		logrus.Debugf("modifySrcHref()之后, kv:%s", str)
//		return str
//	}
//}
//
//func modifySrcHrefValue(v string) string {
//	vu, err := url.Parse(v)
//	if err != nil {
//		logrus.Errorf("解析src|href|content的值到URL结构失败,%s", err.Error())
//		return v
//	}
//
//	var me url.URL
//	me = *vu
//	me.Scheme = selfScheme
//	me.Host = selfHost
//
//	var dst = make(url.Values)
//	dst.Add("dstZite", v)
//
//	me.RawQuery += "&" + dst.Encode()
//	logrus.Debugf("**** %v ***", me)
//
//	return me.String()
//}
//
//func login(wr http.ResponseWriter, req *http.Request) {
//	uname := req.FormValue("uname")
//	upwd := req.FormValue("upwd")
//	if uname == "" || upwd == "" {
//		sendAuth(wr)
//		return
//	}
//
//	h := Md5sum(uname + "/" + upwd)
//	if !auth(h) {
//		sendAuth(wr)
//		return
//	}
//
//	ck := http.Cookie{
//		Name:     "zite",
//		Value:    h,
//		Path:     "/",
//		HttpOnly: true,
//	}
//
//	http.SetCookie(wr, &ck)
//	wr.Write([]byte(newSiteHtml))
//	return
//}
//
//func setCurSiteCookie(wr http.ResponseWriter, s string) {
//	ck := http.Cookie{
//		Name:     "curZite",
//		Value:    s,
//		Path:     "/",
//		HttpOnly: true,
//	}
//	http.SetCookie(wr, &ck)
//}
//
//func sendAuth(wr http.ResponseWriter) {
//	wr.Write([]byte(loginHtml))
//}
//
//func auth(hashStr string) bool {
//	for i, _ := range userAuth {
//		if hashStr == userAuth[i] {
//
//			return true
//		}
//	}
//	return false
//}
//
//func Md5sum(str string) string {
//	w := md5.New()
//	io.WriteString(w, str+salt)
//	return hex.EncodeToString(w.Sum(nil))
//}

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
