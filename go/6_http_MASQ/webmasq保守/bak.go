package webmasq

//import (
////"bytes"
//"crypto/md5"
//"encoding/hex"
//"github.com/sirupsen/logrus"
//"io"
//"io/ioutil"
//"net/http"
//"net/url"
//"regexp"
//"strings"
//)
//
//var (
//	selfHost   string
//	selfScheme = "http"
//	catchReg   = regexp.MustCompile(`(src|href|content) *= *"http[^"]+`)
//)
//
//func Serve(self string) {
//	selfHost = self
//
//	http.HandleFunc("/newzite", newSite)
//	http.HandleFunc("/new_zite_form_parse", newSiteFormParse)
//	http.HandleFunc("/", masquerade)
//
//	//http.HandleFunc("/sess_ctl", sessionControl)
//	http.HandleFunc("/favicon.ico", func(http.ResponseWriter, *http.Request) { return })
//	http.ListenAndServe(self, nil)
//}
//
//func newSite(wr http.ResponseWriter, req *http.Request) {
//	wr.Write([]byte(ziteForm))
//}
//
//func newSiteFormParse(wr http.ResponseWriter, req *http.Request) {
//	logrus.Debugf("newSite(), url=%s", req.URL.String())
//	formH := req.FormValue("formH")
//	if formH == "" {
//		wr.Write([]byte(`<html><script language='javascript' type='text/javascript'> setTimeout("javascript:location.href='/newzite'", 0); </script></html>`))
//		return
//	}
//
//	logrus.Debugf("从form表单发起的请求,%s", req.URL.String())
//	zScheme := req.FormValue("formS")
//
//	cs := strings.TrimRight(zScheme+formH, "/")
//	setCurSiteCookie(wr, cs)
//	u, err := url.Parse(cs)
//	if err != nil {
//		logrus.Errorf("newSite(),url.Parse(%s) failed, %s", cs, err.Error())
//		return
//	}
//	roundTrip(u, wr)
//
//}
//
//func masquerade(wr http.ResponseWriter, req *http.Request) {
//	if req.Method != "GET" {
//		logrus.Errorf("masquerade(), 不支持的方法,%s", req.Method)
//		//http.Error(wr,"Unsupport MEthod",404)
//		http.NotFound(wr, req)
//		return
//	}
//
//	var dstURL url.URL
//	dstURL = *req.URL
//
//	dz := req.FormValue("dstZite")
//	if dz != "" {
//		u, err := url.Parse(dz)
//		if err != nil {
//			logrus.Errorf("masquerade(),解析dz为url失败")
//			return
//		}
//		dstURL = *u
//		logrus.Debugf("使用dstZite,scheme:%s, host:%s", dstURL.Scheme, dstURL.Host)
//	} else {
//		cz, err := req.Cookie("curZite")
//		if err != nil {
//			logrus.Errorf("未带cookie,%s", req.URL.String())
//			http.Error(wr, "No-CoOKIE", 502)
//			return
//		}
//
//		u, err := url.Parse(cz.Value)
//		if err != nil {
//			logrus.Errorf("masquerade(), 甜饼解析成url失败,%s", cz.Value)
//			http.Error(wr, "Parse-COokie", 502)
//			return
//		}
//		dstURL.Scheme = u.Scheme
//		dstURL.Host = u.Host
//	}
//
//	logrus.Debugf("masquerade(), dst url=%s", dstURL.String())
//
//	roundTrip(&dstURL, wr)
//}
//
////安全备份
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
//	logrus.Debugf("buf len ： %d", len(buf))
//	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
//	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
//	wr.Write([]byte(str))
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
//		//logrus.Debugf("modifySrcHref()之后, v:%s", v)
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
//func sessionControl(wr http.ResponseWriter, req *http.Request) {
//	uname := req.FormValue("zuname")
//	upwd := req.FormValue("zupwd")
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
//	wr.Write([]byte("<html>" + ziteForm + "</html>"))
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
//	wr.Write([]byte(acForm))
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
//
////安全备份
////func roundTrip(dstURL *url.URL, wr http.ResponseWriter) {
////	resp, err := http.Get(dstURL.String())
////	if err != nil {
////		logrus.Errorf("roundTrip(),http.Get(%s) failed, %s", dstURL.String(), err.Error())
////		http.Error(wr, err.Error(), resp.StatusCode)
////		return
////	}
////	defer resp.Body.Close()
////
////	buf, err := ioutil.ReadAll(resp.Body)
////	if err != nil {
////		logrus.Errorf("roundTrip(),ioutil.ReadAll(resp.Body) failed, %s", err.Error())
////		http.Error(wr, err.Error(), 503)
////		return
////	}
////
////	str := catchReg.ReplaceAllStringFunc(string(buf), genReplFunc())
////	//setCurSiteCookie(wr, dstURL.Scheme+"://"+dstURL.Host)
////	wr.Write([]byte(str))
////}
//
////func proxy(dstURL *url.URL, wr http.ResponseWriter, r *http.Request) {
////	logrus.Debugf("proxy: %s", dstURL.String())
////
////	transport := http.DefaultTransport
////
////	outreq, err := http.NewRequest("GET", dstURL.String(), nil)
////	if err != nil {
////		logrus.Errorf("roundTrip(),http.NewRequest(),%v", err)
////		return
////	}
////
////	outreq.Header = cloneHeader(r.Header)
////	outreq.Close = false
////
////	removeConnectionHeaders(outreq.Header)
////
////	// Remove hop-by-hop headers to the backend. Especially
////	// important is "Connection" because we want a persistent
////	// connection, regardless of what the client sent to us.
////	for _, h := range hopHeaders {
////		hv := outreq.Header.Get(h)
////		if hv == "" {
////			continue
////		}
////		if h == "Te" && hv == "trailers" {
////			// Issue 21096: tell backend applications that
////			// care about trailer support that we support
////			// trailers. (We do, but we don't go out of
////			// our way to advertise that unless the
////			// incoming client request thought it was
////			// worth mentioning)
////			continue
////		}
////		outreq.Header.Del(h)
////	}
////
////	*outreq.URL = *dstURL
////	res, err := transport.RoundTrip(outreq)
////	if err != nil {
////		logrus.Errorf("roundTrip(),transport.RoundTrip(), %s", err.Error())
////		return
////	}
////
////	removeConnectionHeaders(res.Header)
////
////	for _, h := range hopHeaders {
////		res.Header.Del(h)
////	}
////
////	copyHeader(wr.Header(), res.Header)
////
////	// The "Trailer" header isn't included in the Transport's response,
////	// at least for *http.Transport. Build it up from Trailer.
////	announcedTrailers := len(res.Trailer)
////	if announcedTrailers > 0 {
////		trailerKeys := make([]string, 0, len(res.Trailer))
////		for k := range res.Trailer {
////			trailerKeys = append(trailerKeys, k)
////		}
////		wr.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
////	}
////
////	wr.WriteHeader(res.StatusCode)
////	if len(res.Trailer) > 0 {
////		// Force chunking if we saw a response trailer.
////		// This prevents net/http from calculating the length for short
////		// bodies and adding a Content-Length.
////		if fl, ok := wr.(http.Flusher); ok {
////			fl.Flush()
////		}
////	}
////
////	replaceURL(wr, res.Body)
////
////	if len(res.Trailer) == announcedTrailers {
////		copyHeader(wr.Header(), res.Trailer)
////		return
////	}
////
////	for k, vv := range res.Trailer {
////		k = http.TrailerPrefix + k
////		for _, v := range vv {
////			wr.Header().Add(k, v)
////		}
////	}
////}
//
////安全备份
//func proxy1(dstURL *url.URL, wr http.ResponseWriter, req *http.Request) {
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
