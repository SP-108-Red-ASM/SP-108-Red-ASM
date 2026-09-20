package main

import (
	"encoding/xml"
	"os/exec"
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
	Name string `xml:"name,attr"`
}

func scan(c Cfg) ([]Host, error) {
	jobs := make(chan string, len(c.Scope))
	res := make(chan Host, len(c.Scope))
	var wg sync.WaitGroup
	for i := 0; i < c.Workers; i++ {
		wg.Add(1)
		go scanWorker(c, jobs, res, &wg)
	}
	for _, t := range c.Scope {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	close(res)

	var hosts []Host
	for h := range res {
		hosts = append(hosts, h)
	}
	return hosts, nil
}

func scanWorker(c Cfg, jobs <-chan string, res chan<- Host, wg *sync.WaitGroup) {
	defer wg.Done()
	for t := range jobs {
		ok := allowed(t, c.Scope)
		if !ok {
			continue
		}
		// -oX - streams XML to stdout instead of a file
		args := append([]string{"-oX", "-"}, c.NmapArgs...)
		args = append(args, t)
		out, err := exec.Command("nmap", args...).Output()
		if err != nil {
			continue
		}
		var nr nmapRun
		err = xml.Unmarshal(out, &nr)
		if err != nil {
			continue
		}
		for _, nh := range nr.Hosts {
			h := Host{State: nh.Status.State}
			for _, a := range nh.Addr {
				if a.Type == "ipv4" || a.Type == "ipv6" {
					h.IP = a.Addr
				}
			}
			for _, p := range nh.Ports.Port {
				h.Ports = append(h.Ports, PortSvc{
					Port:    p.PortID,
					Proto:   p.Proto,
					Service: p.Service.Name,
					State:   p.State.State,
				})
			}
			res <- h
		}
	}
}
