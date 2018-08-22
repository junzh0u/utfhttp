package httpx

import (
	"io/ioutil"
	"net/http"
	"os/exec"
	"time"

	"github.com/benbjohnson/phantomjs"
	"golang.org/x/net/html/charset"

	"github.com/junzh0u/ioutilx"
)

// GetContentFunc is a function type that takes a url and returns its body in
// string
type GetContentFunc func(string) (string, error)

// GetContentInUTF8 takes a GetFunc and returns a GetContentFunc which calls the
// GetFunc and returns the content of response in UTF8
func GetContentInUTF8(getfunc GetFunc) GetContentFunc {
	return func(url string) (string, error) {
		resp, err := getfunc(url)
		if err != nil {
			return "", err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		contentType := resp.Header.Get("Content-Type")
		encoding, _, _ := charset.DetermineEncoding(body, contentType)
		utfBody, err := encoding.NewDecoder().Bytes(body)
		if err != nil {
			return "", err
		}
		return string(utfBody), nil
	}
}

// GetContentViaPhantomJS returns a GetContentFunc that uses PhantomJS wrapper
func GetContentViaPhantomJS(cookies []*http.Cookie, waitDuration time.Duration) GetContentFunc {
	return func(url string) (string, error) {
		page, err := phantomjs.DefaultProcess.CreateWebPage()
		if err != nil {
			return "", err
		}
		defer page.Close()

		page.SetCookies(cookies)
		err = page.Open(url)
		if err != nil {
			return "", err
		}

		if waitDuration > 0 {
			page.Reload()
			time.Sleep(waitDuration)
		}

		return page.Content()
	}
}

// GetContentViaStandalonePhantomJS returns a GetContentFunc that uses
// PhantomJS in a standalone process
func GetContentViaStandalonePhantomJS() GetContentFunc {
	return func(url string) (string, error) {
		script, err := ioutilx.TempFileWithContent(
			"/tmp",
			"phatomjs.savepage",
			`var system = require('system');
			var page = require('webpage').create();

			page.onError = function(msg, trace) {
				// do nothing
			};

			page.open(system.args[1], function(status) {
			console.log(page.content);
			phantom.exit();
			});`)
		if err != nil {
			return "", err
		}

		content, err := exec.Command("phantomjs", script, url).Output()
		return string(content), err
	}
}
