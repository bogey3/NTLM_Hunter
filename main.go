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

func generateUrl(host string, port int) []*url.URL {
	var targetUrl []*url.URL
	var parsedUrl *url.URL
	var err error
	switch port {
	case 25:
		parsedUrl, err = url.Parse(fmt.Sprintf("smtp://%s", host))
		targetUrl = append(targetUrl, parsedUrl)
	case 80:
		for _, path := range ntlmPaths {
			parsedUrl, err = url.Parse(fmt.Sprintf("http://%s%s", host, path))
			targetUrl = append(targetUrl, parsedUrl)
		}
	case 443:
		for _, path := range ntlmPaths {
			parsedUrl, err = url.Parse(fmt.Sprintf("https://%s%s", host, path))
			targetUrl = append(targetUrl, parsedUrl)
		}
	case 3389:
		parsedUrl, err = url.Parse(fmt.Sprintf("rdp://%s", host))
		targetUrl = append(targetUrl, parsedUrl)
	}
	if err == nil {
		return targetUrl
	} else {
		return nil
	}
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
		conn.Close()
		for _, generatedUrl := range generateUrl(host, port) {
			output <- generatedUrl
		}
	}
}

var ports []int

func main() {
	ports = []int{25, 80, 443, 3389}
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
