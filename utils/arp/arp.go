package arp

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync"
	"test/model"
	"time"

	"github.com/mdlayher/arp"
)

func FindDomainIPs(dom model.Domain) []string {
	ips := []string{}
	macs := []string{}
	for _, i2 := range dom.Interfaces {
		macs = append(macs, *i2.Mac.Address)
	}


	database,err := getIPIface()
	if err != nil {
		return []string{}
	}

	for k, v := range database {
		for _, v2 := range macs {
			if v2 == k {
				ips = append(ips, v)
			}
		}
	}

	return ips
}
func getIPIface() (ret map[string]string, err error) { // TODO
	stop := false
	ret = map[string]string{}
	mut := &sync.Mutex{}

	ifis,err := net.Interfaces()
	for _, i2 := range ifis {
		if  i2.Flags & net.FlagLoopback == net.FlagLoopback || 
			i2.Flags & net.FlagRunning  != net.FlagRunning  ||
			i2.Flags & net.FlagUp       != net.FlagUp {
			continue
		}


		client,err := arp.Dial(&i2)
		if err != nil {
			continue
		}



		go func ()  {
			addrs,err := i2.Addrs()
			if err != nil {
				return
			}

			for {
				if stop {
					break
				}

				pkt,_,err := client.Read()
				if err != nil {
					continue
				}
				if pkt.Operation != arp.OperationReply || 
				  !strings.Contains(addrs[0].String(), pkt.TargetIP.String()) {
					continue
				}

				
				mut.Lock()
				ret[pkt.SenderHardwareAddr.String()] = pkt.SenderIP.String()
				mut.Unlock()
			}
		}()

		for i := 0; i < 100; i++ {
			addr,err := netip.ParseAddr(fmt.Sprintf("192.168.1.%d",i))
			if err != nil {
				continue
			}
			err = client.Request(addr)
			if err != nil {
				continue
			}
		}
	}

	time.Sleep(1 * time.Second)
	stop = true
	return 
}