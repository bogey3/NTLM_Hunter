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

var ntlmPaths = []string{"/", "/owa/", "/ews/", "/autodiscover/", "/mapi/", "/oab/", "/activesync/", "/rpc/", "/rpc/rpcproxy.dll", "/PowerShell/", "/ecp/", "/CertProv/", "/WebTicket/", "/ucwa/", "/MCX/", "/_vti_bin/", "/_layouts/", "/sites/", "/mysite/", "/search/", "/adfs/ls/", "/adfs/services/trust/", "/adfs/services/", "/Reports/", "/ReportServer/", "/teams/auth/", "/ms-als/", "/ntlm/", "/auth/", "/login/", "/api/", "/sap/bc/", "/adfs/services/trust/2005/windowstransport"}
var portToScheme = map[int]string{25: "smtp", 80: "http", 443: "https", 445: "smb", 8080: "http", 8443: "https", 3389: "rdp"}

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

func doNTLMLookups(input chan *url.URL, wg *sync.WaitGroup) {

	writerChan := make(chan NTLM_Info.TargetStruct, len(input))
	go processWriterChan(writerChan)
	for host := range input {
		target := NTLM_Info.TargetStruct{}
		target.TargetURL = host
		wg.Add(1)
		go doNTLMLookup(target, wg, writerChan)
	}
	wg.Done()
	close(writerChan)
}

func processWriterChan(writerChan chan NTLM_Info.TargetStruct) {
	for target := range writerChan {
		target.Print()
		fmt.Println()
	}
}

func doNTLMLookup(target NTLM_Info.TargetStruct, wg *sync.WaitGroup, writerChan chan NTLM_Info.TargetStruct) {
	defer wg.Done()
	if err := target.GetChallenge(); err == nil {
		writerChan <- target
	}
}

func testPorts(host string, output chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	for port, scheme := range portToScheme {
		wg.Add(1)
		go testPort(host, port, scheme, output, wg)
	}

}

func testPort(host string, port int, scheme string, output chan *url.URL, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second*5)
	if err == nil {
		if err2 := conn.Close(); err2 != nil {
		}
		for _, generatedUrl := range buildUrl(scheme, host, port) {
			output <- generatedUrl
		}
	}
}

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	hosts := strings.Split(string(data), "\n")
	targetChan := make(chan *url.URL, len(hosts)*len(portToScheme))
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
