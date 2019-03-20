package cc

import (
	"cc/core/module/anti_cc/common"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"html/template"
	"net/http"
)

type verCodeData struct {
	PngBase64 *template.URL
	ReqURI    string
}

var CaptchaTemp *template.Template

func sendCaptcha(wr http.ResponseWriter, req *http.Request) {
	idKey, cap := base64Captcha.GenerateCaptcha("", DefaultCaptchaConf)
	pngB64 := base64Captcha.CaptchaWriteToBase64Encoding(cap)

	cip := common.SplitHostPort(req.RemoteAddr, 0)
	storeCaptcha(cip, idKey)

	vcd := new(verCodeData)
	vcd.PngBase64 = (*template.URL)(&pngB64)
	vcd.ReqURI = req.RequestURI

	err := CaptchaTemp.Execute(wr, vcd)
	if err != nil {
		logrus.Errorf("CaptchaTemp.Execute() failed,%dB,%v", err)
	}
}

func storeCaptcha(ip, idkey string) {
	httpBlacklist.Lock()
	defer httpBlacklist.Unlock()
	if _, ok := httpBlacklist.data[ip]; ok {
		httpBlacklist.data[ip].idKey = idkey
	}
}

var DefaultCaptchaConf = base64Captcha.ConfigDigit{
	Height:     90,
	Width:      165,
	MaxSkew:    1,
	DotCount:   100,
	CaptchaLen: 5,
}
