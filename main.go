package main

import (
	"fmt"
	"github.com/bogey3/NTLM_Info"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var ntlmPaths = []string{"/", "/owa/", "/ews/", "/autodiscover/", "/mapi/", "/oab/", "/activesync/", "/rpc/", "/rpc/rpcproxy.dll", "/PowerShell/", "/ecp/", "/CertProv/", "/WebTicket/", "/ucwa/", "/MCX/", "/_vti_bin/", "/_layouts/", "/sites/", "/mysite/", "/search/", "/adfs/ls/", "/adfs/services/trust/", "/adfs/services/", "/Reports/", "/ReportServer/", "/teams/auth/", "/ms-als/", "/ntlm/", "/auth/", "/login/", "/api/", "/sap/bc/"}

func buildUrl(scheme string, host string, port int) []*url.URL {
	urls := []*url.URL{}
	if strings.HasPrefix(scheme, "http") {
		for _, path := range ntlmPaths {
			parsedUrl, err := url.Parse(fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path))
			if err == nil {
				urls = append(urls, parsedUrl)
			}
		}
	} else {
		parsedUrl, err := url.Parse(fmt.Sprintf("%s://%s:%d", scheme, host, port))
		if err == nil {
			urls = append(urls, parsedUrl)
		}

	}
	return urls
}

func generateUrls(host string, port int) []*url.URL {
	var targetUrl []*url.URL
	switch port {
	case 25:
		targetUrl = append(targetUrl, buildUrl("smtp", host, port)...)
	case 80:
		targetUrl = append(targetUrl, buildUrl("http", host, port)...)
	case 443:
		targetUrl = append(targetUrl, buildUrl("https", host, port)...)
	case 8080:
		targetUrl = append(targetUrl, buildUrl("http", host, port)...)
	case 8443:
		targetUrl = append(targetUrl, buildUrl("https", host, port)...)
	case 3389:
		targetUrl = append(targetUrl, buildUrl("rdp", host, port)...)
	}
	return targetUrl
}

func doNTLMLookups(input chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	for host := range input {
		target := NTLM_Info.TargetStruct{}
		target.TargetURL = host
		wg.Add(1)
		go doNTLMLookup(target, wg)
	}
}

func doNTLMLookup(target NTLM_Info.TargetStruct, wg *sync.WaitGroup) {
	defer wg.Done()
	if err := target.GetChallenge(); err == nil {
		target.Print()
		fmt.Println()
	}
}

func testPorts(host string, output chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, port := range ports {
		wg.Add(1)
		go testPort(host, port, output, wg)
	}

}

func testPort(host string, port int, output chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second*5)
	if err == nil {
		if err2 := conn.Close(); err2 != nil {
		}
		for _, generatedUrl := range generateUrls(host, port) {
			output <- generatedUrl
		}
	}
}

var ports []int

func main() {
	ports = []int{25, 80, 443, 8080, 8443, 3389}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	hosts := strings.Split(string(data), "\n")
	targetChan := make(chan *url.URL, len(hosts)*len(ports))
	portScanWG := sync.WaitGroup{}
	ntlmLookupWG := sync.WaitGroup{}
	ntlmLookupWG.Add(1)
	go doNTLMLookups(targetChan, &ntlmLookupWG)
	for _, host := range hosts {
		portScanWG.Add(1)
		go testPorts(host, targetChan, &portScanWG)
	}
	portScanWG.Wait()
	close(targetChan)
	ntlmLookupWG.Wait()
}
