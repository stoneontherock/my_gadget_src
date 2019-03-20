package parse

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"cc/core/module/anti_cc"
	"cc/core/module/anti_cc/basic_rule"
)

type mainConf struct {
	CookieSalt     string        `yaml:"cookie_salt"`
	FRJsDelay      time.Duration `yaml:"first_js_deley_seconds"`
	L1BListChkIntv time.Duration `yaml:"blacklist1_check_interval"`
	L2BListChkIntv time.Duration `yaml:"blacklist2_check_interval"`
	FRJsHTML       string        `yaml:"first_js_html"`
	CaptchaHTML    string        `yaml:"captcha_html"`
}

var mainconfig mainConf

type siteConf struct {
	Name        string        `yaml:"name"`
	AmongTime   time.Duration `yaml:"among_time,omitempty"`
	Threshold   int           `yaml:"threshold,omitempty"`
	L1BlistTerm time.Duration `yaml:"prison1_term,omitempty"` //一级黑名单刑期，刑满释放
	CaptchaTTL  int           `yaml:"captcha_fail,omitempty"`
	L2BlistTerm time.Duration `yaml:"prison2_term,omitempty"` //二级黑名单刑期，刑满释放
	Code        int           `yaml:"redirect_code,omitempty"`
	URL         string        `yaml:"redirect_URL,omitempty"`
}

type pathRule struct {
	Rule   string     `yaml:"path_rule"`
	CCconf []siteConf `yaml:"cc_conf,flow"`
}

type domainSelector struct {
	Rule     string     `yaml:"domain_rule"`
	CCconf   []siteConf `yaml:"cc_conf,flow"`
	PathList []pathRule `yaml:"path_rule_list"`
}

func ParseConfFile(confd string) (*cc.DomainGroup, error) {
	//解析主配置
	mcbuf, err := ioutil.ReadFile(filepath.Dir(confd) + "/main.ccconf")
	if err != nil {
		return nil, fmt.Errorf("parse.ParseConfFile(),common.ReadFil,mainconf,%s", err.Error())
	}

	err = yaml.Unmarshal(formatColonSpace(mcbuf), &mainconfig)
	if err != nil {
		return nil, fmt.Errorf("parse.ParseConfFile(),Unmarshal mainconf,%s", err.Error())
	}

	err = mainconfig.getMainConf()
	if err != nil {
		return nil, fmt.Errorf("getMainConf(),%s", err.Error())
	}

	// 解析站点配置
	files, err := filepath.Glob(confd + "/*.ccconf")
	if err != nil {
		return nil, fmt.Errorf("parse.ParseConfFile(),filepath.Glob,%s", err.Error())
	}

	var dslist []*cc.DomainSelector
	for _, f := range files {
		buf, err := ioutil.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse.ParseConfFile(),common.ReadFile,siteconf,%s", err.Error())
		}
		//logrus.Debugf("***BUF:%s****\n", string(buf))

		var ds domainSelector
		err = yaml.Unmarshal(formatColonSpace(buf), &ds)
		if err != nil {
			return nil, fmt.Errorf("parse.ParseConfFile(),yaml.Unmarshal,%s", err.Error())
		}

		ccds, err := validateDomainSelector(&ds)
		if err != nil {
			return nil, fmt.Errorf("parse.validate(),%s", err.Error())
		}
		dslist = append(dslist, ccds)
		//logrus.Debugf("ccds:%v", ccds)
	}

	//logrus.Debugf("dslist:%v", dslist)

	return cc.NewDomainGroup(dslist), nil
}

func (mc *mainConf) getMainConf() error {
	switch {
	case mc.CookieSalt == "":
		return errors.New("'first_js_salt' is empty")
	case mc.FRJsDelay < 1:
		return errors.New("'first_js_deley_seconds' must >= 1")
	case mc.FRJsHTML == "", mc.CaptchaHTML == "":
		return errors.New("'first_js_html' or 'captcha_html' empty")
	case mc.L1BListChkIntv < time.Second*15:
		return errors.New("'blacklist1_check_interval' must >= 15s")
	case mc.L2BListChkIntv < time.Second*15:
		return errors.New("'blacklist2_check_interval' must >= 15s")
	}

	jshtml, err := ioutil.ReadFile(mc.FRJsHTML)
	if err != nil {
		return fmt.Errorf("validataMainConf(),readfile,%s", err.Error())
	}
	cc.FRJsHTML = string(jshtml)

	cc.CaptchaTemp, err = template.ParseFiles(mc.CaptchaHTML)
	if err != nil {
		return fmt.Errorf("template.ParseFiles(),%s", err.Error())
	}

	cc.CookieSalt = mc.CookieSalt
	cc.FRJsDelay = mc.FRJsDelay
	cc.L1BlacklistChkIntv = mc.L1BListChkIntv
	cc.L2BlacklistChkIntv = mc.L2BListChkIntv

	return nil
}

func validateDomainSelector(ds *domainSelector) (*cc.DomainSelector, error) {
	var err error

	var ccPathlist []cc.PathRule

	for _, p := range ds.PathList {
		var ccPr cc.PathRule
		ccPr.Rule, err = convert2BasicRule(p.Rule)
		//logrus.Debugf("ccPathRule.Rule:%v,err:%v", ccPr.Rule, err)
		if err != nil {
			return nil, err
		}

		ccPr.CCconf, err = getCCConf(p.CCconf)
		//logrus.Debugf("ccPathRule.CCconf:%v,err:%v", p.CCconf, err)
		if err != nil {
			return nil, err
		}

		ccPathlist = append(ccPathlist, ccPr)

	}

	var ccDS cc.DomainSelector
	ccDS.Rule, err = convert2BasicRule(ds.Rule)
	if err != nil {
		return nil, err
	}
	ccDS.CCconf, err = getCCConf(ds.CCconf)
	if err != nil {
		return nil, err
	}
	ccDS.PathList = ccPathlist

	return &ccDS, nil
}

func convert2BasicRule(str string) (basic_rule.Rule, error) {
	var r basic_rule.Rule
	var err error
	if strings.HasPrefix(str, "wildcard=") {
		r, err = basic_rule.NewRule(str[9:], true)
	} else if strings.HasPrefix(str, "regex=") {
		r, err = basic_rule.NewRule(str[6:], false)
	} else {
		return basic_rule.Rule{}, errors.New("only 'regex' and 'wildcard' been surpported ")
	}

	if err != nil {
		return basic_rule.Rule{}, fmt.Errorf("parse regex/wildcard string failed,string=%s, %s", err.Error())
	}

	return r, nil
}

func getCCConf(confList []siteConf) ([]interface{}, error) {
	var ccconf []interface{}
	for i, conf := range confList {
		//fmt.Printf("conf.Name=%s,conf.Code=%d\n", conf.Name, conf.Code)
		switch conf.Name {
		case "limit_req":
			if conf.Threshold < 1 || conf.CaptchaTTL < 1 || conf.AmongTime < time.Millisecond*500 ||
				conf.L1BlistTerm < mainconfig.L1BListChkIntv || conf.L2BlistTerm < mainconfig.L2BListChkIntv {
				return nil, fmt.Errorf("limit_req must satisfied: among_time>=500ms, captcha_fail: >=1, prison1_term: >=blacklist1_check_interval, prison2_term: >=blacklist1_check_interval, threshold: >=1")
			}
			dbp := cc.NewReqFreqLimitConf(conf.AmongTime, conf.Threshold, conf.L1BlistTerm, conf.CaptchaTTL, conf.L2BlistTerm)
			ccconf = append(ccconf, dbp)
		case "js":
			if conf.Threshold < 1 || conf.CaptchaTTL < 1 || conf.AmongTime < time.Second*20 ||
				conf.L1BlistTerm < mainconfig.L1BListChkIntv || conf.L2BlistTerm < mainconfig.L2BListChkIntv {
				return nil, fmt.Errorf("limit_req must satisfied: among_time>=20s, captcha_fail: >=1, prison1_term: >=blacklist1_check_interval, prison2_term: >=blacklist1_check_interval, threshold: >=1")
			}
			jsp := cc.NewFirstReqJsConf(conf.AmongTime, conf.Threshold, conf.L1BlistTerm, conf.CaptchaTTL, conf.L2BlistTerm)
			ccconf = append(ccconf, jsp)
		case "redirect":
			fmt.Printf("conf.Name=%s,conf.Code=%d\n", conf.Name, conf.Code)
			if conf.Code < 300 || conf.Code > 399 { /*|| !strings.HasPrefix(conf.URL, "http")*/
				return nil, fmt.Errorf("redirect_code should between 300 and 399")
			}
			var redi cc.Redirect_conf
			redi.Code = conf.Code
			redi.URL = conf.URL
			logrus.Debugf("RediURL:_%s_", conf.URL)
			ccconf = append(ccconf, &redi)
		default:
			logrus.Errorf("unknown cc conf type, type=%t, index=%d", conf, i)
			return nil, fmt.Errorf("unknown cc conf type,index=%d", i)
		}
	}

	//logrus.Debugf("len ccconf=%d", len(ccconf))
	return ccconf, nil
}

func formatColonSpace(src []byte) []byte {
	src0 := bytes.Replace(src, []byte(":"), []byte(": "), -1)
	return bytes.Replace(src0, []byte(": //"), []byte("://"), -1)
}
