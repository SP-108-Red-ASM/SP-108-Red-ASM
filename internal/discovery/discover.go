package discovery

import (
	"SP_108_Red_ASM/internal/config"
	"SP_108_Red_ASM/internal/model"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

type nmapRun struct {
	Hosts []nmapHost `xml:"host"`
}

type nmapHost struct {
	Status nmapStatus `xml:"status"`
	Addr   []nmapAddr `xml:"address"`
	Ports  nmapPorts  `xml:"ports"`
}

type nmapStatus struct {
	State string `xml:"state,attr"`
}

type nmapAddr struct {
	Addr string `xml:"addr,attr"`
	Type string `xml:"addrtype,attr"`
}

type nmapPorts struct {
	Port []nmapPort `xml:"port"`
}

type nmapPort struct {
	Proto   string     `xml:"protocol,attr"`
	PortID  int        `xml:"portid,attr"`
	State   nmapPState `xml:"state"`
	Service nmapSvc    `xml:"service"`
}

type nmapPState struct {
	State string `xml:"state,attr"`
}

type nmapSvc struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

func Scan(c model.Cfg) ([]model.Host, error) {
	fmt.Printf("[*] starting discovery: %d target(s), %d worker(s)\n\n", len(c.Scope), c.Workers)

	jobs := make(chan string, len(c.Scope))
	res := make(chan model.Host, len(c.Scope))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < c.Workers; i++ {
		wg.Add(1)
		go scanWorker(c, jobs, res, &wg, &mu)
	}
	for _, t := range c.Scope {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	close(res)

	var hosts []model.Host
	for h := range res {
		hosts = append(hosts, h)
	}
	return hosts, nil
}

func scanWorker(c model.Cfg, jobs <-chan string, res chan<- model.Host, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	for t := range jobs {
		ok := config.Allowed(t, c.Scope)
		if !ok {
			mu.Lock()
			fmt.Printf("[!] %s skipped: not in scope\n", t)
			mu.Unlock()
			continue
		}

		mu.Lock()
		fmt.Printf("[*] scanning %s ...\n", t)
		mu.Unlock()

		// -oX - streams XML to stdout instead of a file
		args := append([]string{"-oX", "-"}, c.NmapArgs...)
		args = append(args, t)
		out, err := exec.Command("nmap", args...).Output()
		if err != nil {
			mu.Lock()
			fmt.Printf("[-] %s error: %v\n", t, err)
			mu.Unlock()
			continue
		}

		var nr nmapRun
		err = xml.Unmarshal(out, &nr)
		if err != nil {
			mu.Lock()
			fmt.Printf("[-] %s xml parse error: %v\n", t, err)
			mu.Unlock()
			continue
		}

		if len(nr.Hosts) == 0 {
			mu.Lock()
			fmt.Printf("[-] %s no response\n", t)
			mu.Unlock()
			continue
		}

		for _, nh := range nr.Hosts {
			h := model.Host{State: nh.Status.State}
			for _, a := range nh.Addr {
				if a.Type == "ipv4" || a.Type == "ipv6" {
					h.IP = a.Addr
				}
			}
			for _, p := range nh.Ports.Port {
				v := strings.TrimSpace(p.Service.Product + " " + p.Service.Version)
				h.Ports = append(h.Ports, model.PortSvc{
					Port:    p.PortID,
					Proto:   p.Proto,
					Service: p.Service.Name,
					Version: v,
					State:   p.State.State,
				})
			}
			mu.Lock()
			fmt.Printf("[+] %s up - %d port(s) found\n", h.IP, len(h.Ports))
			mu.Unlock()
			res <- h
		}
	}
}
